package testrunner

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/thoas/go-funk"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
)

var uploads []*devicefarm.Upload

func getARNforUpload(arn, uploadType string) (*string, error) {
	uploads, err := listUploads(arn)
	if err != nil {
		return nil, err
	}

	filteredUploads := funk.Filter(uploads, func(u *devicefarm.Upload) bool {
		if u.Type == nil {
			return false
		}

		return *u.Type == uploadType
	}).([]*devicefarm.Upload)

	if len(uploads) == 0 {
		return nil, nil
	}

	sort.Slice(filteredUploads, func(i, j int) bool {
		previous := *filteredUploads[i].Created
		next := *filteredUploads[j].Created

		return next.Before(previous)
	})

	for _, asset := range filteredUploads {
		fmt.Println("-------------------")
		fmt.Println(*asset.Arn)
		fmt.Println(*asset.Name)
		fmt.Println(*asset.Category)
		fmt.Println(*asset.Type)
		fmt.Println(*asset.Created)
	}

	lastUpload := filteredUploads[0]
	if lastUpload == nil || lastUpload.Category == nil || *lastUpload.Category == "CURATED" {
		// This is not an asset we uploaded, it exists in AWS
		return nil, nil
	}

	return lastUpload.Arn, nil
}

// listUploads lists uploads for a given ARN (run, job, suite, or test).
func listUploads(arn string) ([]*devicefarm.Upload, error) {
	// Example for listing artifacts of a specific type, e.g., "FILE"
	input := &devicefarm.ListUploadsInput{
		Arn: aws.String(arn),
		// Type: aws.String(uploadType),
	}

	err := input.Validate()
	if err != nil {
		return []*devicefarm.Upload{}, err
	}

	// TODO: API returns paginated results, will need to make repeated calls
	result, err := client.ListUploads(input)
	if err != nil {
		return []*devicefarm.Upload{}, err
	}
	uploads = result.Uploads

	fmt.Println("successfully fetched uploads")
	return uploads, nil
}

// uploadDfFile creates an upload in AWS Device Farm and uploads a file.
func uploadDfFile(projectArn, unique, filePath, fileType string) (string, error) {
	fileName := filepath.Base(filePath)
	contentType := "application/octet-stream"

	// Create an upload in Device Farm
	createUploadOutput, err := client.CreateUpload(&devicefarm.CreateUploadInput{
		ProjectArn:  aws.String(projectArn),
		Name:        aws.String(unique + "_" + fileName),
		Type:        aws.String(fileType),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create upload in Device Farm: %w", err)
	}

	uploadURL := createUploadOutput.Upload.Url
	uploadARN := createUploadOutput.Upload.Arn

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Read the file content into a buffer
	buffer := make([]byte, fileSize)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Upload the file to the provided upload URL using HTTP PUT
	req, err := http.NewRequest("PUT", *uploadURL, bytes.NewReader(buffer))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload file, status code: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully uploaded %s to Device Farm\n", filePath)
	return *uploadARN, nil
}

func listDevicePools(projectARN string) (*string, error) {
	input := &devicefarm.ListDevicePoolsInput{
		Arn: aws.String(projectARN),
	}

	err := input.Validate()
	if err != nil {
		return nil, err
	}

	result, err := client.ListDevicePools(input)
	if err != nil {
		return nil, err
	}

	for _, pool := range result.DevicePools {
		fmt.Println("-------------------")
		fmt.Println(*pool.Arn)
		fmt.Println(*pool.Name)
		fmt.Println(*pool.Type)

		if *pool.Name == "Default" {
			return pool.Arn, nil
		}
	}

	return nil, nil
}
