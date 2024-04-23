package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tsel-ticketmaster/tm-order/config"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var db *sql.DB
var dbSyncOnce sync.Once

func GetDatabase() *sql.DB {
	dbSyncOnce.Do(func() {
		db = createConnection()
	})

	return db
}

func createConnection() *sql.DB {
	cfg := config.Get()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Postgresql.Host, cfg.Postgresql.Port, cfg.Postgresql.User, cfg.Postgresql.Password, cfg.Postgresql.DBName, cfg.Postgresql.SSLMode)
	conn, err := otelsql.Open(
		"pgx",
		dsn,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName(cfg.Postgresql.DBName),
	)

	if err != nil {
		log.Println(err)
		return nil
	}

	conn.SetMaxOpenConns(cfg.Postgresql.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.Postgresql.MaxIdleConns)

	return conn
}
