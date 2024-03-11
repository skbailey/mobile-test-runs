# Automated Screenshot Generation

Use AWS Device Farm to run Xcode UI tests that generate screenshots of individual screens.

1. Set Environment Variables
```bash
// AWS Credentials
export AWS_ACCESS_KEY_ID="access-key-id"
export AWS_SECRET_ACCESS_KEY="secret-access-key"
export AWS_SESSION_TOKEN="session-token"

// Application Env Vars
export DF_PROJECTARN="arn-for-device-farm-project"
export DF_NAMEPREFIX="ScreenshotTestRun"
export DF_APPFILEPATH="App.ipa"
export DF_TESTPACKAGE="AppUITests-Runner.ipa"
```

2. Run program
```bash
go run main.go
```