package handler

import (
	"bytes"
	"cw1/cmd/socket"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateRobotCorrect(t *testing.T) {
	json := []byte(`{"owner_user_id": 1,"is_favourite": true,"is_active": true}`)
	req, err := http.NewRequest("POST", "/robot", bytes.NewBuffer(json))
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token + "changedToken",
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 1, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.createRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("createRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusCreated)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("createRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestDeleteRobotCorrect(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/robot/5", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.deleteRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("deleteRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("deleteRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestDeleteRobotNotFound(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/robot/5", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	time, _ := format.NewNullTime()

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1, DeletedAt: time}, // robot is deleted
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.deleteRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("deleteRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"robot with id 5 don't exist"}`
	if rr.Body.String() != expected {
		t.Errorf("deleteRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetRobotsCorrect(t *testing.T) {
	req, err := http.NewRequest("GET", "/robots", nil)
	if err != nil {
		t.Fatalf("can't create request %v", err)
	}
	ticker := "AAPL"

	q := req.URL.Query()
	q.Add("ticker", ticker)
	q.Add("id", "5")

	tick := &format.NullString{V: sql.NullString{ticker, true}}

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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1, Ticker: tick},
		{RobotID: 6, OwnerUserID: 1, Ticker: tick},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.getRobots)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getRobots handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `[{"robot_id":5,"owner_user_id":1,"is_favourite":false,"is_active":false,"ticker":"AAPL"},` +
		`{"robot_id":6,"owner_user_id":1,"is_favourite":false,"is_active":false,"ticker":"AAPL"}]`
	if rr.Body.String() != expected {
		t.Errorf("getRobots handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestMakeFavouriteCorrect(t *testing.T) {
	req, err := http.NewRequest("PUT", "/robot/5/favourite", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1, IsFavourite: false},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.makeFavourite)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("makeFavourite handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"robot_id":5,"owner_user_id":1,"parent_robot_id":5,"is_favourite":true,"is_active":false}`
	if rr.Body.String() != expected {
		t.Errorf("makeFavourite handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestActivateCorrect(t *testing.T) {
	req, err := http.NewRequest("PUT", "/robot/5/activate", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	now, _ := format.NewNullTime()
	start := &format.NullTime{sql.NullTime{now.V.Time.Add(-2 * time.Hour), true}}
	end := &format.NullTime{sql.NullTime{now.V.Time.Add(2 * time.Hour), true}}
	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1, IsActive: false, PlanStart: start, PlanEnd: end},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.activate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("activate handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	// don't check time here
	expected := fmt.Sprintf(`"robot_id":5,"owner_user_id":1,"is_favourite":false,"is_active":true`)
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("activate handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetRobot(t *testing.T) {
	req, err := http.NewRequest("GET", "/robot/5", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.getRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("getRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"robot_id":5,"owner_user_id":1,"is_favourite":false,"is_active":false}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestGetRobotNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/robot/5", nil)
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	time, _ := format.NewNullTime()

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1, DeletedAt: time}, // robot is deleted
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.getRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("getRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusNotFound)
	}

	expected := `{"error":"robot with id 5 don't exist"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("getRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}

func TestUpdateRobotCorrect(t *testing.T) {
	json := []byte(`{"owner_user_id": 1,"is_favourite": true,"is_active": true,"parent_robot_id": 1,` +
		`"ticker": "AAPL","buy_price": 56.5,"sell_price": 46.78,"plan_start": "2002-10-02T15:00:00.05Z","plan_end": "2002-10-02T19:00:00.05Z",` +
		`"plan_yield": 1000,"fact_yield": 100,"deals_count": 10,"activated_at": "2002-10-02T15:00:00.05Z",` +
		`"deactivated_at": "2002-10-02T19:00:00.05Z","created_at": "2000-10-02T19:00:00.05Z"}`)
	req, err := http.NewRequest("PUT", "/robot/5", bytes.NewBuffer(json))
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
		ID:       1, // not zero value
		Email:    "email",
		Password: hash,
	}

	s := &session.Session{
		SessionID: token,
		UserID:    1,
	}

	rbts := []*robot.Robot{
		{RobotID: 5, OwnerUserID: 1},
	}

	mockUserStorage.u = u
	mockSessionStorage.s = s
	mockRobotStorage.rr = rbts

	h, _ := New(l, mockUserStorage, mockSessionStorage, mockRobotStorage, hub)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(h.updateRobot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("updateRobot handler returned wrong status code: got %v, want %v",
			status, http.StatusOK)
	}

	expected := `{"robot_id":5,"owner_user_id":1,"parent_robot_id":1,"is_favourite":true,"is_active":true,"ticker":"AAPL",` +
		`"buy_price":56.5,"sell_price":46.78,"plan_start":"2002-10-02T15:00:00Z","plan_end":"2002-10-02T19:00:00Z",` +
		`"plan_yield":1000,"fact_yield":100,"deals_count":10,"activated_at":"2002-10-02T15:00:00Z",` +
		`"deactivated_at":"2002-10-02T19:00:00Z","created_at":"2000-10-02T19:00:00Z"}`
	if !respContains(rr.Body.String(), expected) {
		t.Errorf("updateRobot handler returned unexpected body: got %v, want %v",
			rr.Body.String(), expected)
	}
}
