package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

var s3AccessKey, s3SecretKey, proxyUrl, s3Bucket, s3Path, fileToUpload string

var (
	awsRegion = aws.Region{Name: "riak-region-1", S3Endpoint: "http://s3.amazonaws.com"}
)

const latticeDebugStreamId = "lattice-debug"

func init() {
	flag.Parse()
}

func die(exitCode int, formatString string, args ...interface{}) {
	fmt.Printf(formatString+"\n", args...)
	os.Exit(exitCode)
}

func main() {
	args := flag.Args()

	if len(args) != 6 {
		die(3, "Usage: s3uploader s3AccessKey s3SecretKey httpProxy s3Bucket s3Path fileToUpload")
	}

	s3AccessKey, s3SecretKey, proxyUrl, s3Bucket, s3Path, fileToUpload = args[0], args[1], args[2], args[3], args[4], args[5]

	s3Auth := aws.Auth{
		AccessKey: s3AccessKey,
		SecretKey: s3SecretKey,
	}

	proxyFunc := func(req *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}

	s3S3 := s3.New(s3Auth, awsRegion, &http.Client{
		Transport: &http.Transport{
			Proxy: proxyFunc,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	})

	bucket := s3S3.Bucket(s3Bucket)

	fileInfo, err := os.Stat(fileToUpload)
	if err != nil {
		die(2, "Error stat'ing %s: %s", fileToUpload, err)
	}

	reader, err := os.OpenFile(fileToUpload, os.O_RDONLY, 0)
	if err != nil {
		die(2, "Error opening %s: %s", fileToUpload, err)
	}
	defer reader.Close()

	err = bucket.PutReader(s3Path, reader, fileInfo.Size(), "application/octet-stream", s3.Private, s3.Options{})
	if err != nil {
		die(2, "Error uploading s3://%s/%s: %s", s3Bucket, s3Path, err)
	}

	fmt.Printf("Uploaded %s to s3://%s/%s.\n", fileToUpload, s3Bucket, s3Path)
}
