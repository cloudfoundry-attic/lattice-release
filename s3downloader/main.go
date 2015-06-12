package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

var s3AccessKey, s3SecretKey, proxyUrl, s3Bucket, s3Path, destinationFilePath string

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
		die(3, "Usage: s3downloader s3AccessKey s3SecretKey httpProxy s3Bucket s3Path destinationFilePath")
	}

	s3AccessKey, s3SecretKey, proxyUrl, s3Bucket, s3Path, destinationFilePath = args[0], args[1], args[2], args[3], args[4], args[5]

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

	reader, err := bucket.GetReader(s3Path)
	if err != nil {
		die(2, "Error downloading s3://%s/%s: %s", s3Bucket, s3Path, err)
	}
	defer reader.Close()

	writer, err := os.OpenFile(destinationFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		die(2, "Error opening %s: %s", destinationFilePath, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		die(2, "Error writing %s: %s", destinationFilePath, err)
	}

	fmt.Printf("Downloaded s3://%s/%s to %s.\n", s3Bucket, s3Path, destinationFilePath)
}
