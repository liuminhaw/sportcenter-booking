package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chromedp/cdproto/network"
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

	cookieObjs, err := getS3Objects(sess, os.Getenv("_S3_BUCKET"), "cookies/")
	if err != nil {
		panic(err)
	}
	registryObjs, err := getS3Objects(sess, os.Getenv("_S3_BUCKET"), "registry/")
	if err != nil {
		panic(err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", *headless),
	)

	var cookies = map[string]bool{}
	for _, obj := range cookieObjs.Contents {
		fmt.Printf("cookie key: %s\n", filepath.Base(*obj.Key))
		cookies[filepath.Base(*obj.Key)] = true
	}
	log.Printf("Initial cookies: %+v", cookies)

	var reserveInfo registry.Reservation
	for _, obj := range registryObjs.Contents {
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

		if cookies[reserveInfo.Username] {
			log.Printf("Cookie for %s exist", reserveInfo.Username)
		} else {
			genCookie(opts, sess, reserveInfo)
			cookies[reserveInfo.Username] = true
		}
	}

	time.Sleep(time.Second * 3)

	// fmt.Printf("Title: %s\n", title)
}

// getS3Objects returns a slice of *s3.ListObjectsV2Output from s3://bucket/prefix/,
// and only returns objects without prefix/
func getS3Objects(sess *session.Session, bucket string, prefix string) (*s3.ListObjectsV2Output, error) {
	svc := s3.New(sess)

	objects, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket),
		Prefix:     aws.String(prefix),
		StartAfter: aws.String(prefix),
	})
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
		return nil, err
	}

	return objects, nil
}

func genCookie(opts []func(*chromedp.ExecAllocator), sess *session.Session, reserveInfo registry.Reservation) {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				if err := chromedp.Run(ctx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					log.Fatal(err)
				}
			}()
		}
	})

	var imgBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(os.Getenv("_DAAN_LOGIN")),
		chromedp.WaitVisible("#ContentPlaceHolder1_loginid", chromedp.ByID),
		chromedp.WaitVisible("#loginpw", chromedp.ByID),
		chromedp.WaitVisible("#ContentPlaceHolder1_Captcha_text", chromedp.ByID),
		chromedp.WaitVisible("#ContentPlaceHolder1_CaptchaImage", chromedp.ByID),
		chromedp.WaitVisible("#login_but", chromedp.ByID),
		chromedp.SendKeys("#ContentPlaceHolder1_loginid", reserveInfo.Username, chromedp.ByID),
		chromedp.SendKeys("#loginpw", reserveInfo.Password, chromedp.ByID),
		chromedp.Screenshot("#ContentPlaceHolder1_CaptchaImage", &imgBuf, chromedp.NodeVisible),
	); err != nil {
		log.Fatal(err)
	}

	reader := bytes.NewReader(imgBuf)
	uploader := s3manager.NewUploader(sess)
	uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("_S3_BUCKET")),
		Key:    aws.String(fmt.Sprintf("captcha/%s.png", reserveInfo.Username)),
		Body:   reader,
	})

	rekognitionSvc := rekognition.New(sess)
	resp, err := rekognitionSvc.DetectText(&rekognition.DetectTextInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(os.Getenv("_S3_BUCKET")),
				Name:   aws.String(fmt.Sprintf("captcha/%s.png", reserveInfo.Username)),
			},
		},
	})
	var captchText string
	if err == nil {
		for _, info := range resp.TextDetections {
			if *info.Type == "LINE" {
				fmt.Printf("Confidence: %f\n", *info.Confidence)
				fmt.Printf("Text: %s\n", *info.DetectedText)
				captchText = *info.DetectedText
				break
			}
		}
	}

	if err := chromedp.Run(ctx,
		chromedp.SendKeys("ContentPlaceHolder1_Captcha_text", captchText, chromedp.ByID),
		chromedp.Click("#login_but", chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}
			for i, c := range cookies {
				log.Printf("chrome cookie %d: %+v - %+v", i, c.Name, c.Value)
				if c.Name == "ASP.NET_SessionId" {
					cookie, err := c.MarshalJSON()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("Marshal cookie: %s\n", cookie)
					reader := bytes.NewReader(cookie)
					uploader := s3manager.NewUploader(sess)
					uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String(os.Getenv("_S3_BUCKET")),
						Key:    aws.String(fmt.Sprintf("cookies/%s", reserveInfo.Username)),
						Body:   reader,
					})
				}
			}
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
}
