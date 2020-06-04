package postgres

import (
	"cw1/internal/robot"
	"database/sql"

	"github.com/pkg/errors"
)

var _ robot.Storage = &RobotStorage{}

type RobotStorage struct {
	statementStorage

	createStmt                 *sql.Stmt
	findByIDStmt               *sql.Stmt
	findByOwnerIDStmt          *sql.Stmt
	findByTickerStmt           *sql.Stmt
	findByOwnerIDAndTickerStmt *sql.Stmt
	findAllRobotsStmt          *sql.Stmt
	updateStmt                 *sql.Stmt
	updateBesidesActiveStmt    *sql.Stmt
	getActiveRobotsStmt        *sql.Stmt
}

func NewRobotStorage(db *DB) (*RobotStorage, error) {
	s := &RobotStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createRobotQuery, Dst: &s.createStmt},
		{Query: findRobotByIDQuery, Dst: &s.findByIDStmt},
		{Query: findRobotByOwnerIDQuery, Dst: &s.findByOwnerIDStmt},
		{Query: findRobotByTickerQuery, Dst: &s.findByTickerStmt},
		{Query: findRobotByOwnerIDAndTickerQuery, Dst: &s.findByOwnerIDAndTickerStmt},
		{Query: findAllRobotsQuery, Dst: &s.findAllRobotsStmt},
		{Query: updateRobotQuery, Dst: &s.updateStmt},
		{Query: updateRobotBesidesActiveQuery, Dst: &s.updateBesidesActiveStmt},
		{Query: getActiveRobotsQuery, Dst: &s.getActiveRobotsStmt},
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

const robotCreateFields = "owner_user_id, is_favourite, is_active" //nolint: misspell
const createRobotQuery = "INSERT INTO robots(" + robotCreateFields + ") VALUES ($1, $2, $3) RETURNING robot_id"

func (s *RobotStorage) Create(r *robot.Robot) error {
	if err := s.createStmt.QueryRow(r.OwnerUserID, r.IsFavourite, r.IsActive).Scan(&r.RobotID); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}

const robotFields = "owner_user_id, parent_robot_id, is_favourite, is_active, ticker, buy_price, sell_price, plan_start, plan_end, " + //nolint: misspell
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
	return find(s.findByOwnerIDStmt, id)
}

const findRobotByTickerQuery = "SELECT robot_id, " + robotFields + " FROM robots WHERE ticker=$1"

func (s *RobotStorage) FindByTicker(ticker string) ([]*robot.Robot, error) {
	return find(s.findByTickerStmt, ticker)
}

const findRobotByOwnerIDAndTickerQuery = "SELECT robot_id, " + robotFields + " FROM robots WHERE owner_user_id=$1 AND ticker=$2"

func (s *RobotStorage) findByOwnerIDAndTicker(id int64, ticker string) ([]*robot.Robot, error) {
	return find(s.findByOwnerIDAndTickerStmt, id, ticker)
}

const findAllRobotsQuery = "SELECT robot_id, " + robotFields + " FROM robots"

func (s *RobotStorage) findAllRobots() ([]*robot.Robot, error) {
	return find(s.findAllRobotsStmt)
}

func (s *RobotStorage) GetAll(id int64, ticker string) ([]*robot.Robot, error) {
	switch {
	case id != 0 && ticker != "":
		return s.findByOwnerIDAndTicker(id, ticker)
	case id != 0:
		return s.FindByOwnerID(id)
	case ticker != "":
		return s.FindByTicker(ticker)
	default:
		return s.findAllRobots()
	}
}

const updateRobotQuery = "UPDATE robots SET " +
	"owner_user_id=$2, parent_robot_id=$3, is_favourite=$4, is_active=$5, ticker=$6, buy_price=$7, " + //nolint: misspell
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

const updateRobotBesidesActiveQuery = "UPDATE robots SET " +
	"owner_user_id=$2, parent_robot_id=$3, is_favourite=$4, ticker=$5, buy_price=$6, " + //nolint: misspell
	"sell_price=$7, plan_start=$8, plan_end=$9, plan_yield=$10, fact_yield=$11, deals_count=$12, activated_at=$13, deactivated_at=$14, " +
	"created_at=$15, deleted_at=$16 " +
	"WHERE robot_id=$1 RETURNING robot_id, " + robotFields

func (s *RobotStorage) UpdateBesidesActive(r *robot.Robot) error {
	row := s.updateBesidesActiveStmt.QueryRow(r.RobotID, r.OwnerUserID, r.ParentRobotID, r.IsFavourite, r.Ticker, r.BuyPrice, r.SellPrice,
		r.PlanStart, r.PlanEnd, r.PlanYield, r.FactYield, r.DealsCount, r.ActivatedAt, r.DeactivatedAt, r.CreatedAt, r.DeletedAt)
	if err := scanRobot(row, r); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}



const getActiveRobotsQuery = "SELECT robot_id, " + robotFields + " FROM robots " +
	"WHERE is_active=true AND ((plan_start at time zone 'utc')::time <= localtime  AND localtime <= (plan_end at time zone 'utc')::time)"

func (s *RobotStorage) GetActiveRobots() ([]*robot.Robot, error) {
	return find(s.getActiveRobotsStmt)
}

func find(stmt *sql.Stmt, args ...interface{}) ([]*robot.Robot, error) {
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, errors.Wrap(err, "can't exec query to get robots")
	}

	defer rows.Close()

	robots := make([]*robot.Robot, 0)

	for rows.Next() {
		var r robot.Robot

		err = scanRobot(rows, &r)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan row with robot")
		}

		robots = append(robots, &r)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows contain error")
	}

	return robots, nil
}
