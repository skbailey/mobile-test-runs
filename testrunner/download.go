package testrunner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
)

// downloadArtifact downloads the artifact from the given URL to the specified local file path.
func downloadArtifact(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// listAndDownloadArtifacts lists and downloads artifacts for a given ARN (run, job, suite, or test).
func listAndDownloadArtifacts(arn *string) error {
	input := &devicefarm.ListArtifactsInput{
		Arn:  arn,
		Type: aws.String("SCREENSHOT"), // Possible values: FILE, LOG, SCREENSHOT
	}

	result, err := client.ListArtifacts(input)
	if err != nil {
		return fmt.Errorf("failed to retrieve artifacts for test run: %w", err)
	}

	for _, artifact := range result.Artifacts {
		fileName := fmt.Sprintf("%s.png", *artifact.Name)
		localPath := filepath.Join("assets/screenshots", fileName)
		if err := downloadArtifact(*artifact.Url, localPath); err != nil {
			return fmt.Errorf("failed to download artifact: %w", err)
		}
		fmt.Printf("Downloaded %s to %s\n", *artifact.Name, localPath)
	}

	return nil
}
