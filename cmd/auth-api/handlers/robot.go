package handlers

import (
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
