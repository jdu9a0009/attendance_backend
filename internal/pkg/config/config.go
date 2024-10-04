package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ErrorBotToken string   `json:"ERROR_BOT_TOKEN" yaml:"ERROR_BOT_TOKEN"`
	ErrorChatID   []string `json:"ERROR_CHAT_ID" yaml:"ERROR_CHAT_ID"`
	DBUsername    string   `yaml:"db_username"`
	DBPassword    string   `yaml:"db_password"`
	DBHost        string   ` yaml:"db_host"`
	DBPort        string   `yaml:"port"`
	DBName        string   `yaml:"db_name"`
	DisableTLS    bool     `yaml:"disable_tls"`
	BaseUrl       string   `yaml:"base_url"`
	JWTKey        string   `yaml:"jwt_key"`
}

func NewConfig() *Config {
	var c *Config

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
