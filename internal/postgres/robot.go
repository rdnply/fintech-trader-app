package postgres

import (
	"cw1/internal/robot"
	"database/sql"
	"github.com/pkg/errors"
)

var _ robot.Storage = &RobotStorage{}

type RobotStorage struct {
	statementStorage

	createStmt        *sql.Stmt
	findByIDStmt      *sql.Stmt
	findByOwnerIDStmt *sql.Stmt
	findByTickerStmt  *sql.Stmt
	updateStmt        *sql.Stmt
}

func NewRobotStorage(db *DB) (*RobotStorage, error) {
	s := &RobotStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createRobotQuery, Dst: &s.createStmt},
		{Query: findRobotByIDQuery, Dst: &s.findByIDStmt},
		{Query: findRobotByOwnerIDQuery, Dst: &s.findByOwnerIDStmt},
		{Query: findRobotByTickerQuery, Dst: &s.findByTickerStmt},
		{Query: updateRobotQuery, Dst: &s.updateStmt},
	}

	if err := s.initStatements(stmts); err != nil {
		return nil, errors.Wrap(err, "can't init statements")
	}

	return s, nil
}


func scanRobot(scanner sqlScanner, r *robot.Robot) error {
	return scanner.Scan(&r.RobotID, &r.OwnerUserID, &r.ParentRobotID, &r.IsFavourite, &r.IsActive, &r.Ticker, &r.BuyPrice, &r.SellPrice,
		&r.PlanStart, &r.PlanEnd, &r.PlanYield, &r.FactYield, &r.DealsCount, &r.ActivatedAt, &r.DeactivatedAt, &r.CreatedAt, &r.DeletedAt)
}

const robotCreateFields = "owner_user_id, is_favourite, is_active"
const createRobotQuery = "INSERT INTO robots(" + robotCreateFields + ") VALUES ($1, $2, $3) RETURNING robot_id"

func (s *RobotStorage) Create(r *robot.Robot) error {
	if err := s.createStmt.QueryRow(r.OwnerUserID, r.IsFavourite, r.IsActive).Scan(&r.RobotID); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}

const robotFields =
	"owner_user_id, parent_robot_id, is_favourite, is_active, ticker, buy_price, sell_price, plan_start, plan_end, " +
	"plan_yield, fact_yield, deals_count, activated_at, deactivated_at, created_at, deleted_at"
const findRobotByIDQuery = "SELECT robot_id, " + robotFields + " FROM robots WHERE robot_id=$1"

func (s *RobotStorage) FindByID(id int64) (*robot.Robot, error) {
	var r robot.Robot

	row := s.findByIDStmt.QueryRow(id)
	if err := scanRobot(row, &r); err != nil {
		if err == sql.ErrNoRows {
			return &r, nil
		}

		return &r, errors.Wrap(err, "can't scan user")
	}

	return &r, nil
}

const findRobotByOwnerIDQuery = "SELECT robot_id, " + robotFields + " FROM robots WHERE owner_user_id=$1"

func (s *RobotStorage) FindByOwnerID(id int64) ([]*robot.Robot, error) {
	rows, err := s.findByOwnerIDStmt.Query(id)
	if err != nil {
		return nil, errors.Wrap(err, "can't exec query to get robots")
	}

	defer rows.Close()

	robots := make([]*robot.Robot, 0)

	for rows.Next() {
		var r *robot.Robot

		err = scanRobot(rows, r)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan row with robot")
		}

		robots = append(robots, r)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows contain error")
	}

	return robots, nil
}

const findRobotByTickerQuery = "SELECT robot_id, " + robotFields + " FROM robots WHERE ticker=$1"

func (s *RobotStorage) FindByTicker(ticker string) (*robot.Robot, error) {
	var r robot.Robot

	row := s.findByTickerStmt.QueryRow(ticker)
	if err := scanRobot(row, &r); err != nil {
		if err == sql.ErrNoRows {
			return &r, nil
		}

		return &r, errors.Wrap(err, "can't scan robots")
	}

	return &r, nil
}

const updateRobotQuery =
	"UPDATE robots SET " +
	"owner_user_id=$2, parent_robot_id=$3, is_favourite=$4, is_active=$5, ticker=$6, buy_price=$7, " +
	"sell_price=$8, plan_start=$9, plan_end=$10, plan_yield=$11, fact_yield=$12, deals_count=$13, activated_at=$14, deactivated_at=$15, " +
	"created_at=$16, deleted_at=$17 " +
	"WHERE robot_id=$1 RETURNING robot_id, " + robotFields

func (s *RobotStorage) Update(r *robot.Robot) error {
	row := s.updateStmt.QueryRow(r.RobotID, r.OwnerUserID, r.ParentRobotID, r.IsFavourite, r.IsActive, r.Ticker, r.BuyPrice, r.SellPrice,
		r.PlanStart, r.PlanEnd, r.PlanYield, r.FactYield, r.DealsCount, r.ActivatedAt, r.DeactivatedAt, r.CreatedAt, r.DeletedAt)
	if err := scanRobot(row, r); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}

//func robotArgs(r *robot.Robot) []interface{} {
//	return []interface{}{&r.RobotID, &r.OwnerUserID, &r.ParentRobotID, &r.IsFavourite, &r.IsActive, &r.Ticker, &r.BuyPrice, &r.SellPrice,
//		&r.PlanStart, &r.PlanEnd, &r.PlanYield, &r.FactYield, &r.DealsCount, &r.ActivatedAt, &r.DeactivatedAt, &r.CreatedAt, &r.DeletedAt}
//}