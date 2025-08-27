package main

import (
	"flag"
	"golang-server/internal/api"
	"golang-server/internal/config"
	"golang-server/internal/server"
	"golang-server/internal/storage/postgres"
	"log"
	"os"
)

func main() {
	configPath := getConfigPath()
	configFile, err := openConfigFile(configPath)
	if err != nil {
		log.Fatalf("Не удалось открыть файл конфигурации: %v", err)
	}
	defer closeFile(configFile)

	cfg := config.LoadConfig(configFile)

	db, err := postgres.GetInstance(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB(db)

	router := setupRouter(db)
	httpServer := server.NewHTTPServer(cfg.ServerConfig, router)

	log.Println("База данных успешно подключена")
	log.Printf("Сервер запущен на порту: %s", cfg.ServerConfig.Port)

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// getConfigPath выбирает путь к конфигурационному файлу
func getConfigPath() string {
	configFlag := flag.String("config", "", "путь к файлу конфигурации")
	flag.Parse()

	if *configFlag != "" {
		return *configFlag
	}
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		return env
	}
	return "config/local.json"
}

// openConfigFile открывает конфигурационный файл
func openConfigFile(path string) (*os.File, error) {
	return os.Open(path)
}

// closeFile закрывает файл с логированием ошибок
func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("Ошибка закрытия файла: %v", err)
	}
}

// closeDB закрывает соединение с базой данных
func closeDB(db *postgres.PgDB) {
	if err := db.Close(); err != nil {
		log.Printf("Ошибка закрытия БД: %v", err)
	}
}

// setupRouter настраивает маршруты и middleware
func setupRouter(db *postgres.PgDB) *api.Router {
	transactionHandler := api.NewTransactionHandler(db)
	walletHandler := api.NewWalletHandler(db)

	r := api.NewRouter()
	r.Use(api.RecoveryMiddleware)
	r.Use(api.LoggingMiddleware)
	r.RegisterRoute("/api/send", transactionHandler)
	r.RegisterRoute("/api/transactions", transactionHandler)
	r.RegisterRoute("/api/wallet/", walletHandler)

	return r
}
