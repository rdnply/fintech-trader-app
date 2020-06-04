package handler

import (
	"cw1/cmd/auth-api/render"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) createRobot(w http.ResponseWriter, r *http.Request) {
	var rbt robot.Robot

	err := json.NewDecoder(r.Body).Decode(&rbt)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json for creating robot: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByToken(token)
	if err != nil {
		h.logger.Errorf("can't find owner by token in storage: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if s.UserID == BottomLineValidID {
		h.logger.Errorf("can't find owner by token")
		render.HTTPError("can't find owner", http.StatusBadRequest, w)
		return
	}

	rbt.OwnerUserID = s.UserID

	err = h.robotStorage.Create(&rbt)
	if err != nil {
		h.logger.Errorf("can't create robot record in storage: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) deleteRobot(w http.ResponseWriter, r *http.Request) {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, r)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromDB, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if rbtFromDB.DeletedAt != nil {
		h.logger.Errorf("can't find robot with id: %v in storage", rbtID)
		msg := fmt.Sprintf("robot with id %v don't exist", rbtID)
		render.HTTPError(msg, http.StatusNotFound, w)
		return
	}

	if rbtFromDB.OwnerUserID != userID {
		msg := fmt.Sprintf("user with id %v don't own robot with id: %v", userID, rbtID)
		h.logger.Errorf(msg)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	rbtFromDB.DeletedAt, err = format.NewNullTime()
	if err != nil {
		h.logger.Errorf("can't create new null time: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = h.robotStorage.Update(rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't delete robot from storage with rbtID: %v: %v", rbtID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	go h.hub.Broadcast(rbtFromDB)
}

func getRobotAndUserID(sessionStorage session.Storage, r *http.Request) (int64, int64, error) {
	rbtID, err := IDFromParams(r)
	if err != nil {
		return -1, -1, errors.Wrap(err, "can't get ID from URL params")
	}

	if rbtID <= BottomLineValidID {
		return -1, -1, errors.Wrapf(err, "don't valid id: %v", rbtID)
	}

	token := tokenFromReq(r)

	session, err := sessionStorage.FindByToken(token)
	if err != nil {
		return -1, -1, errors.Wrap(err, "can't find owner by token in storage")
	}

	if session.UserID == BottomLineValidID {
		return -1, -1, errors.Wrap(err, "can't find owner by token")
	}

	return rbtID, session.UserID, nil
}

func findRobot(robotStorage robot.Storage, rbtID int64) (*robot.Robot, error) {
	rbtFromDB, err := robotStorage.FindByID(rbtID)
	if err != nil {
		return nil, errors.Wrapf(err, "can't find robot with id: %v in storage", rbtID)
	}

	if rbtFromDB.RobotID == BottomLineValidID {
		return nil, errors.Wrapf(err, "robot with id: %v doesn't exist", rbtID)
	}

	return rbtFromDB, nil
}

func (h *Handler) getRobots(w http.ResponseWriter, r *http.Request) {
	ownerID, ticker, err := IDAndTickerFromParams(r)
	if err != nil {
		h.logger.Errorf("can't get user's id and/or ticker from URL params: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	if ownerID < BottomLineValidID {
		msg := fmt.Sprintf("incorrect id: %v", ownerID)
		h.logger.Errorf(msg)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	token := tokenFromReq(r)

	u, err := h.sessionStorage.FindByToken(token)
	if err != nil {
		h.logger.Errorf("can't find owner by token in storage: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if u.UserID == BottomLineValidID {
		h.logger.Errorf("Can't find user by token")
		msg := fmt.Sprintf("can't find user")
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	robots, err := h.robotStorage.GetAll(ownerID, ticker)
	if err != nil {
		h.logger.Errorf("Can't get robots from storage (owner's id: %v, ticker: %v): %v", ownerID, ticker, err)
		msg := fmt.Sprintf("can't find user")
		render.HTTPError(msg, http.StatusInternalServerError, w)
		return
	}

	err = respondWithData(w, r, h.tmplts, robots...)
	if err != nil {
		h.logger.Errorf("can't respond with data: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}
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

func (h *Handler) makeFavourite(w http.ResponseWriter, rr *http.Request) {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, rr)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromDB, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	rbt := copyForFavourite(rbtFromDB, userID)

	err = h.robotStorage.Create(rbt)
	if err != nil {
		h.logger.Errorf("can't create copy for favourite robot with id: %v: %v", rbtID, err) //nolint: misspell
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondJSON(w, rbt)
	if err != nil {
		h.logger.Errorf("can't respond with json: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	go h.hub.Broadcast(rbt)
}

func copyForFavourite(old *robot.Robot, ownerID int64) *robot.Robot {
	res := old
	res.OwnerUserID = ownerID
	res.ParentRobotID = format.NewNullInt64(old.RobotID)
	res.IsFavourite = true
	res.IsActive = false
	res.ActivatedAt = nil

	return res
}

func (h *Handler) activate(w http.ResponseWriter, rr *http.Request) {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, rr)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromDB, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if !canBeChangeActivation(rbtFromDB, userID) || rbtFromDB.IsActive {
		msg := fmt.Sprintf("can't activate robot with id: %v", rbtID)
		h.logger.Errorf(msg)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	rbtFromDB.IsActive = true

	rbtFromDB.ActivatedAt, err = format.NewNullTime()
	if err != nil {
		h.logger.Errorf("can't create new null time: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = h.robotStorage.Update(rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't create copy for active robot with id: %v: %v", rbtID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondJSON(w, rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't respond with json: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	go h.hub.Broadcast(rbtFromDB)
}

func canBeChangeActivation(rbt *robot.Robot, userID int64) bool {
	if rbt.OwnerUserID != userID || !intoPlanRange(rbt.PlanStart, rbt.PlanEnd) {
		return false
	}

	return true
}

func intoPlanRange(start *format.NullTime, end *format.NullTime) bool {
	t := time.Now()

	switch {
	case start == nil || end == nil:
		return false
	case t.Before(start.V.Time) || t.After(end.V.Time):
		return false
	default:
		return true
	}
}

func (h *Handler) deactivate(w http.ResponseWriter, rr *http.Request) {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, rr)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromDB, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if !canBeChangeActivation(rbtFromDB, userID) || !rbtFromDB.IsActive {
		msg := fmt.Sprintf("can't deactivate robot with id: %v", rbtID)
		h.logger.Errorf(msg)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	rbtFromDB.IsActive = false

	rbtFromDB.DeactivatedAt, err = format.NewNullTime()
	if err != nil {
		h.logger.Errorf("can't create new null time: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = h.robotStorage.Update(rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't create copy for active robot with id: %v: %v", rbtID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondJSON(w, rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't respond with json: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	go h.hub.Broadcast(rbtFromDB)
}

func (h *Handler) getRobot(w http.ResponseWriter, rr *http.Request) {
	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, rr)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromDB, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if rbtFromDB.DeletedAt != nil {
		h.logger.Errorf("can't find robot with id: %v in storage", rbtID)
		msg := fmt.Sprintf("robot with id %v don't exist", rbtID)
		render.HTTPError(msg, http.StatusNotFound, w)
		return
	}

	if rbtFromDB.OwnerUserID != userID {
		h.logger.Errorf("can get robot with id: %v for user with id: %v", rbtID, userID)
		msg := fmt.Sprintf("user with id: %v don't have permission to get robot with id: %v", userID, rbtID)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	err = respondWithData(w, rr, h.tmplts, rbtFromDB)
	if err != nil {
		h.logger.Errorf("can't respond with data: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}
}

func (h *Handler) updateRobot(w http.ResponseWriter, rr *http.Request) {
	var rbt robot.Robot

	err := json.NewDecoder(rr.Body).Decode(&rbt)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json for update robot: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtID, userID, err := getRobotAndUserID(h.sessionStorage, rr)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	rbtFromID, err := findRobot(h.robotStorage, rbtID)
	if err != nil {
		h.logger.Errorf(err.Error())
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if rbtFromID.OwnerUserID != userID {
		msg := fmt.Sprintf("user with id: %v don't have permission to update robot with id: %v", userID, rbtID)
		h.logger.Errorf(msg)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	rbt.RobotID = rbtID

	err = h.robotStorage.Update(&rbt)
	if err != nil {
		h.logger.Errorf("can't update  robot with id: %v in storage: %v", rbtID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondJSON(w, rbt)
	if err != nil {
		h.logger.Errorf("can't respond with json: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	go h.hub.Broadcast(&rbt)
}
