/*
Package api containing API initialization and API route handler
*/
package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"
	"github.com/reyhanfikridz/ecom-product-service/internal/config"
	"github.com/reyhanfikridz/ecom-product-service/internal/middleware"
	"github.com/reyhanfikridz/ecom-product-service/internal/model"
	"github.com/reyhanfikridz/ecom-product-service/internal/validator"
)

// API contain database connection and router GoFiber for product service API
type API struct {
	DB       *sql.DB
	FiberApp *fiber.App
}

// InitDB initialize API database connection
func (a *API) InitDB(DBConfig map[string]string) error {
	// connect to db
	connString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DBConfig["user"], DBConfig["password"], DBConfig["dbname"])

	var err error
	a.DB, err = sql.Open("postgres", connString)
	if err != nil {
		return err
	}

	// create db tables if not exist
	tableCreationQuery := `
		CREATE TABLE IF NOT EXISTS product_productinfo (
			id SERIAL PRIMARY KEY NOT NULL,
			sku VARCHAR(15) UNIQUE NOT NULL,
			name VARCHAR(100) NOT NULL,
			price NUMERIC NOT NULL,
			weight REAL NOT NULL,
			description TEXT,
			stock INT NOT NULL,
			account_user_id INT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS product_productimage
		(
			id SERIAL PRIMARY KEY NOT NULL,
			image_path VARCHAR(250) NOT NULL,
			product_productinfo_id INT NOT NULL,
			CONSTRAINT fk_product_productinfo
				FOREIGN KEY(product_productinfo_id) 
					REFERENCES product_productinfo(id)
					ON DELETE CASCADE
		);
	`

	_, err = a.DB.Exec(tableCreationQuery)
	if err != nil {
		return err
	}

	return nil
}

// InitRouter initialize GoFiber router for API
func (a *API) InitRouter() {
	a.FiberApp = fiber.New()

	// add middleware CORS and logger to all route
	a.FiberApp.Use(
		cors.New(
			cors.Config{
				AllowOrigins: fmt.Sprintf("%s,%s",
					config.FrontendURL, config.AccountServiceURL),
				AllowHeaders: "Authorization, Origin, Content-Type, Accept",
			},
		),
	)
	a.FiberApp.Use(logger.New())

	// create main router group (prefix: "/api") with middleware authorization
	mainRouter := a.FiberApp.Group("/api", middleware.AuthorizationMiddleware())

	//// route add product
	mainRouter.Post("/product/", a.AddProductHandler)

	//// route get products
	mainRouter.Get("/products/", a.GetProductsHandler)

	//// route get products by user ID
	mainRouter.Get("/products/user/", a.GetProductsByUserIDHandler)

	//// route get product by sku
	mainRouter.Get("/product/", a.GetProductHandler)

	//// route update product by sku
	mainRouter.Put("/product/", a.UpdateProductHandler)

	//// route delete product by sku
	mainRouter.Delete("/product/", a.DeleteProductHandler)

	//// route decrease product stock by sku
	mainRouter.Put("/product/decrease/stock/", a.DecreaseStockHandler)

	// route static media
	a.FiberApp.Static("/media", "./../media")
}

// AddProductHandler handling route add product (method: POST, user: seller)
func (a *API) AddProductHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is seller
	if u.Role != "seller" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// parse product info from form data
	pInfo := model.ProductInfo{}
	err := c.BodyParser(&pInfo)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// validate product info data
	err = validator.IsProductInfoValid(pInfo)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// insert product info into database
	pInfo.UserID = u.ID
	pInfo, err = model.InsertProductInfo(a.DB, pInfo)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// get image form (multi images)
	imageForm, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// insert product images into database and media folder
	fileHeaders := imageForm.File["product_images"]
	if len(fileHeaders) > 0 {
		err = model.InsertProductImages(a.DB, fileHeaders, pInfo)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": fmt.Sprintf("Product info data created successfully, "+
					"but saving product images failed => %s", err.Error()),
			})
		}
	}

	return c.Status(http.StatusCreated).JSON(pInfo)
}

// GetProductsHandler handling route get products (method: GET, user: buyer)
func (a *API) GetProductsHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is buyer
	if u.Role != "buyer" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// get products from database
	products, err := model.GetProducts(a.DB, model.ProductInfo{}, c.Query("search"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": fmt.Sprintf(
				"There's an error when getting the products data => %s",
				err.Error()),
		})
	}

	return c.Status(http.StatusOK).JSON(products)
}

