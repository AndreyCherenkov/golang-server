package config

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"
)

// duration обёртка для time.Duration с поддержкой JSON-строкового формата.
type duration time.Duration

// Config хранит конфигурацию приложения, включая базу данных и сервер.
type Config struct {
	DbConfig     DbConfig     `json:"db_config"`
	ServerConfig ServerConfig `json:"server_config"`
}

// ServerConfig хранит настройки сервера.
type ServerConfig struct {
	Host        string   `json:"host"`         // Адрес хоста, например "127.0.0.1"
	Port        string   `json:"port"`         // Порт сервера, например "8080"
	Timeout     duration `json:"timeout"`      // Время ожидания запроса
	IdleTimeout duration `json:"idle_timeout"` // Время простоя соединения
}

// DbConfig хранит настройки подключения к базе данных.
type DbConfig struct {
	DbName     string `json:"dbname"`      // Имя базы данных
	User       string `json:"user"`        // Пользователь базы данных
	Password   string `json:"password"`    // Пароль пользователя
	Host       string `json:"host"`        // Адрес базы данных
	Port       string `json:"port"`        // Порт базы данных
	SSLMode    string `json:"sslmode"`     // Режим SSL (disable, require и т.д.)
	InitScript string `json:"init_script"` // Путь к скрипту инициализации
}

var (
	cfg  *Config
	once sync.Once
)

// LoadConfig загружает конфигурацию из io.Reader (например, файл или сетевой источник).
// Использует singleton-подход: конфигурация загружается только один раз.
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

// UnmarshalJSON реализует кастомный разбор JSON для duration.
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

// SetDefaults устанавливает значения по умолчанию для сервера.
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
