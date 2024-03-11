package config

type Config struct {
	ProjectARN     string
	AppPackageARN  string
	TestPackageARN string
	TestSpecARN    string
	ExtraDataARN   string
	PoolArn        string

	NamePrefix  string
	AppFilePath string
	TestPackage string
}
