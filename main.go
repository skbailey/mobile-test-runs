package main

import (
	"fmt"
	"log"

	"screenshotter/config"
	"screenshotter/testrunner"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var appConfig config.Config
	err := envconfig.Process("df", &appConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	testrunner.Initialize()

	testRunArn, err := testrunner.ScheduleRun(appConfig)
	if testRunArn == nil || err != nil {
		fmt.Printf("failed to schedule run: %s", err)
		return
	}

	err = testrunner.PollTestRun(appConfig, testRunArn, func() error {
		fmt.Println("downloading artifacts for test run", *testRunArn)
		return testrunner.DownloadArtifacts(testRunArn)
	})
	if err != nil {
		fmt.Printf("failed to retrieve test artifacts: %s\n", err)
		return
	}
}
