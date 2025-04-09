package config

const (
	AppName    = "gitdig"
	AppVersion = "0.0.1"
)

type Config struct {
	Token string
}

type AppFlags struct {
	URL         string
	Token       string
	Output      string
	Recursive   bool
	Concurrency int
	Verbose     bool
}
