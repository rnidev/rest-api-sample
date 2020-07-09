package v1

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
)

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

func TestListAllUsersSuccess(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	users, err := ListAllUsers(conn)
	var expectErr error
	resp, _ := json.Marshal(users)
	expectResp := `[{"ID":1,"Name":"John","Age":31,"City":"New York"},{"ID":2,"Name":"Doe","Age":22,"City":"Vancouver"}]`
	if err != expectErr {
		t.Errorf("error: got %s, expected %s", err.Error(), expectErr.Error())
	}

	if string(resp) != expectResp {
		t.Errorf("ListAllUsers() = %s, expect %s", string(resp), expectResp)
	}
}

func TestListAllUsersNotFound(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())

	users, err := ListAllUsers(conn)
	var expectErr error

	if err != expectErr {
		t.Errorf("error: got %s, expected %s", err.Error(), expectErr.Error())
	}

	if users != nil {
		t.Errorf("ListAllUsers() = %v, expect null", users)
	}
}

func TestFindUserByIDSuccess(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	user, err := FindUserByID(conn, 1)
	var expectErr error
	resp, _ := json.Marshal(user)
	expectResp := `{"ID":1,"Name":"John","Age":31,"City":"New York"}`
	if err != expectErr {
		t.Errorf("error: got %s, expected %s", err.Error(), expectErr.Error())
	}

	if string(resp) != expectResp {
		t.Errorf("FindUserByID() = %s, expect %s", string(resp), expectResp)
	}
}

func TestFindUserByIDNotFound(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindUserByID(conn, 4)
	expectErr := ErrNoUserFound
	if err != expectErr {
		t.Errorf("error: got %s, expected %s", err.Error(), expectErr.Error())
	}

}

func TestFindUserByIDInvalidValues(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Do("SET", "user:1", "1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = FindUserByID(conn, 1)
	expectErr := errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	if err.Error() != expectErr.Error() {
		t.Errorf("error: got %s, expected %s", err.Error(), expectErr.Error())
	}

}

func TestCreateUserSuccess(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	user := &User{
		Name: "Doe",
		Age:  33,
		City: "Vancouver",
	}
	err = CreateOrUpdateUser(conn, user)
	if err != nil {
		t.Errorf("error: got %s, expected no error", err.Error())
	}
}

func TestCreateUsersSuccess(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	user1 := &User{Name: "Doe", Age: 33, City: "Vancouver"}
	err = CreateOrUpdateUser(conn, user1)
	if err != nil {
		t.Errorf("error: got %s, expected no error", err.Error())
	}
	user2 := &User{Name: "Doe", Age: 33, City: "Vancouver"}
	err = CreateOrUpdateUser(conn, user2)
	if err != nil {
		t.Errorf("error: got %s, expected no error", err.Error())
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}
	user := &User{
		ID:   1,
		Name: "Doe",
		Age:  33,
		City: "Vancouver",
	}
	err = CreateOrUpdateUser(conn, user)
	if err != nil {
		t.Errorf("error: got %s, expected no error", err.Error())
	}
}

func TestCreateOrUpdateUserNotFound(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	conn, err := redis.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	err = loadInitUserData(conn)
	if err != nil {
		t.Fatal(err)
	}
	user := &User{
		ID:   4,
		Name: "Doe",
		Age:  33,
		City: "Vancouver",
	}
	err = CreateOrUpdateUser(conn, user)
	if err != ErrNoUserFound {
		t.Errorf("error: got %s, expected %s", err.Error(), ErrNoUserFound)
	}
}
