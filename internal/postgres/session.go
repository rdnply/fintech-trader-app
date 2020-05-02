package postgres

import (
	"cw1/internal/session"
	"database/sql"
	"github.com/pkg/errors"
)

var _ session.Storage = &SessionStorage{}

type SessionStorage struct {
	statementStorage

	createStmt *sql.Stmt
	findStmt   *sql.Stmt
}

func NewSessionStorage(db *DB) (*SessionStorage, error) {
	s := &SessionStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createSessionQuery, Dst: &s.createStmt},
		{Query: findSessionQuery, Dst: &s.findStmt},
	}

	if err := s.initStatements(stmts); err != nil {
		return nil, errors.Wrap(err, "can't init statements")
	}

	return s, nil
}

const sessionFields = "session_id, user_id, created_at, valid_until"

func scanSession(scanner sqlScanner, s *session.Session) error {
	return scanner.Scan(&s.SessionID, &s.UserID, &s.CreatedAt, &s.ValidUntil)
}

const createSessionQuery = "INSERT INTO sessions(" + sessionFields + ") VALUES ($1, $2, $3, $4) RETURNING session_id"

func (st *SessionStorage) Create(s *session.Session) error {
	if err := st.createStmt.QueryRow(s.SessionID, s.UserID, s.CreatedAt, s.ValidUntil).Scan(&s.SessionID); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}

const findSessionQuery = "SELECT " + sessionFields + " FROM sessions WHERE user_id=$1"

func (st *SessionStorage) Find(userID int64) (*session.Session, error) {
	var s session.Session
	row := st.findStmt.QueryRow(userID)
	if err := scanSession(row, &s); err != nil {
		if err == sql.ErrNoRows {
			return &s, nil
		}

		return &s, errors.Wrap(err, "can't scan session")
	}

	return &s, nil
}
