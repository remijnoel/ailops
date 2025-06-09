package cmd

import (
    "fmt"
    "os"
    "strings"
    "github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_CONFIG = `
# Default configuration for AILOPS
log_level: warn
cmd_whitelist:
cmd_blacklist:
azure_openai_api_version: "2024-12-01-preview"
azure_openai_endpoint: "https://ailops.cognitiveservices.azure.com/openai/deployments/gpt-4.1-mini/"
initial_commands:
  - "top -b -n1 | head -20"
  - "ps aux | head -10"
  - "df -h"
  - "free -h"
  - "dmesg | tail -n 50"
`

func config(userConfigPath string){
    viper.SetConfigType("yaml")

    // Load embedded default config
	log.Info("Loading default configuration")
    if err := viper.ReadConfig(strings.NewReader(string(DEFAULT_CONFIG))); err != nil {
        fmt.Fprintf(os.Stderr, "failed to load default config: %v\n", err)
        os.Exit(1)
    }

    // If user provides config file, merge it
    if userConfigPath != "" {
		log.Infof("Loading user configuration from %s", userConfigPath)
        viper.SetConfigFile(userConfigPath)
        if err := viper.MergeInConfig(); err != nil {
            fmt.Fprintf(os.Stderr, "failed to load user config: %v\n", err)
            os.Exit(1)
        }
    }else {
		log.Info("No user configuration file provided, using defaults")
	}

    // Read from environment variables (override previous)
    viper.SetEnvPrefix("AILOPS")
	log.Debugf("Loading environment variables with prefix %s", viper.GetEnvPrefix())
    viper.AutomaticEnv() // Maps env vars to keys
}