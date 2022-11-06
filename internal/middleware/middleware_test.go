/*
Package middleware collection of middleware used for API
*/
package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// TestGetTokenFromHeader test GetTokenFromHeader
func TestGetTokenFromHeader(t *testing.T) {
	validToken := "This is valid token"
	bearer := "Bearer " + validToken

	// test function get token
	token := GetTokenFromHeader(map[string]string{
		"Authorization": bearer,
	})
	if validToken != token {
		t.Errorf("Expected token " + validToken + ", but got token " + token)
	}
}

// TestGetUserDataFromAuthorizationResp test GetUserFromAuthorizationResp
func TestGetUserDataFromAuthorizationResp(t *testing.T) {
	expectedU := User{
		ID:          1,
		Email:       "buyer@gmail.com",
		FullName:    "buyer",
		Address:     "address",
		PhoneNumber: "08111111111",
		Role:        "buyer",
	}

	// create testing body response
	body, err := json.Marshal(expectedU)
	if err != nil {
		t.Errorf("There's an error when marshal user data to json => %s",
			err.Error())
	}

	// create testing response
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(string(body))),
	}

	// get user data
	u, err := GetUserFromAuthorizationResp(resp)
	if err != nil {
		t.Errorf("Expected error nil, but got error => %s", err.Error())
	}
	if expectedU.ID != u.ID {
		t.Errorf("Expected ID %d, but got %d", expectedU.ID, u.ID)
	}
	if expectedU.Email != u.Email {
		t.Errorf("Expected Email %s, but got %s", expectedU.Email, u.Email)
	}
	if expectedU.FullName != u.FullName {
		t.Errorf("Expected FullName %s, but got %s", expectedU.FullName, u.FullName)
	}
	if expectedU.Address != u.Address {
		t.Errorf("Expected Address %s, but got %s", expectedU.Address, u.Address)
	}
	if expectedU.PhoneNumber != u.PhoneNumber {
		t.Errorf("Expected PhoneNumber %s, but got %s", expectedU.PhoneNumber, u.PhoneNumber)
	}
	if expectedU.Role != u.Role {
		t.Errorf("Expected Role %s, but got %s", expectedU.Role, u.Role)
	}
}
