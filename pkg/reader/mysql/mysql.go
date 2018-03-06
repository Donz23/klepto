package mysql

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/hellofresh/klepto/pkg/reader"
)

type driver struct{}

func (m *driver) IsSupported(dsn string) bool {
	if dsn == "" {
		return false
	}

	_, err := mysql.ParseDSN(dsn)
	return err == nil
}

func (m *driver) NewConnection(opts reader.ConnOpts) (reader.Reader, error) {
	conn, err := sql.Open("mysql", opts.DSN)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(opts.MaxConns)
	conn.SetMaxIdleConns(opts.MaxIdleConns)
	conn.SetConnMaxLifetime(opts.MaxConnLifetime)

	return NewStorage(conn, opts.Timeout), nil
}

func init() {
	reader.Register("mysql", &driver{})
}
