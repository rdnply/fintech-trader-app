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

	ownerFromDB, err := h.userStorage.FindByID(rbt.OwnerUserID)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user with id: %v in storage", rbt.OwnerUserID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if ownerFromDB.ID == BottomLineValidID {
		ctx := fmt.Sprintf("User with id: %v don't exist", rbt.OwnerUserID)
		s := fmt.Sprintf("user %v is already registered", rbt.OwnerUserID)

		return NewHTTPError(ctx, err, s, http.StatusBadRequest)
	}

	err = h.robotStorage.Create(&rbt)
	if err != nil {
		return NewHTTPError("Can't create robot record in storage", err, "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *Handler) deleteRobot(w http.ResponseWriter, r *http.Request) error {
	id, err := IDFromParams(r)
	if err != nil {
		return NewHTTPError("Can't get ID from URL params", err, "", http.StatusBadRequest)
	}

	err = checkIDCorrectness(id)
	if err != nil {
		return err
	}

	fromDB, err := h.robotStorage.FindByID(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", id)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if fromDB.RobotID == BottomLineValidID {
		ctx := fmt.Sprintf("Can't find robot with id: %v in storage", id)
		s := fmt.Sprintf("robot with id: %v don't exist", id)

		return NewHTTPError(ctx, nil, s, http.StatusNotFound)
	}

	fromDB.DeletedAt = format.NewTime()

	err = h.robotStorage.Update(fromDB)
	if err != nil {
		ctx := fmt.Sprintf("Can't delete robot from storage with id: %v", id)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}
