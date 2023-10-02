package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"path/filepath"
	"strings"
)

type fileTypes struct {
	filetype        string
	contentType     string
	contentEncoding string
}

var (
	allFileTypes = []fileTypes{
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
			"wasm.gz",
			"application/wasm",
			"gzip",
		},
		{
			"gz",
			"gzip",
			"",
		},
		{
			"js",
			"application/javascript",
			"",
		},
	}
)

func getMetadata(filename string) (map[string]*string, error) {
	for _, filetype := range allFileTypes {
		if strings.HasSuffix(filename, filetype.filetype) {
			metadata := make(map[string]*string)
			metadata["Content-Type"] = &filetype.contentType
			if filetype.contentEncoding != "" {
				metadata["Content-Encoding"] = &filetype.contentEncoding
			}
			return metadata, nil
		}

	}

	return nil, fmt.Errorf("could not get proper metadata for %s", filename)
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
		metadata, err := getMetadata(path)
		if err != nil {
			return err
		}

		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:   aws.String(myBucket),
			Key:      &s3File,
			Body:     f,
			Metadata: metadata,
		})
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
