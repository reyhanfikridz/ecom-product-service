/*
Package middleware collection of middleware used for API
*/
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/reyhanfikridz/ecom-product-service/internal/config"
)

// User containing user data after authorization
type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}

// AuthorizationMiddleware authorize each API route by checking JWT Token
func AuthorizationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get token
		token := GetTokenFromHeader(c.GetReqHeaders())
		if token == "" {
			return c.Status(http.StatusForbidden).JSON(map[string]string{
				"message": "Token authorization empty/not found",
			})
		}

		// set form data
		formData := map[string]io.Reader{
			"token": strings.NewReader(token),
		}

		// transform form data to bytes buffer
		var bFormData bytes.Buffer
		bFormDataWriter := multipart.NewWriter(&bFormData)
		for key, formDataReader := range formData {
			fieldWriter, err := bFormDataWriter.CreateFormField(key)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(map[string]string{
					"message": err.Error(),
				})
			}

			_, err = io.Copy(fieldWriter, formDataReader)
			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(map[string]string{
					"message": err.Error(),
				})
			}
		}
		bFormDataWriter.Close()

		// authorize to account service
		resp, err := http.Post(config.AccountServiceURL+"/api/authorize/",
			bFormDataWriter.FormDataContentType(),
			&bFormData)
		if err != nil { // if error occured
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": err.Error(),
			})
		}
		if resp.StatusCode != http.StatusOK { // if unauthorized
			return c.Status(http.StatusForbidden).JSON("Token authorization invalid")
		}

		// get user data from authorization response
		user, err := GetUserFromAuthorizationResp(resp)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": err.Error(),
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}

// GetTokenFromHeader getting token (bearer) from request header
func GetTokenFromHeader(headers map[string]string) string {
	rawToken := headers["Authorization"]
	if rawToken == "" {
		return ""
	}

	splitToken := strings.Split(rawToken, "Bearer ")
	if len(splitToken) <= 1 {
		return ""
	}

	token := splitToken[1]
	return token
}

// GetUserFromAuthorizationResp get user data from authorization response
func GetUserFromAuthorizationResp(resp *http.Response) (User, error) {
	user := User{}

	err := json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}
