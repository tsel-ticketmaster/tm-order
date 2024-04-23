package postgresql_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsel-ticketmaster/tm-order/pkg/postgresql"
)

func TestGetDatabase(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "root")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_DBNAME", "mydb")
	t.Setenv("POSTGRES_SSLMODE", "disable")

	t.Run("try to build connection and get the db object", func(t *testing.T) {
		db := postgresql.GetDatabase()
		assert.NotNil(t, db, "db object should not be null")
	})
	t.Run("db object is singleton", func(t *testing.T) {
		db1 := postgresql.GetDatabase()
		db2 := postgresql.GetDatabase()

		assert.Equal(t, db1, db2, "both db1 and db2 should have the same reference")
	})
}
