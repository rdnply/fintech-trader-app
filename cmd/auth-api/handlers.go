package main

import (
	"crypto/sha256"
	"cw1/internal/postgres"
	"cw1/internal/session"
	"cw1/internal/user"
	"cw1/pkg/log/logger"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	logger         logger.Logger
	userStorage    *postgres.UserStorage
	sessionStorage *postgres.SessionStorage
}

func NewHandler(logger logger.Logger, ut *postgres.UserStorage, st *postgres.SessionStorage) *Handler {
	return &Handler{
		logger:         logger,
		userStorage:    ut,
		sessionStorage: st,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", rootHandler{h.signUp, h.logger}.ServeHTTP)
		r.Post("/signin", rootHandler{h.signIn, h.logger}.ServeHTTP)
		r.Put("/users/{id}", rootHandler{h.updateUser, h.logger}.ServeHTTP)
		r.Get("/users/{id}", rootHandler{h.getUser, h.logger}.ServeHTTP)
	})

	return r
}

type rootHandler struct {
	H      func(http.ResponseWriter, *http.Request) error
	logger logger.Logger
}

func (fn rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn.H(w, r)
	if err == nil {
		return
	}

	clientError, ok := err.(ClientError)
	if !ok {
		fn.logger.Errorf("Can't cast error to Client's error: %v", clientError)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	fn.logger.Errorf(clientError.Error())

	body, err := clientError.ResponseBody()
	if err != nil {
		fn.logger.Errorf("Can't get info about error because of : %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	status, headers := clientError.ResponseHeaders()

	for k, v := range headers {
		if body == nil && v == "application/json" {
			continue
		}

		w.Header().Set(k, v)
	}

	w.WriteHeader(status)

	c, err := w.Write(body)
	if err != nil {
		fn.logger.Errorf("Can't write json data in respond, code: %v, error: %v", c, err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

const BottomLineValidID = 0

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) error {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		return NewHTTPError("Can't unmarshal input json for sign up", err, "", http.StatusBadRequest)
	}

	u.Password, err = generateHash(u.Password)
	if err != nil {
		return NewHTTPError("Can't generate hash for password", err, "", http.StatusInternalServerError)
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user with id: %v", u.ID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if fromDB.ID == BottomLineValidID {
		err = h.userStorage.Create(&u)
		if err != nil {
			ctx := fmt.Sprintf("Can't create user record in storage with id: %v", u.ID)
			return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
	} else {
		s := fmt.Sprintf("user %s is already registered", u.Email)
		return NewHTTPError("This user already exist, email: "+u.Email, nil, s, http.StatusConflict)
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
		return NewHTTPError("Can't unmarshal input json for sign in", err, "", http.StatusBadRequest)
	}

	fromDB, err := h.userStorage.FindByEmail(u.Email)
	if err != nil {
		ctx := fmt.Sprintf("Can't find user by id: %v", u.ID)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if !isMatch(fromDB.Password, u.Password) || fromDB.Email != u.Email {
		ctx := fmt.Sprintf("Can't authorize because password or email: %v incorrect", u.Email)
		return NewHTTPError(ctx, err, "incorrect email or password", http.StatusBadRequest)
	}

	token, err := generateToken(u.Email + u.Password)
	if err != nil {
		return NewHTTPError("Can't create new token", err, "", http.StatusInternalServerError)
	}

	s, err := session.New(token, fromDB.ID)
	if err != nil {
		return NewHTTPError("Can't create struct for session", err, "", http.StatusInternalServerError)
	}

	err = h.sessionStorage.Create(s)
	if err != nil {
		return NewHTTPError("Can't create session in storage", err, "", http.StatusInternalServerError)
	}

	respondJSON(w, http.StatusOK, h.logger, map[string]string{"bearer": token})

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
		return NewHTTPError("Can't unmarshal input json for update user", err, "", http.StatusBadRequest)
	}

	id, err := IDFromParams(r)
	if err != nil {
		return NewHTTPError("Can't get ID from URL params", err, "", http.StatusInternalServerError)
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.Find(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't find session by user ID: %v", id)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if token == s.SessionID {
		fromDB, err := h.userStorage.FindByEmail(u.Email)
		if err != nil {
			ctx := fmt.Sprintf("Can't find user with email: %v", u.Email)
			return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		if fromDB.ID != BottomLineValidID && fromDB.ID == id {
			ctx := fmt.Sprintf("New user's email: %v, is already exist", u.Email)
			s := fmt.Sprintf("user %s is already registered", u.Email)

			return NewHTTPError(ctx, nil, s, http.StatusBadRequest)
		}

		err = initUser(&u, id)
		if err != nil {
			return NewHTTPError("Can't init user", err, "", http.StatusInternalServerError)
		}

		err = h.userStorage.Update(&u)
		if err != nil {
			ctx := fmt.Sprintf("Can't update user with id= %v", id)
			return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		u.Password = ""
		respondJSON(w, http.StatusOK, h.logger, u)
	} else {
		s := fmt.Sprintf("user %s is already registered", u.Email)
		return NewHTTPError("Don't contain same token in storage", nil, s, http.StatusNotFound)
	}

	return nil
}

func initUser(u *user.User, id int64) error {
	t := user.NewTime()
	u.ID = id
	u.UpdatedAt = t

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

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) error {
	id, err := IDFromParams(r)
	if err != nil {
		return NewHTTPError("Can't get ID from URL params", err, "", http.StatusInternalServerError)
	}

	if id <= BottomLineValidID {
		ctx := fmt.Sprintf("Don't valid ID: %v", id)
		s := fmt.Sprintf("user %v is already registered", id)

		return NewHTTPError(ctx, err, s, http.StatusBadRequest)
	}

	tokenFromReq := tokenFromReq(r)

	s, err := h.sessionStorage.Find(id)
	if err != nil {
		ctx := fmt.Sprintf("Can't find session by user's ID: %v", id)
		return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	if tokenFromReq == s.SessionID {
		var u *user.User

		u, err = h.userStorage.FindByID(id)
		if err != nil {
			ctx := fmt.Sprintf("Can't find user in storage by ID: %v", id)
			return NewHTTPError(ctx, err, "", http.StatusInternalServerError)
		}

		info := user.NewInfo(u)

		respondJSON(w, http.StatusOK, h.logger, info)
	} else {
		s := fmt.Sprintf("user %v is already registered", id)
		return NewHTTPError("Don't contain same token in storage", err, s, http.StatusNotFound)
	}

	return nil
}

func respondJSON(w http.ResponseWriter, status int, l logger.Logger, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		l.Errorf("Can't marshal respond to json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	c, err := w.Write(response)
	if err != nil {
		l.Errorf("Can't write json data in respond, code: %v, error: %v", c, err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
