/*
Package main the executeable file
*/
package main

import "testing"

// TestInitAPI test InitAPI
func TestInitAPI(t *testing.T) {
	_, err := InitAPI()
	if err != nil {
		t.Errorf("There's an error when initialize API => " + err.Error())
	}
}
