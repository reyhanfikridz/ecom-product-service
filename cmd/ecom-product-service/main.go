/*
Package main the executeable file
*/
package main

import (
	"log"

	"github.com/reyhanfikridz/ecom-product-service/api"
	"github.com/reyhanfikridz/ecom-product-service/internal/config"
)

// main
func main() {
	// init API
	a, err := InitAPI()
	if err != nil {
		log.Fatal(err)
	}

	// serve server
	log.Fatal(a.FiberApp.Listen(config.ServerURL))
}

// InitAPI initialize API
func InitAPI() (api.API, error) {
	a := api.API{}

	// init all config before can be used
	err := config.InitConfig()
	if err != nil {
		return a, err
	}

	// init database
	DBConfig := map[string]string{
		"user":     config.DBUsername,
		"password": config.DBPassword,
		"dbname":   config.DBName,
	}
	err = a.InitDB(DBConfig)
	if err != nil {
		return a, err
	}

	// init router
	a.InitRouter()

	return a, nil
}
