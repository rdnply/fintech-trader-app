package robot

import (
	"cw1/internal/format"
)

type Robot struct {
	RobotID       int64       `json:"robot_id"`
	OwnerUserID   int64       `json:"owner_user_id"`
	ParentRobotID int64       `json:"parent_user_id,omitempty"`
	IsFavorite    bool        `json:"is_favorite"`
	IsActive      bool        `json:"is_active"`
	Ticker        string      `json:"ticker,omitempty"`
	BuyPrice      float64     `json:"buy_price,omitempty"`
	SellPrice     float64     `json:"sell_price,omitempty"`
	PlanStart     format.Time `json:"plan_start,omitempty"`
	PlanEnd       format.Time `json:"plan_end,omitempty"`
	PlanYield     float64     `json:"plan_yield,omitempty"`
	FactYield     float64     `json:"fact_yield,omitempty"`
	DealsCount    int         `json:"deals_count,omitempty"`
	ActivatedAt   format.Time `json:"activated_at,omitempty"`
	DeactivatedAt format.Time `json:"deactivated_at,omitempty"`
	CreatedAt     format.Time `json:"created_at,omitempty"`
	DeletedAt     format.Time `json:"deleted_at,omitempty"`
}

type Storage interface {
	Create(r *Robot) error
	FindByOwnerID(id int64) (*Robot, error)
	FindByTicker(ticker string) (*Robot, error)
	Update(r *Robot) error
}
