package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	case "put":
		putAction(args[1:])
	default:
		fmt.Println("Usage: davtool [put|delete] arguments...")
		os.Exit(3)
	}
}

func sanitizeURL(u string) string {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return u
	}

	parsedURL.User = nil

	return parsedURL.String()
}

func deleteAction(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: davtool delete url")
		os.Exit(3)
	}

	davURL := args[0]

	req, err := http.NewRequest("DELETE", davURL, nil)
	if err != nil {
		fmt.Printf("Error deleting %s: %s\n", sanitizeURL(davURL), err)
		os.Exit(2)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error deleting %s: %s\n", sanitizeURL(davURL), err)
		os.Exit(2)
	}

	if resp.StatusCode != http.StatusNoContent {
		fmt.Printf("Error deleting %s: %s\n", sanitizeURL(davURL), resp.Status)
		os.Exit(2)
	}

	fmt.Printf("Deleted %s.\n", sanitizeURL(davURL))
}

func putAction(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: davtool put url fileToUpload")
		os.Exit(3)
	}

	davURL, sourcePath := args[0], args[1]

	sourceFile, err := os.OpenFile(sourcePath, os.O_RDONLY, 0444)
	if err != nil {
		fmt.Printf("Error opening %s: %s\n", sourcePath, err)
		os.Exit(2)
	}
	defer sourceFile.Close()

	req, err := http.NewRequest("PUT", davURL, sourceFile)
	if err != nil {
		fmt.Printf("Error uploading %s: %s\n", sourcePath, err)
		os.Exit(2)
	}

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		fmt.Printf("Error uploading %s: %s\n", sourcePath, err)
		os.Exit(2)
	}
	req.ContentLength = sourceFileInfo.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error uploading %s: %s\n", sourcePath, err)
		os.Exit(2)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Printf("Error uploading %s: %s\n", sourcePath, resp.Status)
		os.Exit(2)
	}

	fmt.Printf("Uploaded %s to %s.\n", sourcePath, sanitizeURL(davURL))
}
