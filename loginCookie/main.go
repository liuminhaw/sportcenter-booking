package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/liuminhaw/sportcenter-booking/registry"
	"github.com/liuminhaw/sportcenter-booking/secrets"
)

// create chrome instance in windowed mode, run, and close after 5 seconds
func main() {
	headless := flag.Bool("headless", true, "Set chrome browser headless mode")
	awsProfile := flag.String("profile", "default",
		"Set aws credential to use in session, no credential will be set if using default value")
	awsRegion := flag.String("region", "ap-northeast-1", "Set aws region to use in session")
	flag.Parse()

	var sess *session.Session
	var err error
	if *awsProfile == "default" {
		log.Println("Default aws profile")
		sess, err = session.NewSession()
	} else {
		log.Printf("Profile %s\n", *awsProfile)
		sess, err = session.NewSessionWithOptions(session.Options{
			Config:  aws.Config{Region: aws.String(*awsRegion)},
			Profile: *awsProfile,
		})
	}
	if err != nil {
		panic(err)
	}

	// Fetch secret
	var keys secrets.Secret
	encKey, err := secrets.GetSecret(sess, os.Getenv("_ENC_KEY"))
	if err != nil {
		log.Fatalf("error fetching encryption key: %v\n", err.Error())
	}
	if err := json.Unmarshal([]byte(encKey), &keys); err != nil {
		log.Fatalf("error when unmarhsal from secret manager: %v\n", err.Error())
	}

	svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(os.Getenv("_S3_BUCKET")),
		MaxKeys: aws.Int64(10),
		Prefix:  aws.String("registry"),
	}
	objects, err := svc.ListObjectsV2(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	var reserveInfo registry.Reservation
	for _, obj := range objects.Contents {
		s3Registry := registry.Registry{
			Bucket:   os.Getenv("_S3_BUCKET"),
			Dirname:  filepath.Dir(*obj.Key),
			Filename: filepath.Base(*obj.Key),
		}

		if err := s3Registry.FetchRegistryFile(sess, keys.S3Enc); err != nil {
			log.Fatalf("fetch registry file content error: %v\n", err.Error())
		}
		if err := json.Unmarshal(s3Registry.Content, &reserveInfo); err != nil {
			log.Fatalf("error when unmarhsal from registry: %v\n", err.Error())
		}
		fmt.Printf("Reserve information: %+v\n", reserveInfo)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", *headless),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			// fmt.Println("closing alert:", ev.Message)
			go func() {
				if err := chromedp.Run(ctx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					log.Fatal(err)
				}
			}()
		}
	})

	var title string
	var imgBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("_DAAN_LOGIN")),
		chromedp.Screenshot("#ContentPlaceHolder1_CaptchaImage", &imgBuf, chromedp.NodeVisible),
		chromedp.Title(&title),
	); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("/tmp/captcha.png", imgBuf, 0o644); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 3)

	// fmt.Printf("Title: %s\n", title)
}
