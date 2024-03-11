package testrunner

import (
	"fmt"
	"time"

	"screenshotter/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/devicefarm"
)

var client *devicefarm.DeviceFarm

func Initialize() {
	config := aws.NewConfig().WithRegion("us-west-2")
	sess := session.Must(session.NewSession(config))
	client = devicefarm.New(sess)
}

func ScheduleRun(appConfig config.Config) (*string, error) {
	unique := appConfig.NamePrefix + "-" + time.Now().Format("2006-01-02") + "-" + generateRandomString(8)
	fmt.Printf("The unique identifier for this run is going to be %s -- all uploads will be prefixed with this.\n", unique)

	appPackageARN, err := getARNforUpload(appConfig.ProjectARN, "IOS_APP")
	if appPackageARN == nil || err != nil {
		// TODO: Upload file if it does not exist in Device Farm
		return nil, fmt.Errorf("failed to retrieve app package arn: %w", err)
	}
	fmt.Println("appPackageARN", *appPackageARN)

	testPackageARN, err := getARNforUpload(appConfig.ProjectARN, "XCTEST_UI_TEST_PACKAGE")
	if testPackageARN == nil || err != nil {
		// TODO: Upload file if it does not exist in Device Farm
		return nil, fmt.Errorf("failed to retrieve test package arn: %w", err)
	}
	fmt.Println("testPackageARN", *testPackageARN)

	testSpecArn, err := getARNforUpload(appConfig.ProjectARN, "XCTEST_UI_TEST_SPEC")
	if testSpecArn == nil || err != nil {
		// TODO: Upload file if it does not exist in Device Farm
		return nil, fmt.Errorf("failed to retrieve test spec arn: %w", err)
	}
	fmt.Println("testSpecArn", *testSpecArn)

	extraDataSpecArn, err := getARNforUpload(appConfig.ProjectARN, "EXTERNAL_DATA")
	if extraDataSpecArn == nil || err != nil {
		// TODO: Upload file if it does not exist in Device Farm
		return nil, fmt.Errorf("failed to retrieve extra data arn: %w", err)
	}
	fmt.Println("extraDataSpecArn", *extraDataSpecArn)

	devicePoolARN, err := listDevicePools(appConfig.ProjectARN)
	if devicePoolARN == nil || err != nil {
		return nil, fmt.Errorf("failed to retrieve device pool arn: %w", err)
	}
	fmt.Println("devicePoolARN", *devicePoolARN)

	// Scheduling the test run
	test := &devicefarm.ScheduleRunTest{
		Type:           aws.String("XCTEST_UI"),
		TestSpecArn:    testSpecArn,
		TestPackageArn: testPackageARN,
	}

	runConfig := &devicefarm.ScheduleRunConfiguration{
		ExtraDataPackageArn: extraDataSpecArn,
	}

	runResp, err := client.ScheduleRun(&devicefarm.ScheduleRunInput{
		ProjectArn:    aws.String(appConfig.ProjectARN),
		AppArn:        appPackageARN,
		DevicePoolArn: devicePoolARN,
		Name:          aws.String(unique),
		Configuration: runConfig,
		Test:          test,
	})

	if err != nil {
		fmt.Printf("Error scheduling run: %v\n", err)
		return nil, err
	}
	runArn := runResp.Run.Arn
	fmt.Printf("Run %s is scheduled as arn %s\n", unique, *runArn)

	return runArn, nil
}

func PollTestRun(config config.Config, testRunArn *string, fn func() error) error {
	input := &devicefarm.GetRunInput{
		Arn: testRunArn,
	}

	for {
		result, err := client.GetRun(input)
		if err != nil {
			return fmt.Errorf("failed to get test run: %w", err)
		}

		if result.Run != nil && result.Run.Result != nil {
			runResult := *result.Run.Result

			if runResult == "PASSED" {
				fmt.Println("scheduled test passed, downloading artifacts")
				fn()
				return nil
			}

			if runResult == "PENDING" {
				fmt.Println("scheduled test is pending, sleep for 30s")
				time.Sleep(30 * time.Second)
				continue
			}

			if runResult != "PENDING" {
				fmt.Println("scheduled test returned with unexpected status", runResult)
				return fmt.Errorf("test run returned unexpected status: %s", runResult)
			}
		}
	}
}

func DownloadArtifacts(testRunArn *string) error {
	return listAndDownloadArtifacts(testRunArn)
}
