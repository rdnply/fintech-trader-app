package handlers

import (
	"cw1/internal/format"
	"cw1/internal/robot"
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) createRobot(w http.ResponseWriter, r *http.Request) error {
	var rbt robot.Robot

	err := json.NewDecoder(r.Body).Decode(&rbt)
	if err != nil {
		return NewHTTPError("Can't unmarshal input json for creating robot", err, "", http.StatusBadRequest)
	}

	token := tokenFromReq(r)

	u, err := h.sessionStorage.FindByToken(token)
	if err != nil{
		return NewHTTPError("Can't find owner by token in storage", err, "", http.StatusInternalServerError)
	}

	if u.UserID == BottomLineValidID {
		s := fmt.Sprintf("can't find owner")
		return NewHTTPError("Can't find owner by token", nil, s, http.StatusBadRequest)
	}

	rbt.OwnerUserID = u.UserID

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
	if err != nil{
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


	if rbtFromDB.RobotID == BottomLineValidID  || rbtFromDB.DeletedAt.Valid {
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
