package persist

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func MustNewMysqlClient(dbURI string) *sqlx.DB {
	const maxOpenConns = 500
	const maxIdleConns = maxOpenConns / 5

	tempDB := sqlx.MustConnect("mysql", dbURI)
	tempDB.SetMaxOpenConns(maxOpenConns)
	tempDB.SetMaxIdleConns(maxIdleConns)
	return tempDB
}

func NewMysqlClient(dbURI string) (*sqlx.DB, error) {
	const maxOpenConns = 500
	const maxIdleConns = maxOpenConns / 5

	tempDB, err := sqlx.Connect("mysql", dbURI)
	if err != nil {
		return nil, err
	}
	tempDB.SetMaxOpenConns(maxOpenConns)
	tempDB.SetMaxIdleConns(maxIdleConns)
	return tempDB, nil
}
