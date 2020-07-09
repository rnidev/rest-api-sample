package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	v1 "github.com/rnidev/go-rest/pkg/service/v1"
)

type App struct {
	pool   *redis.Pool
	Router *mux.Router
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

var (
	ErrIDRequired    = errors.New("id is required")
	ErrInvalidUserID = errors.New("invalid userID")
)

func (app *App) Initialize(redisURL, redisPassword string) {
	app.pool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisURL, redis.DialPassword(redisPassword))
		},
	}
	app.Router = mux.NewRouter()
	app.setRoutes()
}

func (app *App) loadInitData(conn redis.Conn) error {
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

func (app *App) setRoutes() {
	app.Router.HandleFunc("/", app.rootHandler)
	app.Router.StrictSlash(true).PathPrefix("/users").HandlerFunc(app.getUsers).Methods("GET")
	app.Router.StrictSlash(true).PathPrefix("/users").HandlerFunc(app.createOrUpdateUser).Methods("POST")
	app.Router.StrictSlash(true).PathPrefix("/user/{id:[0-9]+}").HandlerFunc(app.getUserByID).Methods("GET")
}

func (app *App) startServer(port string) {
	fmt.Printf("Listening on port :%s", port)
	http.ListenAndServe(":"+port, app.Router)
}

func (app *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello World</h1><div>My name is Rui Ni</div>")
}
func (app *App) createOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	var userData v1.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		renderJSONErrorResp(w, http.StatusBadRequest, err)
		return
	}
	userData.ID = user.ID
	userData.Name = user.Name
	userData.Age = user.Age
	userData.City = user.City
	conn := app.pool.Get()
	err = v1.CreateOrUpdateUser(conn, &userData)
	if err == v1.ErrNoUserFound {
		renderJSONErrorResp(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		renderJSONErrorResp(w, http.StatusInternalServerError, err)
		return
	}
	if user.ID > 0 {
		renderJSONResp(w, http.StatusOK, map[string]string{"message": "user updated successfully"})
	} else {
		renderJSONResp(w, http.StatusCreated, map[string]string{"message": "user created successfully"})
	}
}

func (app *App) getUsers(w http.ResponseWriter, r *http.Request) {
	conn := app.pool.Get()
	usersData, err := v1.ListAllUsers(conn)
	if err != nil {
		renderJSONErrorResp(w, http.StatusInternalServerError, err)
		return
	}
	var users []User
	for _, userData := range usersData {
		var user User
		user.ID = userData.ID
		user.Name = userData.Name
		user.Age = userData.Age
		user.City = userData.City
		users = append(users, user)
	}
	renderJSONResp(w, http.StatusOK, users)
}

func (app *App) getUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		renderJSONErrorResp(w, http.StatusBadRequest, ErrInvalidUserID)
		return
	}
	conn := app.pool.Get()
	userData, err := v1.FindUserByID(conn, userID)
	if err == v1.ErrNoUserFound {
		renderJSONErrorResp(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		renderJSONErrorResp(w, http.StatusInternalServerError, err)
		return
	}
	var user User
	user.ID = userData.ID
	user.Name = userData.Name
	user.Age = userData.Age
	user.City = userData.City
	renderJSONResp(w, http.StatusOK, user)
}

func renderJSONResp(w http.ResponseWriter, httpStatus int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(response)
}

func renderJSONErrorResp(w http.ResponseWriter, httpStatus int, err error) {
	response, _ := json.Marshal(map[string]string{"error": err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(response)
}

func main() {
	app := &App{}
	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("PORT")

	app.Initialize(redisURL, redisPassword)
	app.loadInitData(app.pool.Get())
	app.startServer(port)
}
