/*
Package config collection of configuration
*/
package config

import (
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

var (
	DBName             string
	DBNameForAPITest   string
	DBNameForModelTest string
	DBUsername         string
	DBPassword         string

	JWTSecretKey     string
	JWTSigningMethod *jwt.SigningMethodHMAC

	FrontendURL       string
	AccountServiceURL string

	MediaFolder string
)

// InitConfig initialize all config variable from environment variable
func InitConfig() error {
	// load all values from .env file into the system
	// .env file must be at root directory (same level as go.mod file)
	err := godotenv.Load(os.ExpandEnv(
		"$GOPATH/src/github.com/reyhanfikridz/ecom-product-service/.env"))
	if err != nil {
		return err
	}

	// set all config variable after all environment variable loaded
	DBName = os.Getenv("ECOM_PRODUCT_SERVICE_DB_NAME")
	DBNameForAPITest = os.Getenv("ECOM_PRODUCT_SERVICE_DB_NAME_FOR_API_TEST")
	DBNameForModelTest = os.Getenv("ECOM_PRODUCT_SERVICE_DB_NAME_FOR_MODEL_TEST")
	DBUsername = os.Getenv("ECOM_PRODUCT_SERVICE_DB_USERNAME")
	DBPassword = os.Getenv("ECOM_PRODUCT_SERVICE_DB_PASSWORD")

	JWTSecretKey = os.Getenv("ECOM_PRODUCT_SERVICE_JWT_SECRET_KEY")
	JWTSigningMethod = jwt.SigningMethodHS256

	FrontendURL = os.Getenv("ECOM_PRODUCT_SERVICE_FRONTEND_URL")
	AccountServiceURL = os.Getenv("ECOM_PRODUCT_SERVICE_ACCOUNT_SERVICE_URL")

	MediaFolder = "/media/"

	return nil
}
