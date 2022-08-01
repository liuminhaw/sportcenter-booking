package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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
	for _, obj := range objects.Contents {
		log.Printf("Object key: %s\n", *obj.Key)
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
	if err := chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("_DAAN_LOGIN")),
		chromedp.Title(&title),
	); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 3)

	// fmt.Printf("Title: %s\n", title)
}
