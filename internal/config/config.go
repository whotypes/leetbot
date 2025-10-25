package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken                     string
	BotPrefix                        string
	FirestoreProjectID               string
    FirestoreDatabaseID              string
	GoogleApplicationCredentialsPath string
}

func Load() (*Config, error) {

	_ = godotenv.Load()

	config := &Config{
		DiscordToken:                     getEnvVar("DISCORD_TOKEN", ""),
		BotPrefix:                        getEnvVar("BOT_PREFIX", "!"),
		FirestoreProjectID:               getEnvVar("FIRESTORE_PROJECT_ID", ""),
        FirestoreDatabaseID:              getEnvVar("FIRESTORE_DATABASE_ID", ""),
		GoogleApplicationCredentialsPath: getEnvVar("GOOGLE_APPLICATION_CREDENTIALS", ""),
	}

	if config.DiscordToken == "" {
		return nil, ErrMissingDiscordToken
	}

	if config.GoogleApplicationCredentialsPath != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.GoogleApplicationCredentialsPath)
	}

	return config, nil
}

func getEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return ErrMissingDiscordToken
	}

	if c.BotPrefix == "" {
		log.Println("Warning: BOT_PREFIX is empty, using default '!'")
		c.BotPrefix = "!"
	}

	if c.FirestoreProjectID != "" {
		if c.GoogleApplicationCredentialsPath == "" {
			log.Println("Warning: FIRESTORE_PROJECT_ID is set but GOOGLE_APPLICATION_CREDENTIALS is not set")
		}
	}

	return nil
}
