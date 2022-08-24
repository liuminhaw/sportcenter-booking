package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/liuminhaw/sportcenter-booking/registry"
	"github.com/liuminhaw/sportcenter-booking/secrets"
)

func dailyCheck() {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	fmt.Printf("S3 Bucket: %s\n", os.Getenv("S3Bucket"))
	s3Reg := registry.S3Register{
		Bucket: os.Getenv("S3Bucket"),
		Prefix: "registry",
	}

	// Fetch secret manager key
	var keys secrets.Secret
	fmt.Printf("Secret key name: %s\n", os.Getenv("SecretKeyName"))
	encKey, err := secrets.GetSecret(sess, os.Getenv("SecretKeyName"))
	if err != nil {
		log.Fatalf("error fetching encryption key: %v\n", err.Error())
	}
	if err := json.Unmarshal([]byte(encKey), &keys); err != nil {
		log.Fatalf("error when unmarhsal from secret manager: %v\n", err.Error())
	}

	err = s3Reg.ListObjects(sess)
	if err != nil {
		log.Fatalf("unable to list objects from s3://%s/%s/", s3Reg.Bucket, s3Reg.Prefix)
	}
	for _, obj := range s3Reg.Objects {
		fmt.Printf("Object name: %s\n", obj.Name)
	}
}

func main() {
	lambda.Start(dailyCheck)
}
