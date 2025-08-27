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

// PgDB обертка над *sql.DB
type PgDB struct {
	*sql.DB
}

// GetInstance возвращает singleton PgDB, создавая подключение и выполняя init скрипт при первом вызове.
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

func insertWallet(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO wallets (id, balance) VALUES (1, 10000)")
	if err != nil {
		return err
	}
	log.Println("Inserted default wallet")
	return nil
}
