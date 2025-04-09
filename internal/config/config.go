package config

type AppFlags struct {
	URL         string
	Token       string
	Output      string
	Recursive   bool
	Concurrency int
	Verbose     bool
	ZipOutput   bool
	Preview     bool
	Update      bool
	ListFile    string
	Retries     int
	User        string
	Interactive bool
}

const (
	AppName    = "gitdig"
	AppVersion = "1.1.0"
)
