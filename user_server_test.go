package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

var testServer *httptest.Server
var uServer *userServer

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	uServer = newUserServer("test.db")
	testServer = httptest.NewServer(uServer.mux)
}

func shutdown() {
	uServer.dropDB()
	defer testServer.Close()
}

func TestUserServer_CreateUser(t *testing.T) {

	createdID := createUserAndGetId("{\"name\":\"test_user\",\"age\":\"11\",\"friends\":[]}", t)
	t.Log("created user ID", createdID)
}

func TestUserServer_UpdateAge(t *testing.T) {

	createdID := createUserAndGetId("{\"name\":\"test_update_user\",\"age\":\"11\",\"friends\":[]}", t)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", testServer.URL, createdID), bytes.NewReader([]byte("{\"new_age\":\"100\"}")))
	req.Header.Set("Content-Type", "application/json")
	handleError(t, err)
	client := &http.Client{}
	resp, err := client.Do(req)
	handleError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Expected 200 response")
	}
	respBody, err := io.ReadAll(resp.Body)
	handleError(t, err)
	if string(respBody) != "Возраст пользователя успешно обновлён" {
		t.Error("Unexpected response")
	}
}

func TestUserServer_MakeFriends(t *testing.T) {
	firstID, secondID, thirdID := createThreeFriends(t)

	resp, err := http.Get(fmt.Sprintf("%s/%d", testServer.URL, firstID))
	handleError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err := io.ReadAll(resp.Body)
	responseStr := string(respBody)
	if !strings.Contains(responseStr, strconv.FormatInt(secondID, 10)) || !strings.Contains(responseStr, strconv.FormatInt(thirdID, 10)) {
		t.Fatal("User 1 does not contain friends", responseStr)
	}

	resp, err = http.Get(fmt.Sprintf("%s/friends/%d", testServer.URL, firstID))
	handleError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err = io.ReadAll(resp.Body)
	responseStr = string(respBody)
	if !strings.Contains(responseStr, strconv.FormatInt(secondID, 10)) || !strings.Contains(responseStr, strconv.FormatInt(thirdID, 10)) {
		t.Fatal("User 1 does not contain friends", responseStr)
	}
}

func TestUserServer_DeleteUser(t *testing.T) {
	firstID, secondID, thirdID := createThreeFriends(t)

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/user", testServer.URL),
		bytes.NewReader([]byte(fmt.Sprintf("{\"target_id\":\"%d\"}", firstID))))
	req.Header.Set("Content-Type", "application/json")
	handleError(t, err)
	client := &http.Client{}
	resp, err := client.Do(req)
	handleError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}

	resp, err = http.Get(fmt.Sprintf("%s/friends/%d", testServer.URL, secondID))
	handleError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err := io.ReadAll(resp.Body)
	responseStr := string(respBody)
	if strings.Contains(responseStr, strconv.FormatInt(firstID, 10)) || strings.Contains(responseStr, strconv.FormatInt(thirdID, 10)) {
		t.Fatal("User 2 contains friends", responseStr)
	}

	resp, err = http.Get(fmt.Sprintf("%s/friends/%d", testServer.URL, thirdID))
	handleError(t, err)
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err = io.ReadAll(resp.Body)
	responseStr = string(respBody)
	if strings.Contains(responseStr, strconv.FormatInt(firstID, 10)) || strings.Contains(responseStr, strconv.FormatInt(secondID, 10)) {
		t.Fatal("User 3 contains friends", responseStr)
	}

}

func createUserAndGetId(userStr string, t *testing.T) int64 {
	resp, err := http.Post(fmt.Sprintf("%s/create", testServer.URL), "application/json", bytes.NewReader([]byte(userStr)))
	handleError(t, err)

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("status code is not 201. actual code is ", resp.StatusCode)
	}

	createdIDBytes, err := io.ReadAll(resp.Body)
	handleError(t, err)

	createdID, err := strconv.ParseInt(string(createdIDBytes), 10, 64)
	handleError(t, err)
	return createdID
}

func createThreeFriends(t *testing.T) (int64, int64, int64) {
	firstID := createUserAndGetId("{\"name\":\"test_user_1\",\"age\":\"11\",\"friends\":[]}", t)
	secondID := createUserAndGetId("{\"name\":\"test_user_2\",\"age\":\"22\",\"friends\":[]}", t)
	thirdID := createUserAndGetId("{\"name\":\"test_user_3\",\"age\":\"33\",\"friends\":[]}", t)

	resp, err := http.Post(fmt.Sprintf("%s/make_friends", testServer.URL), "application/json", bytes.NewReader([]byte(fmt.Sprintf("{\"source_id\":\"%d\",\"target_id\":\"%d\"}", firstID, secondID))))
	handleError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err := io.ReadAll(resp.Body)
	handleError(t, err)
	if string(respBody) != "test_user_1 и test_user_2 теперь друзья" {
		t.Error("Unexpected response")
	}

	resp, err = http.Post(fmt.Sprintf("%s/make_friends", testServer.URL), "application/json", bytes.NewReader([]byte(fmt.Sprintf("{\"source_id\":\"%d\",\"target_id\":\"%d\"}", firstID, thirdID))))
	handleError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Error("Expected 200 response")
	}
	respBody, err = io.ReadAll(resp.Body)
	handleError(t, err)
	if string(respBody) != "test_user_1 и test_user_3 теперь друзья" {
		t.Error("Unexpected response")
	}
	return firstID, secondID, thirdID
}

func handleError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
