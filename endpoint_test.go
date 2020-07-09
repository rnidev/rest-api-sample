package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomodule/redigo/redis"

	"github.com/alicebob/miniredis/v2"
)

func setup() *App {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	app := &App{}
	app.Initialize(s.Addr(), "")
	return app
}

func loadInitUserData(conn redis.Conn) error {
	user1 := map[string]string{"id": "1", "name": "John", "age": "31", "city": "New York"}
	user2 := map[string]string{"id": "2", "name": "Doe", "age": "22", "city": "Vancouver"}
	if _, err := conn.Do("HMSET", redis.Args{}.Add("user:1").AddFlat(user1)...); err != nil {
		return err
	}
	if _, err := conn.Do("HMSET", redis.Args{}.Add("user:2").AddFlat(user2)...); err != nil {
		return err
	}
	return nil
}

func TestRootHandler(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.rootHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusOK)
	}
	expected := `<h1>Hello World</h1><div>My name is Rui Ni</div>`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestGetUsersInternalServerError(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	conn.Do("SET", "user:1", "invalid data type")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUsers)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetUsersSuccess(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUsers)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusOK)
	}
	expected := `[{"ID":1,"Name":"John","Age":31,"City":"New York"},{"ID":2,"Name":"Doe","Age":22,"City":"Vancouver"}]`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestCreateUserSuccess(t *testing.T) {
	app := setup()
	requestDataString := []byte(`{"name": "John", "age": 31, "city": "New York"}`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestDataString))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusCreated)
	}
	expected := `{"message":"user created successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	app := setup()
	conn := app.pool.Get()
	err := loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}
	requestDataString := []byte(`{"id": 1, "name": "John", "age": 31, "city": "New York"}`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestDataString))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusOK)
	}
	expected := `{"message":"user updated successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestCreateOrUpdateUserBadRequest(t *testing.T) {
	app := setup()
	conn := app.pool.Get()
	err := loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}
	requestDataString := []byte(`{"id": 1, "name": "John", "age": 31, "city": "New Y`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestDataString))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateOrUpdateUserInternalServerError(t *testing.T) {
	app := setup()
	conn := app.pool.Get()
	conn.Do("SET", "userIncrID", "safdkj")
	requestDataString := []byte(`{"name": "John", "age": 31, "city": "New York"}`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestDataString))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	app := setup()
	requestDataString := []byte(`{"id": 1, "name": "John", "age": 31, "city": "New York"}`)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(requestDataString))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusNotFound)
	}
}

func TestGetUserByIDSuccess(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/1/", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusOK)
	}
	expected := `{"id":1,"name":"John","age":31,"city":"New York"}`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestGetUserByIDMissingID(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusNotFound)
	}
}

func TestGetUserByIDNonIntegerID(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/asfsadfsa/", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusNotFound)
	}
}

func TestGetUserByIDNoUserFound(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/124/", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusNotFound)
	}
	expected := `{"error":"no user found"}`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}

func TestGetUserByIDInternalServerError(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/1/", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := app.pool.Get()
	if _, err := conn.Do("SET", "user:1", "INVALID DATA"); err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetUserByIDInvalidID(t *testing.T) {
	app := setup()
	req, err := http.NewRequest("GET", "/user/", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("id", "sadfasf")
	req.URL.RawQuery = q.Encode()
	conn := app.pool.Get()
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.getUserByID)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("http status code: got %v, expected %v", rr.Code, http.StatusBadRequest)
	}
	expected := `{"error":"invalid userID"}`
	if rr.Body.String() != expected {
		t.Errorf("response body: got %v, expected %v", rr.Body.String(), expected)
	}
}
