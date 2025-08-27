package postgres

import (
	"database/sql"
	"fmt"
	"golang-server/internal/config"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	postgresInstance *PgDB
	once             sync.Once
)

// PgDB представляет собой обертку над стандартным *sql.DB.
// Используется для централизованного управления подключением к PostgreSQL.
type PgDB struct {
	*sql.DB
}

// GetInstance возвращает singleton-экземпляр PgDB.
// При первом вызове:
//  1. Создает подключение к базе данных по конфигурации.
//  2. Проверяет соединение с помощью Ping.
//  3. Выполняет SQL init-скрипт, если он указан в конфигурации.
//
// Параметры:
//   - configuration: конфигурация приложения с настройками базы данных.
//
// Возвращает:
//   - *PgDB: экземпляр базы данных.
//   - error: ошибку при создании или инициализации базы данных.
func GetInstance(configuration *config.Config) (*PgDB, error) {
	var err error
	cfg := configuration.DbConfig
	once.Do(func() {
		connStr := fmt.Sprintf("user=%s password=%s port=%s dbname=%s sslmode=%s host=%s",
			cfg.User,
			cfg.Password,
			cfg.Port,
			cfg.DbName,
			cfg.SSLMode,
			cfg.Host,
		)

		db, e := sql.Open("postgres", connStr)
		if e != nil {
			err = fmt.Errorf("failed to open db connection: %w", e)
			return
		}

		if e = db.Ping(); e != nil {
			err = fmt.Errorf("failed to ping db: %w", e)
			return
		}

		if err = ExecuteInitScripts(db, configuration); err != nil {
			_ = db.Close()
			err = fmt.Errorf("init scripts error: %w", err)
			return
		}

		postgresInstance = &PgDB{db}
	})

	return postgresInstance, err
}

// ExecuteInitScripts выполняет SQL-инициализацию базы данных.
// 1. Читает SQL-скрипт из файла, указанного в конфигурации.
// 2. Выполняет SQL-операции.
// 3. Проверяет наличие данных в таблице wallets и добавляет дефолтный кошелек, если таблица пустая.
// Параметры:
//   - db: открытое подключение к базе данных.
//   - configuration: конфигурация с путем к init-скрипту.
//
// Возвращает:
//   - error: ошибку при чтении файла, выполнении скрипта или вставке данных.
func ExecuteInitScripts(db *sql.DB, configuration *config.Config) error {
	cfg := configuration.DbConfig
	sqlBytes, err := os.ReadFile(cfg.InitScript)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	sqlStatements := string(sqlBytes)

	_, err = db.Exec(sqlStatements)
	if err != nil {
		return fmt.Errorf("failed to execute SQL script: %w", err)
	}

	var count int
	selectQuery := "SELECT COUNT(*) FROM wallets"
	err = db.QueryRow(selectQuery).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query wallets count: %w", err)
	}

	if count == 0 {
		if err := insertWallet(db); err != nil {
			return fmt.Errorf("failed to insert wallet: %w", err)
		}
	}
	return nil
}

// insertWallet добавляет дефолтный кошелек в таблицу wallets.
// Значения по умолчанию: id = 1, balance = 10000.
// Используется внутри ExecuteInitScripts.
// Параметры:
//   - db: открытое подключение к базе данных.
//
// Возвращает:
//   - error: ошибку при вставке данных.
func insertWallet(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO wallets (id, balance) VALUES (1, 100)")
	if err != nil {
		return err
	}
	log.Println("Inserted default wallet")
	return nil
}
