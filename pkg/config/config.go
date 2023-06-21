package config

import "github.com/spf13/viper"

// Config is the configuration for gale
type Config struct {
	GhxVersion string `mapstructure:"ghx_version"`
}

// Load loads the configuration from the config file or environment variables
func Load(paths ...string) (*Config, error) {
	viper.AddConfigPath(".")
	viper.AddConfigPath(".gale")

	// apply additional paths if provided
	for _, path := range paths {
		viper.AddConfigPath(path)
	}

	viper.SetConfigName("gale")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// if the config file is not found, we'll ignore the error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
