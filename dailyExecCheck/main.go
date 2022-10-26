package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/liuminhaw/sportcenter-booking/registry"
	"github.com/liuminhaw/sportcenter-booking/secrets"
	"github.com/liuminhaw/sportcenter-booking/storage"
)

func dailyCheck() {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	var keys secrets.Secret
	encKey, err := secrets.GetSecret(sess, os.Getenv("SecretKeyName"))
	if err != nil {
		log.Fatalf("failed to fetch encryption key: %s", err)
	}
	if err := json.Unmarshal([]byte(encKey), &keys); err != nil {
		log.Fatalf("secret key unmarshal failed: %s", err)
	}

	svc := s3.New(sess)
	bucket := os.Getenv("S3Bucket")
	listRegistryInput := s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String("registry"),
	}

	var reservation registry.Reservation
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Fatal("timezone Asia/Taipei not found")
	}
	queuedTime := time.Now().In(loc).AddDate(0, 0, 14)
	if err := svc.ListObjectsV2Pages(&listRegistryInput, func(objs *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range objs.Contents {
			fmt.Printf("Object key: %s\n", *obj.Key)
			content, err := storage.DownloadEncObj(sess, bucket, *obj.Key, keys.S3Enc)
			if err != nil {
				log.Fatalf("failed to list object: %s", err)
			}
			if err := json.Unmarshal(content, &reservation); err != nil {
				log.Fatalf("json unmarshal failed: %s", err)
			}
			if queuedTime.Format("2006-01-02") == reservation.ReserveDate.Format("2006-01-02") {
				fmt.Printf("Move object: %s to queued list\n", *obj.Key)
				_, err := svc.CopyObject(&s3.CopyObjectInput{
					CopySource: aws.String(fmt.Sprintf("%s/%s", bucket, *obj.Key)),
					Bucket:     aws.String(bucket),
					Key:        aws.String(fmt.Sprintf("%s/%s", "queued", strings.TrimPrefix(*obj.Key, "registry"))),
				})
				if err != nil {
					log.Printf("failed to copy queued object: %s", *obj.Key)
					return *objs.KeyCount == *objs.MaxKeys
				}

				_, err = svc.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(bucket),
					Key:    obj.Key,
				})
				if err != nil {
					log.Printf("failed to delete registry object: %s", *obj.Key)
				}
			}
		}

		return *objs.KeyCount == *objs.MaxKeys
	}); err != nil {
		log.Fatalf("failed to list registry objects: %s", err)
	}
}

func main() {
	lambda.Start(dailyCheck)
}
