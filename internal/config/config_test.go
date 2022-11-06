/*
Package config collection of configuration
*/
package config

import "testing"

// TestInitConfig test InitConfig
func TestInitConfig(t *testing.T) {
	err := InitConfig()
	if err != nil {
		t.Errorf("Expected initialize config success, but failed => %s",
			err.Error())
	}
}
