package postgres

import (
	"cw1/pkg/log/logger"
	"database/sql"
	_ "github.com/lib/pq" // init postgres driver
	"github.com/pkg/errors"
	"time"
)

type DB struct {
	Session *sql.DB
	Logger  logger.Logger
}

func New(logger logger.Logger, filename string) (*DB, error) {
	URL, err := ParseConfig(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse configuration for database")
	}

	db, err := sql.Open("postgres", URL)
	if err != nil {
		return nil, errors.Wrap(err, "can't open connection to postgres")
	}

	return &DB{
		Session: db,
		Logger:  logger,
	}, nil
}

func (d *DB) CheckConnection() error {
	var err error
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = d.Session.Ping(); err == nil {
			break
		}
		nextAttemptWait := time.Duration(attempt) * time.Second
		d.Logger.Errorf("attempt %d: can't establish a connection with the db, wait for %v: %s",
			attempt,
			nextAttemptWait,
			err,
		)
		time.Sleep(nextAttemptWait)
	}
	return errors.Wrap(err, "can't connect to db")
}


func (d *DB) Close() error {
	if err := d.Session.Close(); err != nil {
		return errors.Wrap(err, "can't close db")
	}
	return nil
}

type sqlScanner interface {
	Scan(dest ...interface{}) error
}