// GetProductsByUserIDHandler handling route get products by user ID
// (method: GET, user: seller)
func (a *API) GetProductsByUserIDHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is seller
	if u.Role != "seller" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// get products by user id from database
	products, err := model.GetProducts(a.DB, model.ProductInfo{UserID: u.ID},
		c.Query("search"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": fmt.Sprintf(
				"There's an error when getting the products data => %s",
				err.Error()),
		})
	}

	return c.Status(http.StatusOK).JSON(products)
}

// GetProductHandler handling route get one product by SKU
// (method: GET, user: all)
func (a *API) GetProductHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	_, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// get SKU from url
	SKU := c.Query("sku")
	if strings.TrimSpace(SKU) == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "parameter 'sku' empty/not found",
		})
	}

	// get product by sku from database
	p, err := model.GetProductBySKU(a.DB, SKU)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": fmt.Sprintf(
				"There's an error when getting the product data => %s",
				err.Error()),
		})
	}

	// set seller info with API get user from account service
	if c.Query("testing") != "1" {
		resp, err := http.Get(config.AccountServiceURL + "/api/user/?id=" +
			strconv.Itoa(p.ProductInfo.UserID))
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": fmt.Sprintf(
					"There's an error when getting seller info => %s",
					err.Error()),
			})
		}
		if resp.StatusCode != http.StatusOK {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": fmt.Sprintf(
					"Status code invalid when getting seller info  => %d",
					resp.StatusCode),
			})
		}
		err = json.NewDecoder(resp.Body).Decode(&p.SellerInfo)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": fmt.Sprintf(
					"There's an error when decoding seller info => %s",
					err.Error()),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(p)
}

// UpdateProductHandler handling route update product (method: PUT, user: seller)
func (a *API) UpdateProductHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is seller
	if u.Role != "seller" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// parse product info from form data
	pInfo := model.ProductInfo{}
	err := c.BodyParser(&pInfo)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// validate product info data
	err = validator.IsProductInfoValid(pInfo)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// get SKU from url
	SKU := c.Query("sku")
	if strings.TrimSpace(SKU) == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "parameter 'sku' empty/not found",
		})
	}

	// update product info in database
	pInfo.UserID = u.ID
	pInfo.SKU = SKU
	pInfo, err = model.UpdateProductInfoBySKU(a.DB, pInfo)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// get image form (multi images)
	imageForm, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// update product images in database and media folder
	fileHeaders := imageForm.File["product_images"]
	if len(fileHeaders) > 0 {
		err = model.InsertProductImages(a.DB, fileHeaders, pInfo)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"message": fmt.Sprintf("Product info data updated successfully, "+
					"but update product images failed => %s", err.Error()),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(pInfo)
}

// DeleteProductHandler handling route delete product (method: DELETE, user: seller)
func (a *API) DeleteProductHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is seller
	if u.Role != "seller" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// get SKU from url
	SKU := c.Query("sku")
	if strings.TrimSpace(SKU) == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "parameter 'sku' empty/not found",
		})
	}

	// delete product by SKU in database
	err := model.DeleteProductBySKU(a.DB, SKU)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "Delete product success!",
	})
}

// DecreaseStockHandler handling route decrease product stock (method: PUT, user: seller)
func (a *API) DecreaseStockHandler(c *fiber.Ctx) error {
	// get user data
	tmpU := c.Locals("user")
	u, ok := tmpU.(middleware.User)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": "user data invalid",
		})
	}

	// check user role is seller
	if u.Role != "seller" {
		return c.Status(http.StatusForbidden).JSON(map[string]string{
			"message": "user doesn't have authority to access this API",
		})
	}

	// parse order quantity from form data
	type OrderQty struct {
		Qty int `form:"qty"`
	}
	oQty := OrderQty{}
	err := c.BodyParser(&oQty)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// get SKU from url
	SKU := c.Query("sku")
	if strings.TrimSpace(SKU) == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "parameter 'sku' empty/not found",
		})
	}

	// get product data
	p, err := model.GetProductBySKU(a.DB, SKU)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	// update product stock
	p.ProductInfo.Stock -= oQty.Qty
	_, err = model.UpdateProductInfoBySKU(a.DB, p.ProductInfo)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "Product stock updated!",
	})
}
