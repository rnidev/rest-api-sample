package v1

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type User struct {
	ID   int    `redis:"id"`
	Name string `redis:"name"`
	Age  int    `redis:"age"`
	City string `redis:"city"`
}

var userKeyPrefix = "user:"

var (
	ErrNoUserFound = errors.New("no user found")
)

func ListAllUsers(conn redis.Conn) ([]*User, error) {
	//Fetch all the keys match this pattern "user:[0-9]"
	userIDPattern := userKeyPrefix + "[0-9]"
	keys, err := redis.Strings(conn.Do("KEYS", userIDPattern))
	if err != nil {
		return nil, err
	}
	//Loop through found user keys and fetch each user record
	var users []*User
	for _, key := range keys {
		idString := strings.TrimPrefix(key, userKeyPrefix)
		id, err := strconv.Atoi(idString)
		if err != nil {
			return nil, err
		}
		user, err := FindUserByID(conn, id)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func FindUserByID(conn redis.Conn, userID int) (*User, error) {
	userKey := userKeyPrefix + strconv.Itoa(userID)
	//get all the values stores for this userKey
	values, err := redis.Values(conn.Do("HGETALL", userKey))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNoUserFound
	}
	var user User
	//map stored values to redis user struct
	err = redis.ScanStruct(values, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateOrUpdateUser(conn redis.Conn, user *User) error {
	var (
		id     int
		err    error
		userID string
		exists int
	)

	//check input user.ID to either create or update
	if user.ID > 0 {
		userID = "user:" + strconv.Itoa(user.ID)
		exists, err = redis.Int(conn.Do("EXISTS", userID))
		if exists == 0 {
			return ErrNoUserFound
		}
	} else {
		id, err = getNewUserID(conn)
		if err != nil {
			return err
		}
		user.ID = id
		userID = "user:" + strconv.Itoa(id)
	}

	if _, err = conn.Do("HMSET", redis.Args{}.Add(userID).AddFlat(user)...); err != nil {
		return err
	}
	return nil
}

//getNewUserID is to use userIncrID as an auto increment key for userID
func getNewUserID(conn redis.Conn) (int, error) {
	var (
		key    = "userIncrID"
		id     int
		err    error
		exists int
	)

	exists, err = redis.Int(conn.Do("EXISTS", key))
	if err != nil {
		return 0, err
	}
	//if userIncrID is not set, set to 1 as initial id
	if exists == 0 {
		_, err = redis.String(conn.Do("SET", key, 1))
		if err != nil {
			return 0, err
		}
		return 1, nil
	}

	id, err = redis.Int(conn.Do("INCR", key))
	if err != nil {
		return 0, err
	}
	return id, nil
}
