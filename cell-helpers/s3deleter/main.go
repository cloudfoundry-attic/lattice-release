package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	args := os.Args[1:]

	if len(args) != 5 {
		fmt.Println("Usage: s3deleter s3AccessKey s3SecretKey httpProxy s3Bucket s3Path")
		os.Exit(3)
	}

	accessKey, secretKey, proxyURL, bucket, path := args[0], args[1], args[2], args[3], args[4]

	client := s3.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         proxyURL,
		Region:           "riak-region-1",
		S3ForcePathStyle: true,
	})
	client.Handlers.Sign.Clear()
	client.Handlers.Sign.PushBack(aws.BuildContentLength)
	client.Handlers.Sign.PushBack(func(request *aws.Request) {
		v2Sign(accessKey, secretKey, request.Time, request.HTTPRequest)
	})

	if _, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(path),
	}); err != nil {
		fmt.Printf("Error deleting s3://%s/%s: %s\n", bucket, path, err)
		os.Exit(2)
	}

	fmt.Printf("Deleted s3://%s/%s.\n", bucket, path)
}
