package config

const (
	AppName    = "github-dir-dl"
	AppVersion = "1.0.0"
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
