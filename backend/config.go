package main

import (
	"os"
)

type Config struct {
	AllowedOrigin string
	RedirectURL   string
	DBDSN         string
}

func appCfgInit() *Config {
	return &Config{
		AllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
		RedirectURL:   os.Getenv("OAUTH_REDIRECT_URL"),
		DBDSN:         os.Getenv("DB_DSN"),
	}
}
