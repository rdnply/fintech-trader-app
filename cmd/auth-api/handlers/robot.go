package handlers

import (
	"cw1/internal/format"
	"cw1/internal/robot"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	rbtID, err := IDFromParams(r)
	if err != nil {
		return NewHTTPError("Can't get ID from URL params", err, "", http.StatusBadRequest)
	}

	err = checkIDCorrectness(rbtID)
	if err != nil {
		return err
	}

	token := tokenFromReq(r)

	u, err := h.sessionStorage.FindByToken(token)
	if err != nil {
		return NewHTTPError("Can't find owner by token in storage", err, "", http.StatusInternalServerError)
	}

	if u.UserID == BottomLineValidID {
		s := fmt.Sprintf("can't find owner")
		return NewHTTPError("Can't find owner by token", nil, s, http.StatusBadRequest)
	}

	rbtFromDB, err := h.robotStorage.FindByID(rbtID)
	if err != nil {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", rbtID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if rbtFromDB.RobotID == BottomLineValidID || rbtFromDB.DeletedAt.V.Valid {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", rbtID)
		s := fmt.Sprintf("robot with id %v don't exist", rbtID)

		return NewHTTPError(ctx, nil, s, http.StatusNotFound)
	}

	if rbtFromDB.OwnerUserID != u.UserID {
		ctx := fmt.Sprintf("User with id: %v don't own robot with id: %v", u.UserID, rbtFromDB.RobotID)
		s := fmt.Sprintf("user with id %v don't own robot with id: %v", u.UserID, rbtID)

		return NewHTTPError(ctx, nil, s, http.StatusBadRequest)
	}

	rbtFromDB.DeletedAt = format.NewTime()

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
