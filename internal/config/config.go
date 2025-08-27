package config

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
)

type duration time.Duration

type Config struct {
	DbConfig     DbConfig     `json:"db_config"`
	ServerConfig ServerConfig `json:"server_config"`
}

type ServerConfig struct {
	Host        string   `json:"host"`
	Port        string   `json:"port"`
	Timeout     duration `json:"timeout"`
	IdleTimeout duration `json:"idle_timeout"`
}

type DbConfig struct {
	DbName     string `json:"dbname"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       string `json:"port"`
	SSLMode    string `json:"sslmode"`
	InitScript string `json:"init_script"`
}

var (
	cfg  *Config
	once sync.Once
)

func LoadConfig(r io.Reader) *Config {
	once.Do(func() {
		if r == nil {
			log.Panic("reader обязателен при первом вызове LoadConfig")
		}
		log.Println("Запуск загрузки конфигурации")

		js, err := io.ReadAll(r)
		if err != nil {
			log.Panic(err)
		}

		var c Config
		if err := json.Unmarshal(js, &c); err != nil {
			log.Panic(err)
		}
		c.SetDefaults()
		cfg = &c

		log.Println("Конфигурация загружена")
	})
	return cfg
}

func (d *duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(dur)
	return nil
}

func (c *Config) SetDefaults() {
	if c.ServerConfig.Host == "" {
		c.ServerConfig.Host = "127.0.0.1"
	}
	if c.ServerConfig.Port == "" {
		c.ServerConfig.Port = "8080"
	}
	if c.ServerConfig.Timeout == 0 {
		c.ServerConfig.Timeout = duration(5 * time.Second)
	}
	if c.ServerConfig.IdleTimeout == 0 {
		c.ServerConfig.IdleTimeout = duration(5 * time.Minute)
	}
}
