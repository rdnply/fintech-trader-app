package postgres

import (
	"cw1/internal/user"
	"database/sql"

	"github.com/pkg/errors"
)

var _ user.Storage = &UserStorage{}

type UserStorage struct {
	statementStorage

	createStmt      *sql.Stmt
	findByEmailStmt *sql.Stmt
	findByIDStmt    *sql.Stmt
	updateStmt      *sql.Stmt
}

func NewUserStorage(db *DB) (*UserStorage, error) {
	s := &UserStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createUserQuery, Dst: &s.createStmt},
		{Query: findUserByEmailQuery, Dst: &s.findByEmailStmt},
		{Query: findUserByIDQuery, Dst: &s.findByIDStmt},
		{Query: updateUserQuery, Dst: &s.updateStmt},
	}

	if err := s.initStatements(stmts); err != nil {
		return nil, errors.Wrap(err, "can't init statements")
	}

	return s, nil
}

func scanUser(scanner sqlScanner, u *user.User) error {
	return scanner.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Birthday, &u.Email, &u.Password, &u.UpdatedAt, &u.CreatedAt)
}

const userCreateFields = "first_name, last_name, birthday, email, password"
const createUserQuery = "INSERT INTO users(" + userCreateFields + ") VALUES ($1, $2, $3, $4, $5) RETURNING id"

func (s *UserStorage) Create(u *user.User) error {
	if err := s.createStmt.QueryRow(u.FirstName, u.LastName, u.Birthday, u.Email, u.Password).Scan(&u.ID); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}

const userFields = "first_name, last_name, birthday, email, password, updated_at, created_at"
const findUserByEmailQuery = "SELECT id, " + userFields + " FROM users WHERE email=$1"

func (s *UserStorage) FindByEmail(email string) (*user.User, error) {
	var u user.User

	row := s.findByEmailStmt.QueryRow(email)
	if err := scanUser(row, &u); err != nil {
		if err == sql.ErrNoRows {
			return &u, nil
		}

		return &u, errors.Wrap(err, "can't scan user")
	}

	return &u, nil
}

const findUserByIDQuery = "SELECT id, " + userFields + " FROM users WHERE id=$1"

func (s *UserStorage) FindByID(id int64) (*user.User, error) {
	var u user.User

	row := s.findByIDStmt.QueryRow(id)
	if err := scanUser(row, &u); err != nil {
		if err == sql.ErrNoRows {
			return &u, nil
		}

		return &u, errors.Wrap(err, "can't scan user")
	}

	return &u, nil
}

const updateUserQuery = "UPDATE users SET first_name=$1, last_name=$2, birthday=$3, email=$4, password=$5, updated_at=$6 " +
	"WHERE id=$7 RETURNING id, " + userFields

func (s *UserStorage) Update(u *user.User) error {
	row := s.updateStmt.QueryRow(u.FirstName, u.LastName, u.Birthday, u.Email, u.Password, u.UpdatedAt, u.ID)
	if err := scanUser(row, u); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}
