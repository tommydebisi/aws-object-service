package handle

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

)

var (
	bucketName = os.Getenv("S3Bucket")
)

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

// struct to store the object key gotten from request body
type ObjKey struct {
	Key string `json:"objectKey"`
}

// struct to store the file extension and the local file path from the request body
type FileData struct {
	B64Str string `json:"b64String"`
	ObjName string `json:"objectName"`
}

// endpoint to list the objects in a particular bucket
func ListS3Objects(ctx context.Context, req Request, s3Client *s3.Client) (Response, error) {
	// check if the request is canceled
	if ctx.Err() != nil {
		return ApiResponse(http.StatusRequestTimeout, "Request Canceled")
	}

	// prepare the input param for listObjects func
	input := &s3.ListObjectsV2Input {
		Bucket: aws.String(bucketName),
	}

	// Get the list of objects in the bucket
	result, err := s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		fmt.Printf("Error listing objects: %v\n", err)
		return ApiResponse(http.StatusInternalServerError, "Failed to list objects in S3")	
	}

	// get and store the object key in a slice
	var objectKeys []string

	for _, obj := range result.Contents {
		// derefence the pointer to store the string
		objectKeys = append(objectKeys, *obj.Key)
	}

	// prepare the response body
	respBody := map[string][]string{"objectKeys": objectKeys}

	return ApiResponse(http.StatusOK, respBody)
}

// endpoint to delete an object from a bucket
func DeleteFromS3Bucket(ctx context.Context, req Request, s3Client *s3.Client) (Response, error) {
	// Check if the request is canceled
	if ctx.Err() != nil {
		return ApiResponse(http.StatusRequestTimeout, "Request Canceled")
	}

	// check if content type is json
	if req.Headers["Content-Type"] != "application/json" {
		return ApiResponse(http.StatusBadRequest, "Invalid Content Type")
	}

	var objKey ObjKey

	// get the json string from body and deserialize it to objKey
	err := json.Unmarshal([]byte(req.Body), &objKey)
	if err != nil {
		return ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	// prepare the input param for delete func
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key: &objKey.Key,
	}

	// delete the key specified using the s3 Client
	_, buckErr := s3Client.DeleteObject(ctx, input)
	if buckErr != nil {
		return ApiResponse(http.StatusInternalServerError, "Failed to Delete Object in s3")
	}

	// prepare the response body
	respBody := map[string]string{"message": fmt.Sprintf("Object [%v] was deleted", objKey.Key)}
	return ApiResponse(http.StatusOK, respBody)
}

// endpoint to upload a file to a bucket
func UploadToS3Bucket(ctx context.Context, req Request, s3Client *s3.Client) (Response, error) {
	if ctx.Err() != nil {
		return ApiResponse(http.StatusRequestTimeout, "Request Canceled")
	}
	
	var fileData FileData

	// check if content type is json
	if req.Headers["Content-Type"] != "application/json" {
		return ApiResponse(http.StatusBadRequest, "Invalid Content Type")
	}

	// get the json string from body and deserialize it to fileData
	err := json.Unmarshal([]byte(req.Body), &fileData)
	if err != nil {
		return ApiResponse(http.StatusBadRequest, "Invalid Request Body")
	}

	// decode the base64 string to bytes
	content, decErr :=	base64.StdEncoding.DecodeString(fileData.B64Str)
	if decErr != nil {
		return ApiResponse(http.StatusBadRequest, "Invalid Base64 String")
	}

	// prepare the input param for put
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(fileData.ObjName),
		Body: bytes.NewReader(content),
	}

	// put the object specified in the input param
	_, putErr := s3Client.PutObject(ctx, input)
	if putErr != nil {
		return ApiResponse(http.StatusInternalServerError, "Failed to Upload Object to s3")
	}

	// prepare the response body
	respBody := map[string]string{"message": fmt.Sprintf("Object [%v] was uploaded", fileData.ObjName)}
	return ApiResponse(http.StatusOK, respBody)
}


// endpoint to handle unhandled methods
func UnhandledMethod() (Response, error) {
	return ApiResponse(http.StatusMethodNotAllowed, "Method Not Allowed")
}