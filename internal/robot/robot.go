package robot

import (
	"cw1/internal/format"
)

type Robot struct {
	RobotID       int64              `json:"robot_id"`
	OwnerUserID   int64              `json:"owner_user_id"`
	ParentRobotID format.NullInt64   `json:"parent_user_id,omitempty"`
	IsFavourite   bool               `json:"is_favourite"`
	IsActive      bool               `json:"is_active"`
	Ticker        format.NullString  `json:"ticker,omitempty"`
	BuyPrice      format.NullFloat64 `json:"buy_price,omitempty"`
	SellPrice     format.NullFloat64 `json:"sell_price,omitempty"`
	PlanStart     format.NullTime    `json:"plan_start,omitempty"`
	PlanEnd       format.NullTime    `json:"plan_end,omitempty"`
	PlanYield     format.NullFloat64 `json:"plan_yield,omitempty"`
	FactYield     format.NullFloat64 `json:"fact_yield,omitempty"`
	DealsCount    format.NullInt64   `json:"deals_count,omitempty"`
	ActivatedAt   format.NullTime    `json:"activated_at,omitempty"`
	DeactivatedAt format.NullTime    `json:"deactivated_at,omitempty"`
	CreatedAt     format.NullTime    `json:"created_at,omitempty"`
	DeletedAt     format.NullTime    `json:"deleted_at,omitempty"`
}

type Storage interface {
	Create(r *Robot) error
	FindByID(id int64) (*Robot, error)
	FindByOwnerID(id int64) (*Robot, error)
	FindByTicker(ticker string) (*Robot, error)
	Update(r *Robot) error
}
