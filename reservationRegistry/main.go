package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const respBodyContent = `Username: %s
Password: %s
Reserve date: %v
Reserve court: %s
Reserve time: %s
`

type reservation struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	ReserveDate  time.Time `json:"reserveDate"`
	ReserveCourt string    `json:"reserveCourt"`
	ReserveTime  string    `json:"reserveTime"`
}

type registry struct {
	bucket   string
	dirname  string
	filename string
	content  []byte
}

func (r registry) createRegistryFile() {
	f, err := os.Create(fmt.Sprintf("/tmp/%s", r.filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(r.content)
}

func (r registry) uploadRegistryFile(sess *session.Session) {
	f, err := os.Open(fmt.Sprintf("/tmp/%s", r.filename))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", r.dirname, r.filename)),
		Body:        f,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		fmt.Printf("Unable to upload %v to %q, %v", f, r.bucket, err)
		return
	}
	fmt.Printf("Successfully uploaded %v to %q\n", f, r.bucket)
}

func reservationRegistry(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data reservation
	// var s3Registry registry

	sess, err := session.NewSession()
	if err != nil {
		exitErrorf("Unable to create session, %v", err)
	}

	// Test to list S3 bucket
	svc := s3.New(sess)

	// Get input payload
	if err := json.Unmarshal([]byte(event.Body), &data); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	filename := fmt.Sprintf("%s-%s-%s-%s",
		data.Username, data.ReserveDate.Format("20060102"), data.ReserveCourt, data.ReserveTime)

	s3Registry := registry{
		bucket:   os.Getenv("S3Bucket"),
		dirname:  "registry",
		filename: fmt.Sprintf("%x.json", (sha256.Sum256([]byte(filename)))),
	}
	s3Registry.content, err = json.Marshal(data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	fmt.Printf("Origin filename: %s\n", filename)
	fmt.Printf("Hased filename: %s\n", s3Registry.filename)
	fmt.Printf("File content: %s\n", s3Registry.content)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3Registry.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s3Registry.dirname, s3Registry.filename)),
	}
	_, err = svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				s3Registry.createRegistryFile()
				s3Registry.uploadRegistryFile(sess)
			default:
				fmt.Println("Error code: ", aerr.Code())
				fmt.Println("Default error: ", aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, err
		}
	} else {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "Registry already exist\n",
		}, nil
	}

	respBody := fmt.Sprintf(respBodyContent,
		data.Username, data.Password, data.ReserveDate, data.ReserveCourt, data.ReserveTime)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       respBody,
	}, nil
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	lambda.Start(reservationRegistry)
}
