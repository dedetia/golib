package sql

import (
	"errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

type (
	Options struct {
		DSN             string
		MaxIdleConn     int
		MaxOpenConn     int
		ConnMaxLifetime time.Duration
	}

	SQL struct {
		DB *sqlx.DB
	}
)

func New(driver Driver, opt *Options) (*SQL, error) {
	if !driver.valid() {
		return nil, errors.New("invalid driver")
	}

	opt.init()

	db, err := connect(driver.String(), opt)
	if err != nil {
		return nil, err
	}

	sql := SQL{
		DB: db,
	}

	return &sql, nil
}

func (opt *Options) init() {
	if opt.MaxIdleConn == 0 {
		opt.MaxIdleConn = 10
	}

	if opt.MaxOpenConn == 0 {
		opt.MaxOpenConn = 30
	}

	if opt.ConnMaxLifetime == 0 {
		opt.ConnMaxLifetime = 5 * time.Minute
	}
}

func connect(driver string, opt *Options) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, opt.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(opt.MaxOpenConn)
	db.SetMaxIdleConns(opt.MaxIdleConn)
	db.SetConnMaxLifetime(opt.ConnMaxLifetime)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
