package main

import (
	"crypto/md5"
	"cw1/internal/db"
	"cw1/internal/session"
	"cw1/internal/user"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	r chi.Router
)

func init() {
	r = chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", signUp)
		r.Post("/signin", signIn)
		r.Put("/users/{id}", update)
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

	//t, err := getCurrentTime()
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	t := user.JSONTime{time.Now()}

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

func generateToken(s string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	hasher := md5.New()
	hasher.Write(hash)

	return hex.EncodeToString(hasher.Sum(nil))
}

func isMatch(hashedPwd string, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func signIn(w http.ResponseWriter, r *http.Request) {
	var u user.User

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := db.GetDBConn()

	var fromDB user.User

	err = db.Where("email = ?", u.Email).First(&fromDB).Error

	if !isMatch(fromDB.Password, u.Password) || err != nil {
		w.Header().Set("Content-Type", "application/json")
		json := fmt.Sprintf("{\"error\" : incorrect email or password}")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(json))
	} else {
		token := generateToken(u.Email + u.Password)
		bearer := "Bearer " + token
		w.Header().Add("Authorization", bearer)
		w.WriteHeader(http.StatusOK)

		err = session.AddSession(token, fromDB.ID)
		if err != nil {
			fmt.Printf("can't create session: %v", err)
		}
	}
}

func getToken(r *http.Request) string {
	const TokenId = 1

	token := r.Header.Get("Authorization")
	s := strings.Split(token, " ")

	return s[TokenId]
}

func update(w http.ResponseWriter, r *http.Request) {
	var u user.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	str := chi.URLParam(r, "id")
	id, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("can't parse string to int: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenFromReq := getToken(r)

	s, ok := session.GetSession(id)
	if ok && tokenFromReq == s.SessionID {
		db := db.GetDBConn()

		var c int
		db.Where("email = ?", u.Email).Count(c)
		if c != 0 {
			fmt.Printf("new user info has dublicate email: %v", err)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		//t, err := getCurrentTime()
		//if err != nil {
		//	fmt.Printf("can't update time: %v", err)
		//	http.Error(w, err.Error(), http.StatusBadRequest)
		//	return
		//}

		t := user.JSONTime{time.Now()}

		u.ID = id
		u.UpdatedAt = t
		u.Password = generateHash(u.Password)
		fmt.Println(t)
		db.Save(&u)

		db.Where("id = ?", id).First(&u)
		fmt.Println(u.CreatedAt)
		u.Password = ""
		w.Header().Set("Content-Type", "application/json")
		json, err := json.Marshal(u)
		fmt.Println(string(json))
		if err != nil {
			fmt.Printf("can't marshal user struct: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func getCurrentTime() (user.JSONTime, error) {
	str := time.Now().Format("2006-01-02T15:04:05Z")
	t, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		return user.JSONTime{}, fmt.Errorf("can't parse current time string: %v", err)
	}

	return user.JSONTime{t}, nil
}
