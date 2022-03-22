package cmd

import "github.com/kelseyhightower/envconfig"

const (
	//AppName contains a service name prefix which used in ENV variables
	AppName = "PGLSN"
)

// Config contains ENV variables
type Config struct {
	DbHost string `split_words:"true" required:"true"`
	DbPort string `split_words:"true" required:"true"`
	DbUser string `split_words:"true" required:"true"`
	DbPass string `split_words:"true" required:"true"`
	DbName string `split_words:"true" required:"true"`

	Slot string `split_words:"true" required:"true"`
	Lsn  string `split_words:"true"`

	TableNames string `split_words:"true"`
	Chunks     string `split_words:"true"`

	KafkaHosts string `split_words:"true" required:"true"`
}

// NewConfig loads ENV variables to Config structure
func NewConfig() (*Config, error) {
	var c Config
	if err := envconfig.Process(AppName, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
