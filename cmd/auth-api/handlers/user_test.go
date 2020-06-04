package handler

import (
	"bytes"
	"cw1/cmd/socket"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"cw1/pkg/log/logger"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockUserStorage struct {
	u *user.User
	user.Storage
}

func (m mockUserStorage) Create(u *user.User) error {
	return nil
}

func (m mockUserStorage) FindByEmail(email string) (*user.User, error) {
	return m.u, nil
}

func (m mockUserStorage) FindByID(id int64) (*user.User, error) {
	return m.u, nil
}

func (m mockUserStorage) Update(u *user.User) error {
	return nil
}

type mockRobotStorage struct {
	r []*robot.Robot
	robot.Storage
}

func (m mockRobotStorage) Create(r *robot.Robot) error {
	return nil
}

func (m mockRobotStorage) FindByID(id int64) (*robot.Robot, error) {
	return nil, nil
}

func (m mockRobotStorage) FindByOwnerID(id int64) ([]*robot.Robot, error) {
	return m.r, nil
}

func (m mockRobotStorage) FindByTicker(ticker string) ([]*robot.Robot, error) {
	return nil, nil
}

func (m mockRobotStorage) GetAll(id int64, ticker string) ([]*robot.Robot, error) {
	return nil, nil
}

func (m mockRobotStorage) Update(r *robot.Robot) error {
	return nil
}

func (m mockRobotStorage) GetActiveRobots() ([]*robot.Robot, error) {
	return nil, nil
}

type mockSessionStorage struct {
	s *session.Session
	session.Storage
}

func (m mockSessionStorage) Create(session *session.Session) error {
	return nil
}

func (m mockSessionStorage) FindByID(id int64) (*session.Session, error) {
	return m.s, nil
}

func (m mockSessionStorage) FindByToken(token string) (*session.Session, error) {
	return nil, nil
}

type mockLogger struct {
	logger.Logger
}

func (m mockLogger) Debugf(format string, args ...interface{}) {}
func (m mockLogger) Infof(format string, args ...interface{})  {}
func (m mockLogger) Warnf(format string, args ...interface{})  {}
func (m mockLogger) Errorf(format string, args ...interface{}) {}
func (m mockLogger) Fatalf(format string, args ...interface{}) {}
func (m mockLogger) Panicf(format string, args ...interface{}) {}

func respContains(in string, want string) bool {
	if in == "" {
		return want == ""
	}

	return strings.Contains(in, want)
}

func TestSignUpCorrect(t *testing.T) {
	json := []byte(`{"first_name" : "name","last_name": "last_name","birthday": "1970-01-01","email": "email","password":"123456"}`)
	req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	u := &user.User{
		ID: 0, // not find user in storage
	}

	mockUserStorage.u = u

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.signUp)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("signUp handler returned wrong status code: got %v, want %v",
			status, http.StatusCreated)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("signUp handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestSignUpIfUserAlreadyRegistered(t *testing.T) {
	json := []byte(`{"first_name" : "name","last_name": "last_name","birthday": "1970-01-01","email": "email","password":"123456"}`)
	req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	u := &user.User{
		ID:    1, // not zero value => find user in storage
		Email: "email",
	}

	mockUserStorage.u = u

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.signUp)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("signUp handler returned wrong status code: got %v, want %v",
			status, http.StatusConflict)
	}

	expected := fmt.Sprintf("{\"error\":\"user %v is already registered\"}", u.Email)
	if rr.Body.String() != expected {
		t.Errorf("signUp handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestSignUpIncorrectJson(t *testing.T) {
	//body contains incorrect json(missing a open bracket)
	json := []byte(`"first_name" : "name","last_name": "last_name","birthday": "1970-01-01","email": "email","password":"123456"}`)
	req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	u := &user.User{
		ID:    1, // not zero value => find user in storage
		Email: "email",
	}

	mockUserStorage.u = u

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.signUp)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("signUp handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("signUp handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestSignInCorrect(t *testing.T) {
	json := []byte(`{"email": "email","password":"123456"}`)
	req, err := http.NewRequest("POST", "/signin", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	mockUserStorage.u = u

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.signIn)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("signIn handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := "bearer"
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("signIn handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestSignInEmailDontEqual(t *testing.T) {
	json := []byte(`{"email": "EMAIL","password":"123456"}`)
	req, err := http.NewRequest("POST", "/signin", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	mockUserStorage.u = u

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.signIn)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("signIn handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := "incorrect email or password"
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("signIn handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestUpdateUserCorrect(t *testing.T) {
	json := []byte(`{"first_name" : "changed","last_name": "changed","birthday": "2000-01-01","email": "NEWEMAIL","password":"123"}`)
	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.updateUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("updateUser handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"first_name":"changed","last_name":"changed","birthday":"2000-01-01","email":"NEWEMAIL"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("updateUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestUpdateUserIncorrectID(t *testing.T) {
	json := []byte(`{}`)
	req, err := http.NewRequest("PUT", "/users/-1", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.updateUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("updateUser handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := fmt.Sprintf("incorrect id: %v", -1)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("updateUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestUpdateUserNotFind(t *testing.T) {
	json := []byte(`{"first_name" : "changed","last_name": "changed","birthday": "2000-01-01","email": "NEWEMAIL","password":"123"}`)
	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(json))
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token + "changeToken",
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.updateUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("updateUser handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := fmt.Sprintf("{\"error\":\"user %v don't find\"}", u.ID)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("updateUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserCorrect(t *testing.T) {
	req, err := http.NewRequest("PUT", "/users/1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	str := "1970-01-01"

	tt, _ := time.Parse(format.DateLayout, str)
	day := format.Day{V: sql.NullTime{tt, true}}

	u := &user.User{
		ID:        1, // not zero value => find user in storage
		Email:     "email",
		FirstName: "name",
		LastName:  "lastname",
		Birthday:  &day,
		Password:  hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getUser handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{"first_name":"%v","last_name":"%v","birthday":"%v","email":"%v"}`,
		u.FirstName, u.LastName, str, u.Email)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserIncorrectID(t *testing.T) {
	req, err := http.NewRequest("PUT", "/users/-1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	u := &user.User{
		ID:    1, // not zero value => find user in storage
		Email: "email",
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("getUser handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := fmt.Sprintf("incorrect id: %v", -1)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserNotFound(t *testing.T) {
	req, err := http.NewRequest("PUT", "/users/1", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token + "changedToken",
		UserID:    1,
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUser)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("getUser handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := fmt.Sprintf("don't find user with ID %v", u.ID)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUser handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserRobotsCorrect(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/1/robots", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1, // not zero value => find user in storage
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 1, OwnerUserID: 1},
		{RobotID: 2, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.r = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUserRobots)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getUserRobots handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `[{"robot_id":1,"owner_user_id":1,"is_favourite":false,"is_active":false},` +
		`{"robot_id":2,"owner_user_id":1,"is_favourite":false,"is_active":false}]`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUserRobots handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserRobotsUserNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/1/robots", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1,
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token + "changedToken",
		UserID:    0, // zero value => not found session in storage
	}

	rbts := []*robot.Robot{
		{RobotID: 1, OwnerUserID: 1},
		{RobotID: 2, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.r = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUserRobots)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("getUserRobots handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := fmt.Sprintf("can't find user with id: %v", u.ID)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUserRobots handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetUserRobotsCheckTokens(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/1/robots", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}

	token := "8b5d7c0b629267f197f0b5d77c6c066c86e9f9fbd51e3d152cfed360bbf5f"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	l := new(mockLogger)
	hub := socket.NewHub()
	mockUserStorage := new(mockUserStorage)
	mockRobotStorage := new(mockRobotStorage)
	mockSessionStorage := new(mockSessionStorage)

	hash, err := generateHash("123456")
	if err != nil {
		t.Fatalf("can't generate hash %v", err)
	}

	u := &user.User{
		ID:       1,
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token + "changedToken",
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 1, OwnerUserID: 1},
		{RobotID: 2, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.r = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(h.getUserRobots)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("getUserRobots handler returned wrong status code: got %v, want %v",
			status, http.StatusBadRequest)
	}

	expected := `{"error":"tokens don't match"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getUserRobots handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}
