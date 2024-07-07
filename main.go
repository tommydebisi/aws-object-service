package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tommydebisi/aws-object-service/handle"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client *s3.Client
)


// main function to start the lambda function
func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, req handle.Request) (handle.Response, error) {
	// load the aws shared config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return handle.ApiResponse(http.StatusInternalServerError, "Failed to Load SDK Configuration")
	}

	// initialize the s3 client
	s3Client = s3.NewFromConfig(cfg)

	fmt.Printf("request: %v\n", req.HTTPMethod)
	// check and handle methods appropriately
	switch req.HTTPMethod {
	case http.MethodGet:
		fmt.Println("here")
		return handle.ListS3Objects(ctx, req, s3Client)
	case http.MethodPost:
		return handle.UploadToS3Bucket(ctx, req, s3Client)
	case http.MethodDelete:
		return handle.DeleteFromS3Bucket(ctx, req, s3Client)
	default:
		return handle.UnhandledMethod()
	}
}
