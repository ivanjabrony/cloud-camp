package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Env                    string        `json:"env"`
	LogFormat              string        `json:"log_format"`
	TargetURL              *url.URL      `json:"target_url"`
	Port                   int           `json:"port"`
	MaxRetries             int           `json:"max_retries"`
	RepositoryTimeout      time.Duration `json:"repository_timeout"`
	BucketConfigureTimeout time.Duration `json:"bucket_configure_timeout"`
	ShutdownTimeout        time.Duration `json:"shutdown_timeout"`
	UserConfig             UserConfig    `json:"user_config"`
	DB                     DBConfig      `json:"db"`
}

type UserConfig struct {
	Tokens     int     `json:"tokens"`
	RatePerSec float64 `json:"rate_per_sec"`
}

type DBConfig struct {
	MaxConns        int32         `json:"max_conns"`
	MinConns        int32         `json:"min_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	User            string        `json:"-"`
	Password        string        `json:"-"`
	Host            string        `json:"-"`
	Port            string        `json:"-"`
	Name            string        `json:"-"`
}

func (p *DBConfig) GetConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", p.User, p.Password, p.Host, p.Port, p.Name)
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
	type config struct {
		Env                    string   `json:"env"`
		LogFormat              string   `json:"log_format"`
		TargetURL              string   `json:"target_url"`
		Port                   int      `json:"port"`
		MaxRetries             int      `json:"max_retries"`
		RepositoryTimeout      duration `json:"repository_timeout"`
		BucketConfigureTimeout duration `json:"bucket_configure_timeout"`
		ShutdownTimeout        duration `json:"shutdown_timeout"`
		UserConfig             UserConfig
		DB                     struct {
			MaxConns        int32    `json:"max_conns"`
			MinConns        int32    `json:"min_conns"`
			MaxConnLifetime duration `json:"max_conn_lifetime"`
			User            string   `json:"-"`
			Password        string   `json:"-"`
			Host            string   `json:"-"`
			Port            string   `json:"-"`
			Name            string   `json:"-"`
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", path)
	}
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("couldn't read config file: %s", path)
	}

	var cfg config
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("couldn't unmarshall config file: %s", path)
	}

	parsedUrl, err := url.Parse(cfg.TargetURL)
	if err != nil {
		log.Fatalf("couldn't parse target url from config file: %s", path)
	}

	cfg.DB.Password = os.Getenv("DATABASE_PASSWORD")
	cfg.DB.User = os.Getenv("DATABASE_USER")
	cfg.DB.Host = os.Getenv("DATABASE_HOST")
	cfg.DB.Port = os.Getenv("DATABASE_PORT")
	cfg.DB.Name = os.Getenv("DATABASE_NAME")

	return &Config{
		cfg.Env,
		cfg.LogFormat,
		parsedUrl,
		cfg.Port,
		cfg.MaxRetries,
		time.Duration(cfg.RepositoryTimeout),
		time.Duration(cfg.BucketConfigureTimeout),
		time.Duration(cfg.ShutdownTimeout),
		cfg.UserConfig,
		DBConfig{cfg.DB.MaxConns,
			cfg.DB.MinConns,
			time.Duration(cfg.DB.MaxConnLifetime),
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.Host,
			cfg.DB.Port,
			cfg.DB.Name},
	}
}
