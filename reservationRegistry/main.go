package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/liuminhaw/sportcenter-booking/registry"
)

const respBodyContent = `Username: %s
Reserve date: %v
Reserve court: %s
Reserve time: %s
`

func reservationRegistry(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var data registry.Reservation

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

	s3Registry := registry.Registry{
		Bucket:   os.Getenv("S3Bucket"),
		Dirname:  "registry",
		Filename: fmt.Sprintf("%x", (sha256.Sum256([]byte(filename)))),
	}
	s3Registry.Content, err = json.Marshal(data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, err
	}

	fmt.Printf("Origin filename: %s\n", filename)
	fmt.Printf("Hashed filename: %s\n", s3Registry.Filename)
	fmt.Printf("File content: %s\n", s3Registry.Content)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s3Registry.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", s3Registry.Dirname, s3Registry.Filename)),
	}
	_, err = svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				s3Registry.CreateRegistryFile(os.Getenv("secretKey"))
				s3Registry.UploadRegistryFile(sess)
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
		data.Username, data.ReserveDate, data.ReserveCourt, data.ReserveTime)

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
