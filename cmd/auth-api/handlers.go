package main

import (
	"cw1/internal/db"
	"cw1/internal/user"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"time"
	"golang.org/x/crypto/bcrypt"
)

var (
	r chi.Router
)

func init() {
	r = chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", signUp)
	})
}

func generateHash(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		log.Printf("can't generate hash: %v", err)
	}

	return string(hash)
}

func signUp(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	str := time.Now().Format(time.RFC3339)
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		log.Printf("can't parse current time string: %v", err)
	}
	u.CreatedAt = t
	u.UpdatedAt = t
	u.Password = generateHash(u.Password)

	db := db.GetDBConn()

	if err := db.Create(&u).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		json := fmt.Sprintf("{\"error\" : \"user %s is already registered\"}", u.Email)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(json))
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}
