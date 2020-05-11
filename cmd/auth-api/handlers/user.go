package handlers

import (
	"crypto/sha256"
	"cw1/cmd/auth-api/httperror"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const BottomLineValidID = 0

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) error {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return httperror.NewHTTPError("Can't unmarshal input json for sign up", err, "", http.StatusBadRequest)
	}

	u.Password, err = generateHash(u.Password)
	if err != nil {
		return httperror.NewHTTPError("Can't generate hash for password", err, "", http.StatusInternalServerError)
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user with id: %v", u.ID)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if fromDB.ID == BottomLineValidID {
		err = h.userStorage.Create(&u)
		if err != nil {
			ctx := fmt.Sprintf("Can't create user record in storage with id: %v", u.ID)
			return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
	} else {
		s := fmt.Sprintf("user %s is already registered", u.Email)
		return httperror.NewHTTPError("This user already exist, email: "+u.Email, nil, s, http.StatusConflict)
	}

	return nil
}

func generateHash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", errors.Wrap(err, "can't generate hash")
	}

	return string(hash), nil
}

func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) error {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return httperror.NewHTTPError("Can't unmarshal input json for sign in", err, "", http.StatusBadRequest)
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user by id: %v", u.ID)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if !isMatch(fromDB.Password, u.Password) || fromDB.Email != u.Email {
		ctx := fmt.Sprintf("Can't authorize because password or email: %v incorrect", u.Email)
		return httperror.NewHTTPError(ctx, err, "incorrect email or password", http.StatusBadRequest)
	}

	token, err := generateToken(u.Email + u.Password)
	if err != nil {
		return httperror.NewHTTPError("Can't create new token", err, "", http.StatusInternalServerError)
	}

	s, err := session.New(token, fromDB.ID)
	if err != nil {
		return httperror.NewHTTPError("Can't create struct for session", err, "", http.StatusInternalServerError)
	}

	err = h.sessionStorage.Create(s)
	if err != nil {
		return httperror.NewHTTPError("Can't create session in storage", err, "", http.StatusInternalServerError)
	}

	err = respondJSON(w, map[string]string{"bearer": token})
	if err != nil {
		return err
	}

	return nil
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

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) error {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return httperror.NewHTTPError("Can't unmarshal input json for update user", err, "", http.StatusBadRequest)
	}

	id, err := IDFromParams(r)
	if err != nil {
		return httperror.NewHTTPError("Can't get ID from URL params", err, "", http.StatusInternalServerError)
	}

	err = checkIDCorrectness(id)
	if err != nil {
		return nil
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't find session by user ID: %v", id)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if token == s.SessionID {
		err = checkEmail(h.userStorage, u.Email, id)
		if err != nil {
			return err
		}

		err = initUser(&u, id)
		if err != nil {
			return httperror.NewHTTPError("Can't init user", err, "", http.StatusInternalServerError)
		}

		err = h.userStorage.Update(&u)
		if err != nil {
			ctx := fmt.Sprintf("Can't update user with id= %v", id)
			return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		err = respondJSON(w, u)
		if err != nil {
			return nil
		}
	} else {
		s := fmt.Sprintf("user %s is already registered", u.Email)
		return httperror.NewHTTPError("Don't contain same token in storage", nil, s, http.StatusNotFound)
	}

	return nil
}

func checkEmail(userStorage user.Storage, email string, id int64) error {
	fromDB, err := userStorage.FindByEmail(email)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user with email: %v", email)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if fromDB.ID != BottomLineValidID && fromDB.ID != id {
		ctx := fmt.Sprintf("New user's email: %v, is already exist", email)
		s := fmt.Sprintf("user %s is already registered", email)

		return httperror.NewHTTPError(ctx, nil, s, http.StatusBadRequest)
	}

	return nil
}

func initUser(u *user.User, id int64) error {
	t := format.NewNullTime()
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
	str := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(str, 10, 64)
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

func checkIDCorrectness(id int64) error {
	if id <= BottomLineValidID {
		ctx := fmt.Sprintf("Don't valid ID: %v", id)
		s := fmt.Sprintf("incorrect id: %v", id)

		return httperror.NewHTTPError(ctx, nil, s, http.StatusBadRequest)
	}

	return nil
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) error {
	id, err := IDFromParams(r)
	if err != nil {
		return httperror.NewHTTPError("Can't get ID from URL params", err, "", http.StatusInternalServerError)
	}

	err = checkIDCorrectness(id)
	if err != nil {
		return err
	}

	tokenFromReq := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't find session by user's ID: %v", id)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if tokenFromReq == s.SessionID {
		var u *user.User

		u, err = h.userStorage.FindByID(id)
		if err != nil {
			ctx := fmt.Sprintf("Can't find user in storage by ID: %v", id)
			return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		err = respondJSON(w, u)
		if err != nil {
			return nil
		}
	} else {
		s := fmt.Sprintf("user %v is already registered", id)
		return httperror.NewHTTPError("Don't contain same token in storage", err, s, http.StatusNotFound)
	}

	return nil
}

func (h *Handler) getUserRobots(w http.ResponseWriter, r *http.Request) error {
	id, err := IDFromParams(r)
	if err != nil {
		return httperror.NewHTTPError("Can't get ID from URL params", err, "", http.StatusBadRequest)
	}

	err = checkIDCorrectness(id)
	if err != nil {
		return err
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.FindByID(id)
	if err != nil {
		return httperror.NewHTTPError("Can't find user by id in storage", err, "", http.StatusInternalServerError)
	}

	if s.UserID == BottomLineValidID {
		ctx := fmt.Sprintf("Can't find user with id: %v", id)
		s := fmt.Sprintf("can't find user with id: %v", id)

		return httperror.NewHTTPError(ctx, nil, s, http.StatusNotFound)
	}

	if token != s.SessionID {
		return httperror.NewHTTPError("Tokens don't match", nil, "incorrect token", http.StatusBadRequest)
	}

	robots, err := h.robotStorage.FindByOwnerID(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't get robots with owner id: %v from storage", id)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	err = respondWithData(w, r,  h.tmplts, robots...)
	if err != nil {
		return err
	}

	return nil
}

func respondWithData(w http.ResponseWriter, r *http.Request, tmplts map[string]*template.Template, rbts ...*robot.Robot) error {
	t := r.Header.Get("Accept")

	switch t {
	case "application/json":
		return respondJSON(w, rbts)
	case "text/html":
		sort.SliceStable(rbts, func(i, j int) bool {
			return rbts[i].RobotID < rbts[j].RobotID
		})

		return renderTemplate(w, "index", "base", tmplts, rbts)
	default:
		return httperror.NewHTTPError("Info's type is absent", nil, "", http.StatusBadRequest)
	}
}

func renderTemplate(w http.ResponseWriter, name string, template string, tmplts map[string]*template.Template, payload interface{}) error {
	tmpl, ok := tmplts[name]
	if !ok {
		ctx := fmt.Sprintf("Can't find template with name: %v", name)
		return httperror.NewHTTPError(ctx, nil, "", http.StatusInternalServerError)
	}

	err := tmpl.ExecuteTemplate(w, template, payload)
	if err != nil {
		ctx := fmt.Sprintf("Can't execute template with name: %v", name)
		return httperror.NewHTTPError(ctx, nil, "", http.StatusInternalServerError)
	}

	return nil
}

func respondJSON(w http.ResponseWriter, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return httperror.NewHTTPError("Can't marshal respond to json", err, "", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	c, err := w.Write(response)
	if err != nil {
		ctx := fmt.Sprintf("Can't write json data in respond, code: %v", c)
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	return  nil
}
