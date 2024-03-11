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

	// appConfig := config.Config{
	// 	ProjectARN:  "arn:aws:devicefarm:us-west-2:404667004693:project:78cdd145-8949-41a1-83d0-3043c3692ec2",
	// 	NamePrefix:  "ScreenshotTestRun",
	// 	AppFilePath: "App.ipa",
	// 	TestPackage: "AppUITests-Runner.ipa",
	// }

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
