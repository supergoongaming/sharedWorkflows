package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3FileType struct {
	filetypeSuffix  string
	contentType     string
	contentEncoding string
}

var (
	allFileTypes = []s3FileType{
		{
			"css",
			"text/css",
			"",
		},
		{
			"html",
			"text/html",
			"",
		},
		{
			"js.gz",
			"application/javascript",
			"gzip",
		},
		{
			"wasm.gz",
			"application/wasm",
			"gzip",
		},
		{
			".wasm",
			"application/wasm",
			"",
		},
		{
			"gz",
			"application/x-gzip",
			"gzip",
		},
		{
			"js",
			"application/javascript",
			"",
		},
		{
			"json",
			"application/json",
			"",
		},
		{
			"png",
			"image/png",
			"",
		},
	}
	defaultFiletype = s3FileType{
		"",
		"binary/octet-stream",
		"",
	}
)

func getMetadata(filename string) s3FileType {
	for _, filetype := range allFileTypes {
		if strings.HasSuffix(filename, filetype.filetypeSuffix) {
			return filetype
		}
	}
	fmt.Fprintf(os.Stderr, "Could not find anything in your thing for %s, going to return a default metadata item", filename)
	return defaultFiletype
}

func main() {

	folderPath := os.Getenv("FOLDER_PATH")
	myBucket := os.Getenv("BUCKET_NAME")

	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		s3File, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		filetype := getMetadata(path)
		if err != nil {
			return err
		}

		input := s3manager.UploadInput{
			Bucket:      aws.String(myBucket),
			Key:         &s3File,
			Body:        f,
			ContentType: &filetype.contentType,
		}

		if filetype.contentEncoding != "" {
			input.ContentEncoding = &filetype.contentEncoding
		}

		result, err := uploader.Upload(&input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to upload file, %v", err)
			os.Exit(1)
		}

		fmt.Printf("file uploaded to, %s\n", result.Location)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to walk the directory %s, %v", folderPath, err)
		os.Exit(1)
	}

}
