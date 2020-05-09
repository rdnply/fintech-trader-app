package handlers

import (
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

func (h *Handler) createRobot(w http.ResponseWriter, r *http.Request) error {
	var rbt robot.Robot

	err := json.NewDecoder(r.Body).Decode(&rbt)
	if err != nil {
		return NewHTTPError("Can't unmarshal input json for creating robot", err, "", http.StatusBadRequest)
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByToken(token)
	if err != nil {
		return NewHTTPError("Can't find owner by token in storage", err, "", http.StatusInternalServerError)
	}

	if s.UserID == BottomLineValidID {
		s := fmt.Sprintf("can't find owner")
		return NewHTTPError("Can't find owner by token", nil, s, http.StatusBadRequest)
	}

	rbt.OwnerUserID = s.UserID

	err = h.robotStorage.Create(&rbt)
	if err != nil {
		return NewHTTPError("Can't create robot record in storage", err, "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *Handler) deleteRobot(w http.ResponseWriter, r *http.Request) error {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, r)
	if err != nil {
		return err
	}

	rbtFromDB, err := getRobot(h.robotStorage, rbtID)
	if err != nil {
		return err
	}

	if rbtFromDB.DeletedAt.V.Valid {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", rbtID)
		s := fmt.Sprintf("robot with id %v don't exist", rbtID)

		return NewHTTPError(ctx, nil, s, http.StatusNotFound)
	}

	if rbtFromDB.OwnerUserID != userID {
		ctx := fmt.Sprintf("User with id: %v don't own robot with id: %v", userID, rbtFromDB.RobotID)
		s := fmt.Sprintf("user with id %v don't own robot with id: %v", userID, rbtID)

		return NewHTTPError(ctx, nil, s, http.StatusBadRequest)
	}

	rbtFromDB.DeletedAt = format.NewNullTime()

	err = h.robotStorage.Update(rbtFromDB)
	if err != nil {
		ctx := fmt.Sprintf("Can't delete robot from storage with rbtID: %v", rbtID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (h *Handler) getRobots(w http.ResponseWriter, r *http.Request) error {
	ownerID, ticker, err := IDAndTickerFromParams(r)
	if err != nil {
		return NewHTTPError("Can't get user's id and/or ticker from URL params", err, "", http.StatusBadRequest)
	}

	if ownerID < BottomLineValidID {
		ctx := fmt.Sprintf("Don't valid ID: %v", ownerID)
		s := fmt.Sprintf("incorrect id: %v", ownerID)

		return NewHTTPError(ctx, nil, s, http.StatusBadRequest)
	}

	token := tokenFromReq(r)

	u, err := h.sessionStorage.FindByToken(token)
	if err != nil {
		return NewHTTPError("Can't find owner by token in storage", err, "", http.StatusInternalServerError)
	}

	if u.UserID == BottomLineValidID {
		s := fmt.Sprintf("can't find user")
		return NewHTTPError("Can't find user by token", nil, s, http.StatusBadRequest)
	}

	robots, err := h.robotStorage.GetAll(ownerID, ticker)
	if err != nil {
		ctx := fmt.Sprintf("Can't get robots from storage (owner's id: %v, ticker: %v)", ownerID, ticker)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	err = respondWithData(w, r, robots, h.tmplts)
	if err != nil {
		return err
	}

	return nil
}

func IDAndTickerFromParams(r *http.Request) (int64, string, error) {
	userStr := r.URL.Query().Get("user")

	var id int64 = BottomLineValidID

	if userStr != "" {
		var err error

		id, err = strconv.ParseInt(userStr, 10, 64)
		if err != nil {
			return -1, "", errors.Wrap(err, "can't parse string to int for get user's id from params")
		}
	}

	tickerStr := r.URL.Query().Get("ticker")

	return id, tickerStr, nil
}

func (h *Handler) makeFavourite(w http.ResponseWriter, r *http.Request) error {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, r)
	if err != nil {
		return err
	}

	rbtFromDB, err := getRobot(h.robotStorage, rbtID)
	if err != nil {
		return err
	}

	rbt := copyForFavourite(rbtFromDB, userID)
	err = h.robotStorage.Create(rbt)
	if err != nil {
		ctx := fmt.Sprintf("Can't create copy for favourite robot with id: %v", rbtID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	err = respondJSON(w, rbt)
	if err != nil {
		return nil
	}

	return nil
}

func copyForFavourite(old *robot.Robot, ownerID int64) *robot.Robot {
	old.OwnerUserID = ownerID
	old.ParentRobotID = format.NewNullInt64(old.RobotID)
	old.IsFavourite = true
	old.IsActive = false

	return old
}

func (h *Handler) activate(w http.ResponseWriter, r *http.Request) error {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, r)
	if err != nil {
		return err
	}

	rbtFromDB, err := getRobot(h.robotStorage, rbtID)
	if err != nil {
		return err
	}

	if !canBeActivated(rbtFromDB, userID) {
		ctx := fmt.Sprintf("Can activate robot with id: %v", rbtID)
		s := fmt.Sprintf("can't activate robot with id: %v", rbtID)
		return NewHTTPError(ctx, err, s, http.StatusBadRequest)
	}

	rbtFromDB.IsActive = true
	rbtFromDB.ActivatedAt = format.NewNullTime()
	err = h.robotStorage.Update(rbtFromDB)
	if err != nil {
		ctx := fmt.Sprintf("Can't create copy for active robot with id: %v", rbtID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	err = respondJSON(w, rbtFromDB)
	if err != nil {
		return nil
	}

	return nil
}

func getRobot(robotStorage robot.Storage, rbtID int64) (*robot.Robot, error) {
	rbtFromDB, err := robotStorage.FindByID(rbtID)
	if err != nil {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", rbtID)
		return nil, NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if rbtFromDB.RobotID == BottomLineValidID {
		ctx := fmt.Sprintf("Robot with id: %v doesn't exist", rbtID)
		s := fmt.Sprintf("can't find robot with id: %v", rbtID)
		return nil, NewHTTPError(ctx, err, s, http.StatusBadRequest)
	}

	return rbtFromDB, nil
}

func getRobotAndUserID(sessionStorage session.Storage, r *http.Request) (int64, int64, error) {
	rbtID, err := IDFromParams(r)
	if err != nil {
		return -1, -1, NewHTTPError("Can't get ID from URL params", err, "", http.StatusBadRequest)
	}

	err = checkIDCorrectness(rbtID)
	if err != nil {
		return -1, -1, err
	}

	token := tokenFromReq(r)

	session, err := sessionStorage.FindByToken(token)
	if err != nil {
		return -1, -1, NewHTTPError("Can't find owner by token in storage", err, "", http.StatusInternalServerError)
	}

	if session.UserID == BottomLineValidID {
		s := fmt.Sprintf("can't find owner")
		return -1, -1, NewHTTPError("Can't find owner by token", nil, s, http.StatusBadRequest)
	}

	return rbtID, session.UserID, nil
}

func canBeActivated(rbt *robot.Robot, userID int64) bool {
	if rbt.OwnerUserID != userID || rbt.IsActive || intoPlanRange(rbt.PlanStart, rbt.PlanEnd) {
		return false
	}

	return true
}

func intoPlanRange(start *format.NullTime, end *format.NullTime) bool {
	t := time.Now()
	switch {
	case start == nil || end == nil:
		return false
	case t.Before(start.V.Time) && t.After(end.V.Time):
		return false
	default:
		return true
	}
}
