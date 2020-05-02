package main

import (
	"crypto/md5"
	"cw1/internal/postgres"
	"cw1/internal/session"
	"cw1/internal/user"
	"cw1/pkg/log/logger"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	logger logger.Logger
	userStorage *postgres.UserStorage
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
		r.Post("/signup", h.signUp)
		r.Post("/signin", h.signIn)
		r.Put("/users/{id}", h.updateUser)
		//r.Get("/users/{id}", getUser)
	})
	return r
}


func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("Can't unmarshal input json for sign up: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u.Password, err = generateHash(u.Password)
	if err != nil {
		h.logger.Errorf("Can't generate hash for password: %v", err)
	}

	fromDB, err := h.userStorage.Find(u.Email)
	if err != nil || fromDB != nil{////////////////////////////
		h.logger.Errorf("Can't register user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json := fmt.Sprintf("{\"error\" : \"user %s is already registered\"}", u.Email)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(json))
	} else {
		err = h.userStorage.Create(&u)
		if err != nil {
			h.logger.Errorf("Can't create fromDB: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}
}


func generateHash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", errors.Wrap(err ,"can't generate hash")
	}

	return string(hash), nil
}


func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("Can't unmarshal input json for sign in: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fromDB, err := h.userStorage.Find(u.Email)

	if !isMatch(fromDB.Password, u.Password) || err != nil {
		h.logger.Infof("Can't authorize with password: %v: error: %v", u.Password, err)
		w.Header().Set("Content-Type", "application/json")
		json := fmt.Sprintf("{\"error\" : incorrect email or password}")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(json))
	} else {
		token, err := generateToken(u.Email + u.Password)
		if err != nil {
			h.logger.Errorf("Can't create new token: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		bearer := "Bearer " + token
		w.Header().Add("Authorization", bearer)
		w.WriteHeader(http.StatusOK)

		s, err := session.New(token, fromDB.ID)
		if err != nil {
			h.logger.Errorf("Can't create struct for session: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.sessionStorage.Create(s)
		if err != nil {
			h.logger.Errorf("Can't create session: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func generateToken(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "can't generate token")
	}

	hasher := md5.New()
	hasher.Write(hash)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func isMatch(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {
		return false
	}

	return true
}


func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var u user.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		h.logger.Errorf("Can't unmarshal input json for update user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := IDFromParams(r)
	if err != nil {
		h.logger.Errorf("Can't get ID from URL params: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := tokenFromReq(r)

	s, err := h.sessionStorage.Find(id)
	if err != nil || token != s.SessionID {
		if err != nil {
			h.logger.Errorf("Don't find session by user ID: %v; because of error: %v", id, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		fromDB, err := h.userStorage.Find(u.Email)
		if err != nil {
			h.logger.Errorf("Can't find user: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if fromDB != nil && id != fromDB.ID {
			h.logger.Errorf("New user info has dublicate email: %v", err)
			w.WriteHeader(http.StatusConflict)
			return
		}

		t := user.JSONTime{time.Now()}
		u.ID = id
		u.UpdatedAt = t
		u.Password, err = generateHash(u.Password)
		if err != nil {
			h.logger.Errorf("Can't generate hash: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.userStorage.Update(&u)
		if err != nil {
			h.logger.Errorf("Can't update user with id=%v: %v", err, id)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u.Password = ""
		w.Header().Set("Content-Type", "application/json")
		json, err := json.Marshal(u)
		if err != nil {
			h.logger.Errorf("Can't marshal user struct: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
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
	const TokenId = 1

	token := r.Header.Get("Authorization")
	s := strings.Split(token, " ")

	return s[TokenId]
}

//
//func getUser(w http.ResponseWriter, r *http.Request) {
//	id, err := IDFromParams(r)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	tokenFromReq := tokenFromReq(r)
//
//	s, ok := session.GetSession(id)
//	if ok && tokenFromReq == s.SessionID {
//		db := db.GetDBConn()
//
//		var u user.User
//		db.Where("id = ?", id).First(&u)
//
//		info := getUserInfo(&u)
//
//		json, err := json.Marshal(info)
//		if err != nil {
//			fmt.Printf("can't marshal struct with user's info: %v", err)
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(http.StatusOK)
//		w.Write(json)
//	} else {
//		w.WriteHeader(http.StatusUnauthorized)
//	}
//
//}
//
//func getUserInfo(u *user.User) user.Info {
//	return user.Info{u.FirstName, u.LastName, u.Birthday, u.Email}
//}
