package config

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Env                 string        `json:"env"`
	LogFormat           string        `json:"log_format"`
	URLs                []*url.URL    `json:"urls"`
	Port                int           `json:"port"`
	MaxRetries          int           `json:"max_retries"`
	MaxAttempts         int           `json:"max_attempts"`
	ShutdownTimeout     time.Duration `json:"shutdown_timeout"`
	RetryTimeout        time.Duration `json:"retry_timeout"`
	HealthPoolTimeout   time.Duration `json:"health_pool_timeout"`
	HealthServerTimeout time.Duration `json:"health_server_timeout"`
}

type duration time.Duration

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func MustLoadConfig(path string) *Config {
	type uncheckedUrlConfig struct {
		Env                 string   `json:"env"`
		LogFormat           string   `json:"log_format"`
		URLs                []string `json:"urls"`
		Port                int      `json:"port"`
		MaxRetries          int      `json:"max_retries"`
		MaxAttempts         int      `json:"max_attempts"`
		ShutdownTimeout     duration `json:"shutdown_timeout"`
		RetryTimeout        duration `json:"retry_timeout"`
		HealthPoolTimeout   duration `json:"health_pool_timeout"`
		HealthServerTimeout duration `json:"health_server_timeout"`
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", path)
	}
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("couldn't read config file: %s", path)
	}

	var cfg uncheckedUrlConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("couldn't unmarshall config file: %s", path)
	}

	parsedUrls := make([]*url.URL, 0, len(cfg.URLs))

	for _, u := range cfg.URLs {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			log.Fatalf("couldn't parse target url from config file: %s", path)
		}

		parsedUrls = append(parsedUrls, parsedUrl)
	}

	return &Config{
		cfg.Env,
		cfg.LogFormat,
		parsedUrls,
		cfg.Port,
		cfg.MaxRetries,
		cfg.MaxAttempts,
		time.Duration(cfg.ShutdownTimeout),
		time.Duration(cfg.RetryTimeout),
		time.Duration(cfg.HealthPoolTimeout),
		time.Duration(cfg.HealthServerTimeout),
	}
}
