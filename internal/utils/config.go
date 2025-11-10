package utils

import (
    "os"
    "path/filepath"
)

type Config struct {
    DBPath          string
    DefaultTaskTime string // e.g., "12:00"
}

func GetConfigDir() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    configDir := filepath.Join(home, ".config", "recur")
    return configDir, nil
}

func EnsureConfigDir() error {
    configDir, err := GetConfigDir()
    if err != nil {
        return err
    }
    return os.MkdirAll(configDir, 0755)
}

func GetDefaultConfig() *Config {
    configDir, _ := GetConfigDir()
    return &Config{
        DBPath:          filepath.Join(configDir, "tasks.db"),
        DefaultTaskTime: "12:00",
    }
}
