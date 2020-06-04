package handler

import (
	"crypto/sha256"
	"cw1/cmd/auth-api/render"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const BottomLineValidID = 0

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json for sign up: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	u.Password, err = generateHash(u.Password)
	if err != nil {
		h.logger.Errorf("can't generate hash for password: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		h.logger.Errorf("can't find user with id: %v: %v", u.ID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if fromDB.ID == BottomLineValidID {
		err = h.userStorage.Create(&u)
		if err != nil {
			h.logger.Errorf("can't create user record in storage with id: %v: %v", u.ID, err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}

		w.WriteHeader(http.StatusCreated)
	} else {
		h.logger.Errorf("user with email: %v is already exist", u.Email)
		msg := fmt.Sprintf("user %s is already registered", u.Email)
		render.HTTPError(msg, http.StatusConflict, w)
		return
	}
}

func generateHash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", errors.Wrap(err, "can't generate hash")
	}

	return string(hash), nil
}

func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json for sign in: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		h.logger.Errorf("can't find user by id: %v: %v", u.ID, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if !isMatch(fromDB.Password, u.Password) || fromDB.Email != u.Email {
		h.logger.Errorf("can't authorize because password or email: %v incorrect: %v", u.Email, err)
		render.HTTPError("incorrect email or password", http.StatusBadRequest, w)
		return
	}

	token, err := generateToken(u.Email + u.Password)
	if err != nil {
		h.logger.Errorf("can't create new token: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	s, err := session.New(token, fromDB.ID)
	if err != nil {
		h.logger.Errorf("can't create struct for session: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = h.sessionStorage.Create(s)
	if err != nil {
		h.logger.Errorf("can't create session in storage: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondJSON(w, map[string]string{"bearer": token})
	if err != nil {
		h.logger.Errorf("can't respond json with token: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}
}

func generateToken(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "can't generate token")
	}

	hasher := sha256.New()

	_, err = hasher.Write(hash)
	if err != nil {
		return "", errors.Wrap(err, "can't write hash")
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func isMatch(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))

	return err == nil
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("can't unmarshal input json for updating user: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	id, err := IDFromParams(r)
	if err != nil {
		h.logger.Errorf("can't get ID from URL params: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if id <= BottomLineValidID {
		h.logger.Errorf("don't valid id: %v", id)
		msg := fmt.Sprintf("incorrect id: %v", id)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		h.logger.Errorf("can't find session by user ID: %v: %v", id, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if token == s.SessionID {
		msg, status, err := checkEmail(h.userStorage, u.Email, id)
		if err != nil {
			h.logger.Errorf("can't init user: %v", id, err)
			render.HTTPError(msg, status, w)
			return
		}

		err = initUser(&u, id)
		if err != nil {
			h.logger.Errorf("can't init user: %v", id, err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}

		err = h.userStorage.Update(&u)
		if err != nil {
			h.logger.Errorf("can't update user with id= %v: %v", id, err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}

		err = respondJSON(w, &u)
		if err != nil {
			h.logger.Errorf("can't respond json with user info: %v", err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}
	} else {
		h.logger.Errorf("can't respond json with user info: %v", err)
		msg := fmt.Sprintf("user %d don't find", id)
		render.HTTPError(msg, http.StatusNotFound, w)
	}
}

func checkEmail(userStorage user.Storage, email string, id int64) (string, int, error) {
	fromDB, err := userStorage.FindByEmail(email)
	if err != nil {
		info := fmt.Sprintf("can't find user with email: %v", email)
		return "", http.StatusInternalServerError, errors.Wrapf(err, info)
	}

	if fromDB.ID != BottomLineValidID && fromDB.ID != id {
		info := fmt.Sprintf("new user's email: %v, is already exist", email)
		msg := fmt.Sprintf("user %s is already registered", email)

		return msg, http.StatusBadRequest, errors.New(info)
	}

	return "", -1, nil
}

func initUser(u *user.User, id int64) error {
	t, err := format.NewNullTime()
	if err != nil {
		return errors.Wrap(err, "can't create new null time")
	}

	u.ID = id
	u.UpdatedAt = *t

	pass, err := generateHash(u.Password)
	if err != nil {
		return errors.Wrap(err, "can't generate hash")
	}

	u.Password = pass

	return nil
}

func IDFromParams(r *http.Request) (int64, error) {
	const IDIndex = 4 // space in first place

	str := r.URL.String()
	params := strings.Split(str, "/")

	id, err := strconv.ParseInt(params[IDIndex], 10, 64)
	if err != nil {
		return -1, errors.Wrap(err, "can't parse string to int for get id from params")
	}

	return id, nil
}

func tokenFromReq(r *http.Request) string {
	const TokenID = 1

	token := r.Header.Get("Authorization")
	s := strings.Split(token, " ")

	return s[TokenID]
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := IDFromParams(r)
	if err != nil {
		h.logger.Errorf("can't get ID from URL params: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if id <= BottomLineValidID {
		h.logger.Errorf("don't valid id: %v", id)
		msg := fmt.Sprintf("incorrect id: %v", id)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	tokenFromReq := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		h.logger.Errorf("can't find session by user's ID: %v: %v", id, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if tokenFromReq == s.SessionID {
		var u *user.User

		u, err = h.userStorage.FindByID(id)
		if err != nil {
			h.logger.Errorf("can't find user in storage by ID: %v: %v", id, err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}

		err = respondJSON(w, u)
		if err != nil {
			h.logger.Errorf("can't respond json with user info: %v", err)
			render.HTTPError("", http.StatusInternalServerError, w)
			return
		}
	} else {
		h.logger.Errorf("don't contains same token in storage: %v")
		msg := fmt.Sprintf("don't find user with ID %v", id)
		render.HTTPError(msg, http.StatusNotFound, w)
		return
	}
}

func (h *Handler) getUserRobots(w http.ResponseWriter, r *http.Request) {
	id, err := IDFromParams(r)
	if err != nil {
		h.logger.Errorf("can't get ID from URL params: %v", err)
		render.HTTPError("", http.StatusBadRequest, w)
		return
	}

	if id <= BottomLineValidID {
		h.logger.Errorf("don't valid id: %v", id)
		msg := fmt.Sprintf("incorrect id: %v", id)
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		h.logger.Errorf("can't find user by id in storage: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	if s.UserID == BottomLineValidID {
		h.logger.Errorf("can't find user with id: %v: %v", id, err)
		msg := fmt.Sprintf("can't find user with id: %v", id)
		render.HTTPError(msg, http.StatusNotFound, w)
		return
	}

	if token != s.SessionID {
		h.logger.Errorf("incorrect token")
		msg := fmt.Sprintf("tokens don't match")
		render.HTTPError(msg, http.StatusBadRequest, w)
		return
	}

	robots, err := h.robotStorage.FindByOwnerID(id)
	if err != nil {
		h.logger.Errorf("can't get robots with owner id: %v from storage: %v", id, err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}

	err = respondWithData(w, r, h.tmplts, robots...)
	if err != nil {
		h.logger.Errorf("can't respond with data: %v", err)
		render.HTTPError("", http.StatusInternalServerError, w)
		return
	}
}

func respondWithData(w http.ResponseWriter, r *http.Request, tmplts map[string]*template.Template, rbts ...*robot.Robot) error {
	v := r.Header.Get("Accept")

	switch v {
	case "application/json":
		return respondJSON(w, rbts)
	case "text/html":
		sort.SliceStable(rbts, func(i, j int) bool {
			return rbts[i].RobotID < rbts[j].RobotID
		})

		return renderTemplate(w, "index", "base", tmplts, rbts)
	default:
		return errors.New("info's type is absent")
	}
}

func renderTemplate(w io.Writer, name string, template string, tmplts map[string]*template.Template, payload interface{}) error {
	tmpl, ok := tmplts[name]
	if !ok {
		info := fmt.Sprintf("can't find template with name: %v", name)
		return errors.New(info)
	}

	err := tmpl.ExecuteTemplate(w, template, payload)
	if err != nil {
		info := fmt.Sprintf("can't execute template with name: %v", name)
		return errors.New(info)
	}

	return nil
}

func respondJSON(w http.ResponseWriter, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "can't marshal respond to json")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	c, err := w.Write(response)
	if err != nil {
		msg := fmt.Sprintf("can't write json data in respond, code: %v", c)
		return errors.Wrapf(err, msg)
	}

	return nil
}
