package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/service"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	args := os.Args[1:]
	var action string

	if len(args) >= 1 {
		action = args[0]
	}

	switch action {
	case "delete":
		deleteAction(args[1:])
	case "get":
		getAction(args[1:])
	case "put":
		putAction(args[1:])
	default:
		fmt.Println("Usage: s3tool [get|put|delete] arguments...")
		os.Exit(3)
	}
}

func deleteAction(args []string) {
	if len(args) != 5 {
		fmt.Println("Usage: s3tool delete s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path")
		os.Exit(3)
	}

	accessKey, secretKey, blobStoreURL, bucket, path := args[0], args[1], args[2], args[3], args[4]

	client := connect(accessKey, secretKey, blobStoreURL)

	if _, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}); err != nil {
		fmt.Printf("Error deleting s3://%s/%s: %s\n", bucket, path, err)
		os.Exit(2)
	}

	fmt.Printf("Deleted s3://%s/%s.\n", bucket, path)
}

func getAction(args []string) {
	if len(args) != 6 {
		fmt.Println("Usage: s3tool get s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path destinationFilePath")
		os.Exit(3)
	}

	accessKey, secretKey, blobStoreURL, bucket, path, destPath := args[0], args[1], args[2], args[3], args[4], args[5]

	client := connect(accessKey, secretKey, blobStoreURL)

	output, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		fmt.Printf("Error downloading s3://%s/%s: %s\n", bucket, path, err)
		os.Exit(2)
	}

	destFile, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening %s: %s\n", destPath, err)
		os.Exit(2)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, output.Body); err != nil {
		fmt.Printf("Error writing response to %s: %s\n", destPath, err)
		os.Exit(2)
	}

	fmt.Printf("Downloaded s3://%s/%s to %s.\n", bucket, path, destPath)
}

func putAction(args []string) {
	if len(args) != 6 {
		fmt.Println("Usage: s3tool put s3AccessKey s3SecretKey blobStoreURL s3Bucket s3Path fileToUpload")
		os.Exit(3)
	}

	accessKey, secretKey, blobStoreURL, bucket, path, sourcePath := args[0], args[1], args[2], args[3], args[4], args[5]

	client := connect(accessKey, secretKey, blobStoreURL)

	sourceFile, err := os.OpenFile(sourcePath, os.O_RDONLY, 0444)
	if err != nil {
		fmt.Printf("Error opening %s: %s\n", sourcePath, err)
		os.Exit(2)
	}
	defer sourceFile.Close()

	if _, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   sourceFile,
	}); err != nil {
		fmt.Printf("Error uploading %s: %s\n", sourcePath, err)
		os.Exit(2)
	}

	fmt.Printf("Uploaded %s to s3://%s/%s.\n", sourcePath, bucket, path)
}

func connect(accessKey, secretKey, url string) *s3.S3 {
	awsRegion, awsS3ForcePathStyle := "riak-region-1", true
	client := s3.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         &url,
		Region:           &awsRegion,
		S3ForcePathStyle: &awsS3ForcePathStyle,
	})
	client.Handlers.Sign.Clear()
	client.Handlers.Sign.PushBack(service.BuildContentLength)
	client.Handlers.Sign.PushBack(func(request *service.Request) {
		v2Sign(accessKey, secretKey, request.Time, request.HTTPRequest)
	})
	return client
}
