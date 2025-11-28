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

func GetDataDir() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    dataDir := filepath.Join(home, ".local", "share", "recur")
    return dataDir, nil
}

func EnsureConfigDir() error {
    configDir, err := GetConfigDir()
    if err != nil {
        return err
    }
    return os.MkdirAll(configDir, 0755)
}

func EnsureDataDir() error {
    dataDir, err := GetDataDir()
    if err != nil {
        return err
    }
    return os.MkdirAll(dataDir, 0755)
}

func GetDefaultConfig() *Config {
    dataDir, _ := GetDataDir()
    return &Config{
        DBPath:          filepath.Join(dataDir, "recur.db"),
        DefaultTaskTime: "12:00",
    }
}
