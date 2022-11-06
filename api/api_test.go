/*
Package api containing API initialization and API route handler
*/
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/reyhanfikridz/ecom-product-service/internal/config"
	"github.com/reyhanfikridz/ecom-product-service/internal/middleware"
	"github.com/reyhanfikridz/ecom-product-service/internal/model"
)

// TestMain do some test before and after all testing in the package
func TestMain(m *testing.M) {
	// init all config before can be used
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("There's an error when initialize config => %s",
			err.Error())
	}

	// get testing app
	u := middleware.User{}
	a, err := GetTestingAPI(u)
	if err != nil {
		log.Fatalf("There's an error when initialize "+
			"testing API => %s", err.Error())
	}

	// truncate product tables before all test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_product before running all test => %s",
			err.Error())
	}

	// delete contents in folder media-test before all test
	contents, err := filepath.Glob("./../media-test/")
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			log.Fatalf("There's an error when deleting "+
				"images in media-test folder => %s",
				err.Error())
		}
	}

	// run all testing
	m.Run()

	// truncate product tables after all test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_product before running all test => %s",
			err.Error())
	}

	// delete contents in folder media-test after all test
	contents, err = filepath.Glob("./../media-test/")
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			log.Fatalf("There's an error when deleting "+
				"images in media-test folder => %s",
				err.Error())
		}
	}
}

// TestInitDB test InitDB
func TestInitDB(t *testing.T) {
	a := API{}

	DBConfig := map[string]string{
		"user":     config.DBUsername,
		"password": config.DBPassword,
		"dbname":   config.DBName,
	}
	err := a.InitDB(DBConfig)
	if err != nil {
		t.Errorf("Expected database connection success,"+
			" but connection failed => %s", err.Error())
	}
}

// TestAddProductHandler test AddProductHandler
func TestAddProductHandler(t *testing.T) {
	// initialize testing table
	testTable := []struct {
		TestName       string
		FormData       map[string]string
		ProductPrice   float64
		ProductWeight  float32
		ProductStock   int
		User           middleware.User
		ExpectedStatus int
	}{
		{
			TestName: "Test Add Product Success",
			FormData: map[string]string{
				"name":        "Product 1",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Product description",
				"stock":       "100",
			},
			ProductPrice:  1000000.50,
			ProductWeight: 1.5,
			ProductStock:  100,
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			ExpectedStatus: http.StatusCreated,
		},
		{
			TestName: "Test Add Product Forbidden",
			FormData: map[string]string{
				"name":        "Product 1",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Product description",
				"stock":       "100",
			},
			User: middleware.User{
				ID:   1,
				Role: "buyer",
			},
			ExpectedStatus: http.StatusForbidden,
		},
		{
			TestName: "Test Add Product Bad Request 1",
			FormData: map[string]string{
				"name":        "",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Product description",
				"stock":       "100",
			},
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "Test Add Product Bad Request 2",
			FormData: map[string]string{
				"name":        "Product 1",
				"price":       "",
				"weight":      "1.5",
				"description": "Product description",
				"stock":       "100",
			},
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "Test Add Product Bad Request 3",
			FormData: map[string]string{
				"name":        "Product 1",
				"price":       "1000000.50",
				"weight":      "",
				"description": "Product description",
				"stock":       "100",
			},
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// initialize testing API
		a, err := GetTestingAPI(test.User)
		if err != nil {
			t.Errorf("[%s] There's an error when getting testing API => %s",
				test.TestName, err.Error())
		}

		// transform form data to bytes buffer
		var bFormData bytes.Buffer
		w := multipart.NewWriter(&bFormData)
		for key, r := range test.FormData {
			fw, err := w.CreateFormField(key)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}

			_, err = io.Copy(fw, strings.NewReader(r))
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}
		}

		// add image files to form
		for i := 1; i <= 10; i++ {
			ffw, err := w.CreateFormFile("product_images", fmt.Sprintf("test_%d.png", i))
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}

			img := CreateTestImage()
			err = png.Encode(ffw, img)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}
		}

		w.Close()

		// create new request
		req, err := http.NewRequest("POST", "/api/product/", &bFormData)
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API create product => %s",
				test.TestName, err.Error())
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		// run request
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response
		if response.StatusCode != test.ExpectedStatus {
			t.Errorf("[%s] Expected status %d got %d",
				test.TestName, test.ExpectedStatus, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)
		} else {
			if test.ExpectedStatus == http.StatusCreated {
				// get response data (product)
				var respPInfo model.ProductInfo
				err = json.NewDecoder(response.Body).Decode(&respPInfo)
				if err != nil {
					t.Errorf("[%s] There's an error when unmarshal body response => %s",
						test.TestName, err.Error())
				}

				// check result
				if respPInfo.ID == 0 {
					t.Errorf("[%s] Expected id not zero, but got zero", test.TestName)
				}
				if test.FormData["name"] != respPInfo.Name {
					t.Errorf("[%s] Expected name '%s', but got name '%s'",
						test.TestName, test.FormData["name"], respPInfo.Name)
				}
				if test.ProductPrice != respPInfo.Price {
					t.Errorf("[%s] Expected price %f, but got price %f",
						test.TestName, test.ProductPrice, respPInfo.Price)
				}
				if test.ProductWeight != respPInfo.Weight {
					t.Errorf("[%s] Expected weight %f, but got weight %f",
						test.TestName, test.ProductWeight, respPInfo.Weight)
				}
				if test.FormData["description"] != respPInfo.Description {
					t.Errorf("[%s] Expected description '%s', but got description '%s'",
						test.TestName, test.FormData["description"], respPInfo.Description)
				}
				if test.ProductStock != respPInfo.Stock {
					t.Errorf("[%s] Expected Stock %d, but got Stock %d",
						test.TestName, test.ProductStock, respPInfo.Stock)
				}
				if test.User.ID != respPInfo.UserID {
					t.Errorf("[%s] Expected user_id %d, but got user_id %d",
						test.TestName, test.User.ID, respPInfo.UserID)
				}

				// check total image inserted/added
				row := a.DB.QueryRow(`
					SELECT 
						COUNT(*)
					FROM product_productimage
					WHERE product_productinfo_id = $1`,
					respPInfo.ID)
				if row.Err() != nil {
					t.Errorf("[%s] There's an error when "+
						"get total image inserted => %s",
						test.TestName, row.Err().Error())
				}

				var totalImage int
				err = row.Scan(&totalImage)
				if err != nil {
					t.Errorf("[%s] There's an error when "+
						"get total image inserted => %s",
						test.TestName, err.Error())
				}

				if totalImage != 10 {
					t.Errorf("[%s] Expected total image 10 inserted, but got %d",
						test.TestName, totalImage)
				}

				// delete data after test create product success
				_, err = a.DB.Exec(`DELETE FROM product_productinfo WHERE sku = $1`,
					respPInfo.SKU)
				if err != nil {
					t.Errorf("[%s] There's an error when deleting previous "+
						"testing data => %s", test.TestName, err.Error())
				}
			}
		}
	}

	// truncate tables after test
	a, err := GetTestingAPI(middleware.User{})
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestGetProductsHandler test GetProductsHandler
func TestGetProductsHandler(t *testing.T) {
	// get testing API for create products
	u := middleware.User{}
	a, err := GetTestingAPI(u)
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	// create products
	sop := []model.Product{
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT A Path 1.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 2.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 3.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      2,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = model.InsertProductInfo(a.DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = a.DB.Exec(`INSERT INTO 
				product_productimage(image_path, product_productinfo_id)
				VALUES($1,$2)`,
				pImage.ImagePath, sop[i].ProductInfo.ID)
			if err != nil {
				t.Errorf("There's an error when insert data product image => %s",
					err.Error())
			}
		}
	}

	// create testing table
	testTable := []struct {
		TestName       string
		Search         string
		User           middleware.User
		ExpectedStatus int
		ExpectedData   []model.Product
	}{
		{
			TestName:       "Get All",
			Search:         "",
			User:           middleware.User{Role: "buyer"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   sop,
		},
		{
			TestName:       "Get By Search <product b>",
			Search:         "product b",
			User:           middleware.User{Role: "buyer"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{sop[1], sop[2]},
		},
		{
			TestName:       "Forbidden",
			Search:         "",
			User:           middleware.User{Role: "seller"},
			ExpectedStatus: http.StatusForbidden,
			ExpectedData:   []model.Product{},
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// get testing API for get products
		a, err = GetTestingAPI(test.User)
		if err != nil {
			t.Errorf("There's an error when getting testing API => %s",
				err.Error())
		}

		// get url params
		params := url.Values{}
		params.Add("search", test.Search)

		// create new request
		req, err := http.NewRequest("GET", "/api/products/", nil)
		req.URL.RawQuery = params.Encode()
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API get products => %s",
				test.TestName, err.Error())
		}

		// run request
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response status
		if response.StatusCode != test.ExpectedStatus {
			t.Errorf("[%s] Expected status %d got %d",
				test.TestName, test.ExpectedStatus, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else if response.StatusCode == test.ExpectedStatus &&
			response.StatusCode != http.StatusForbidden {
			// get response data (product)
			var result []model.Product
			err = json.NewDecoder(response.Body).Decode(&result)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			// check response data length
			if len(test.ExpectedData) != len(result) {
				t.Errorf("[%s] Expected length data %d, but got %d",
					test.TestName, len(test.ExpectedData),
					len(result))
			}

			// check response data
			for _, ep := range test.ExpectedData {
				// check product exist
				pExist := model.Product{}
				for _, p := range result {
					if p.ProductInfo.SKU == ep.ProductInfo.SKU &&
						p.ProductInfo.Name == ep.ProductInfo.Name &&
						p.ProductInfo.Price == ep.ProductInfo.Price &&
						p.ProductInfo.Weight == ep.ProductInfo.Weight &&
						p.ProductInfo.Description == ep.ProductInfo.Description &&
						p.ProductInfo.Stock == ep.ProductInfo.Stock &&
						p.ProductInfo.UserID == ep.ProductInfo.UserID {
						pExist = p
						break
					}
				}

				if pExist.ProductInfo.ID == 0 {
					t.Errorf("[%s] Expected product with SKU %s "+
						"not found", test.TestName, ep.ProductInfo.SKU)
				} else {
					// check all product images exist
					for _, ePImage := range ep.ProductImages {
						ePImageExist := false
						for _, pImage := range pExist.ProductImages {
							if pImage.ImagePath == ePImage.ImagePath {
								ePImageExist = true
								break
							}
						}

						if !ePImageExist {
							t.Errorf("[%s] Expected product with image path %s "+
								"not found", test.TestName, ePImage.ImagePath)
						}
					}
				}
			}
		}
	}

	// truncate tables after test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestGetProductsByUserIDHandler test GetProductsByUserIDHandler
func TestGetProductsByUserIDHandler(t *testing.T) {
	// get testing API for create products
	u := middleware.User{}
	a, err := GetTestingAPI(u)
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	// create products
	sop := []model.Product{
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT A Path 1.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 2.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 3.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      2,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = model.InsertProductInfo(a.DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = a.DB.Exec(`INSERT INTO 
				product_productimage(image_path, product_productinfo_id)
				VALUES($1,$2)`,
				pImage.ImagePath, sop[i].ProductInfo.ID)
			if err != nil {
				t.Errorf("There's an error when insert data product image => %s",
					err.Error())
			}
		}
	}

	// create testing table
	testTable := []struct {
		TestName       string
		Search         string
		User           middleware.User
		ExpectedStatus int
		ExpectedData   []model.Product
	}{
		{
			TestName:       "Get By User ID <1>",
			Search:         "",
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{sop[0], sop[2]},
		},
		{
			TestName:       "Get By User ID <1> and Search <product b>",
			Search:         "product b",
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{sop[2]},
		},
		{
			TestName:       "Forbidden",
			Search:         "",
			User:           middleware.User{ID: 1, Role: "buyer"},
			ExpectedStatus: http.StatusForbidden,
			ExpectedData:   []model.Product{},
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// get testing API for get products
		a, err = GetTestingAPI(test.User)
		if err != nil {
			t.Errorf("There's an error when getting testing API => %s",
				err.Error())
		}

		// get url params
		params := url.Values{}
		params.Add("search", test.Search)

		// create new request
		req, err := http.NewRequest("GET", "/api/products/user/", nil)
		req.URL.RawQuery = params.Encode()
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API get products => %s",
				test.TestName, err.Error())
		}

		// run request
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response status
		if response.StatusCode != test.ExpectedStatus {
			t.Errorf("[%s] Expected status %d got %d",
				test.TestName, test.ExpectedStatus, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else if response.StatusCode == test.ExpectedStatus &&
			response.StatusCode != http.StatusForbidden {
			// get response data (product)
			var result []model.Product
			err = json.NewDecoder(response.Body).Decode(&result)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			// check response data length
			if len(test.ExpectedData) != len(result) {
				t.Errorf("[%s] Expected length data %d, but got %d",
					test.TestName, len(test.ExpectedData),
					len(result))
			}

			// check response data
			for _, ep := range test.ExpectedData {
				// check product exist
				pExist := model.Product{}
				for _, p := range result {
					if p.ProductInfo.SKU == ep.ProductInfo.SKU &&
						p.ProductInfo.Name == ep.ProductInfo.Name &&
						p.ProductInfo.Price == ep.ProductInfo.Price &&
						p.ProductInfo.Weight == ep.ProductInfo.Weight &&
						p.ProductInfo.Description == ep.ProductInfo.Description &&
						p.ProductInfo.Stock == ep.ProductInfo.Stock &&
						p.ProductInfo.UserID == ep.ProductInfo.UserID {
						pExist = p
						break
					}
				}

				if pExist.ProductInfo.ID == 0 {
					t.Errorf("[%s] Expected product with SKU %s "+
						"not found", test.TestName, ep.ProductInfo.SKU)
				} else {
					// check all product images exist
					for _, ePImage := range ep.ProductImages {
						ePImageExist := false
						for _, pImage := range pExist.ProductImages {
							if pImage.ImagePath == ePImage.ImagePath {
								ePImageExist = true
								break
							}
						}

						if !ePImageExist {
							t.Errorf("[%s] Expected product with image path %s "+
								"not found", test.TestName, ePImage.ImagePath)
						}
					}
				}
			}
		}
	}

	// truncate tables after test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestGetProductHandler test GetProductHandler
func TestGetProductHandler(t *testing.T) {
	// get testing API for create products
	u := middleware.User{}
	a, err := GetTestingAPI(u)
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	// create products
	sop := []model.Product{
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT A Path 1.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 2.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 3.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      2,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = model.InsertProductInfo(a.DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = a.DB.Exec(`INSERT INTO 
				product_productimage(image_path, product_productinfo_id)
				VALUES($1,$2)`,
				pImage.ImagePath, sop[i].ProductInfo.ID)
			if err != nil {
				t.Errorf("There's an error when insert data product image => %s",
					err.Error())
			}
		}
	}

	// create testing table
	testTable := []struct {
		TestName       string
		SKU            string
		User           middleware.User
		ExpectedStatus int
		ExpectedData   model.Product
	}{
		{
			TestName:       "Get Existed Product Success (User: Seller)",
			SKU:            sop[0].ProductInfo.SKU,
			User:           middleware.User{Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   sop[0],
		},
		{
			TestName:       "Get Existed Product Success (User: Buyer)",
			SKU:            sop[1].ProductInfo.SKU,
			User:           middleware.User{Role: "buyer"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   sop[1],
		},
		{
			TestName:       "Get Existed Product Bad Request (User: Seller)",
			SKU:            "",
			User:           middleware.User{Role: "seller"},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedData:   model.Product{},
		},
		{
			TestName:       "Get Existed Product Bad Request (User: Buyer)",
			SKU:            "",
			User:           middleware.User{Role: "buyer"},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedData:   model.Product{},
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// get testing API for get product
		a, err = GetTestingAPI(test.User)
		if err != nil {
			t.Errorf("There's an error when getting testing API => %s",
				err.Error())
		}

		// get url params
		params := url.Values{}
		params.Add("sku", test.SKU)
		params.Add("testing", "1")

		// create new request
		req, err := http.NewRequest("GET", "/api/product/", nil)
		req.URL.RawQuery = params.Encode()
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API get product => %s",
				test.TestName, err.Error())
		}

		// run request
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response status
		if response.StatusCode != test.ExpectedStatus {
			t.Errorf("[%s] Expected status %d got %d",
				test.TestName, test.ExpectedStatus, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else if response.StatusCode == test.ExpectedStatus &&
			response.StatusCode != http.StatusBadRequest {
			// get response data (product)
			var pResult model.Product
			err = json.NewDecoder(response.Body).Decode(&pResult)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			// check product info
			if test.ExpectedData.ProductInfo.SKU != pResult.ProductInfo.SKU {
				t.Errorf("Expected SKU '%s', but got SKU '%s'",
					test.ExpectedData.ProductInfo.SKU, pResult.ProductInfo.SKU)
			}
			if test.ExpectedData.ProductInfo.Name != pResult.ProductInfo.Name {
				t.Errorf("Expected Name '%s', but got Name '%s'",
					test.ExpectedData.ProductInfo.Name, pResult.ProductInfo.Name)
			}
			if test.ExpectedData.ProductInfo.Price != pResult.ProductInfo.Price {
				t.Errorf("Expected Price %f, but got Price %f",
					test.ExpectedData.ProductInfo.Price, pResult.ProductInfo.Price)
			}
			if test.ExpectedData.ProductInfo.Weight != pResult.ProductInfo.Weight {
				t.Errorf("Expected Weight %f, but got Weight %f",
					test.ExpectedData.ProductInfo.Weight, pResult.ProductInfo.Weight)
			}
			if test.ExpectedData.ProductInfo.Description != pResult.ProductInfo.Description {
				t.Errorf("Expected Description '%s', but got Description '%s'",
					test.ExpectedData.ProductInfo.Description, pResult.ProductInfo.Description)
			}
			if test.ExpectedData.ProductInfo.Stock != pResult.ProductInfo.Stock {
				t.Errorf("Expected Stock %d, but got Stock %d",
					test.ExpectedData.ProductInfo.Stock, pResult.ProductInfo.Stock)
			}
			if test.ExpectedData.ProductInfo.UserID != pResult.ProductInfo.UserID {
				t.Errorf("Expected UserID %d, but got UserID %d",
					test.ExpectedData.ProductInfo.UserID, pResult.ProductInfo.UserID)
			}

			// check all product images
			for _, ePImage := range test.ExpectedData.ProductImages {
				ePImageExist := false
				for _, pImage := range pResult.ProductImages {
					if pImage.ImagePath == ePImage.ImagePath {
						ePImageExist = true
						break
					}
				}

				if !ePImageExist {
					t.Errorf("Expected product with image path %s "+
						"not found", ePImage.ImagePath)
				}
			}
		}
	}

	// truncate tables after test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestUpdateProductHandler test UpdateProductHandler
//
// Required for the test:
//
// - AddProductHandler
//
// - GetProductHandler
func TestUpdateProductHandler(t *testing.T) {
	// initialize testing table
	testTable := []struct {
		TestName          string
		User              middleware.User
		FormData          map[string]string
		FormDataUpdate    map[string]string
		PriceAfterUpdate  float64
		WeightAfterUpdate float32
		StockAfterUpdate  int
		ExpectedStatus    int
	}{
		{
			TestName: "Test Update Product Success",
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"name":        "After Update",
				"price":       "2000000.50",
				"weight":      "2.5",
				"description": "After Update",
				"stock":       "200",
			},
			PriceAfterUpdate:  2000000.50,
			WeightAfterUpdate: 2.5,
			StockAfterUpdate:  200,
			ExpectedStatus:    http.StatusOK,
		},
		{
			TestName: "Test Update Product Forbidden",
			User: middleware.User{
				ID:   1,
				Role: "buyer",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"name":        "After Update",
				"price":       "2000000.50",
				"weight":      "2.5",
				"description": "After Update",
				"stock":       "200",
			},
			PriceAfterUpdate:  2000000.50,
			WeightAfterUpdate: 2.5,
			StockAfterUpdate:  200,
			ExpectedStatus:    http.StatusForbidden,
		},
		{
			TestName: "Test Update Product Bad Request 1",
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"name":        "",
				"price":       "2000000.50",
				"weight":      "2.5",
				"description": "After Update",
				"stock":       "200",
			},
			PriceAfterUpdate:  2000000.50,
			WeightAfterUpdate: 2.5,
			StockAfterUpdate:  200,
			ExpectedStatus:    http.StatusBadRequest,
		},
		{
			TestName: "Test Update Product Bad Request 2",
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"name":        "After Update",
				"price":       "",
				"weight":      "2.5",
				"description": "After Update",
				"stock":       "200",
			},
			PriceAfterUpdate:  2000000.50,
			WeightAfterUpdate: 2.5,
			StockAfterUpdate:  200,
			ExpectedStatus:    http.StatusBadRequest,
		},
		{
			TestName: "Test Update Product Bad Request 3",
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"name":        "After Update",
				"price":       "2000000.50",
				"weight":      "",
				"description": "After Update",
				"stock":       "200",
			},
			PriceAfterUpdate:  2000000.50,
			WeightAfterUpdate: 2.5,
			StockAfterUpdate:  200,
			ExpectedStatus:    http.StatusBadRequest,
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// initialize testing API for add product
		a, err := GetTestingAPI(middleware.User{ID: 1, Role: "seller"})
		if err != nil {
			t.Errorf("[%s] There's an error when getting testing API => %s",
				test.TestName, err.Error())
		}

		// transform form data to bytes buffer
		var bFormData bytes.Buffer
		w := multipart.NewWriter(&bFormData)
		for key, r := range test.FormData {
			fw, err := w.CreateFormField(key)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}

			_, err = io.Copy(fw, strings.NewReader(r))
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}
		}

		// add image files to form
		for i := 1; i <= 10; i++ {
			ffw, err := w.CreateFormFile("product_images", fmt.Sprintf("test_%d.png", i))
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}

			img := CreateTestImage()
			err = png.Encode(ffw, img)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}
		}

		w.Close()

		// create new request for add product
		req, err := http.NewRequest("POST", "/api/product/", &bFormData)
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API create product => %s",
				test.TestName, err.Error())
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		// run request for add product
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response add product
		if response.StatusCode != http.StatusCreated {
			t.Errorf("[%s] There's an error when add product data,"+
				" expected status %d but got %d",
				test.TestName, http.StatusCreated, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else {
			// get product info data after success add product
			var respPInfo model.ProductInfo
			err = json.NewDecoder(response.Body).Decode(&respPInfo)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			// initialize testing API for update product
			a, err := GetTestingAPI(test.User)
			if err != nil {
				t.Errorf("[%s] There's an error when getting testing API => %s",
					test.TestName, err.Error())
			}

			// transform form data update to bytes buffer
			var bFormData bytes.Buffer
			w := multipart.NewWriter(&bFormData)
			for key, r := range test.FormDataUpdate {
				fw, err := w.CreateFormField(key)
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}

				_, err = io.Copy(fw, strings.NewReader(r))
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}
			}

			// add image files to form update
			for i := 1; i <= 3; i++ {
				ffw, err := w.CreateFormFile("product_images",
					fmt.Sprintf("test_update_%d.png", i))
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}

				img := CreateTestImage()
				err = png.Encode(ffw, img)
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}
			}

			w.Close()

			// get url params for update product
			params := url.Values{}
			params.Add("sku", respPInfo.SKU)

			// create new request for update product
			req, err := http.NewRequest("PUT", "/api/product/", &bFormData)
			req.URL.RawQuery = params.Encode()
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"request API update product => %s",
					test.TestName, err.Error())
			}
			req.Header.Set("Content-Type", w.FormDataContentType())

			// run request for update product
			response, err := a.FiberApp.Test(req)
			if err != nil {
				t.Errorf("[%s] There's an error serve http testing => %s",
					test.TestName, err.Error())
			}
			defer response.Body.Close()

			// check response update product
			if response.StatusCode != test.ExpectedStatus {
				t.Errorf("[%s] Expected status %d but got %d",
					test.TestName, test.ExpectedStatus, response.StatusCode)

				var resp map[string]string
				err = json.NewDecoder(response.Body).Decode(&resp)
				if err != nil {
					t.Errorf("[%s] There's an error when "+
						"unmarshal body response => %s",
						test.TestName, err.Error())
				}

				t.Error(resp)

			} else if response.StatusCode == http.StatusOK {
				// initialize testing API for get product after update
				a, err := GetTestingAPI(middleware.User{ID: 1, Role: "seller"})
				if err != nil {
					t.Errorf("[%s] There's an error when getting testing API => %s",
						test.TestName, err.Error())
				}

				// get url params for get product after update
				params := url.Values{}
				params.Add("sku", respPInfo.SKU)
				params.Add("testing", "1")

				// create new request get product after update
				req, err := http.NewRequest("GET", "/api/product/", nil)
				req.URL.RawQuery = params.Encode()
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"request API get product => %s",
						test.TestName, err.Error())
				}

				// run request get product after update
				response, err := a.FiberApp.Test(req)
				if err != nil {
					t.Errorf("[%s] There's an error serve "+
						"http testing => %s",
						test.TestName, err.Error())
				}
				defer response.Body.Close()

				// check response status get product after update
				if response.StatusCode != http.StatusOK {
					t.Errorf("[%s] There's an error when get product data,"+
						" expected status %d but got %d",
						test.TestName, http.StatusOK, response.StatusCode)

					var resp map[string]string
					err = json.NewDecoder(response.Body).Decode(&resp)
					if err != nil {
						t.Errorf("[%s] There's an error when "+
							"unmarshal body response => %s",
							test.TestName, err.Error())
					}

					t.Error(resp)

				} else {
					// get response data (product)
					var pResult model.Product
					err = json.NewDecoder(response.Body).Decode(&pResult)
					if err != nil {
						t.Errorf("[%s] There's an error when "+
							"unmarshal body response => %s",
							test.TestName, err.Error())
					}

					// check product info
					if respPInfo.SKU != pResult.ProductInfo.SKU {
						t.Errorf("Expected SKU '%s', but got SKU '%s'",
							respPInfo.SKU, pResult.ProductInfo.SKU)
					}
					if test.FormDataUpdate["name"] !=
						pResult.ProductInfo.Name {
						t.Errorf("Expected Name '%s', but got Name '%s'",
							test.FormDataUpdate["name"],
							pResult.ProductInfo.Name)
					}
					if test.PriceAfterUpdate != pResult.ProductInfo.Price {
						t.Errorf("Expected Price %f, but got Price %f",
							test.PriceAfterUpdate, pResult.ProductInfo.Price)
					}
					if test.WeightAfterUpdate != pResult.ProductInfo.Weight {
						t.Errorf("Expected Weight %f, but got Weight %f",
							test.WeightAfterUpdate, pResult.ProductInfo.Weight)
					}
					if test.FormDataUpdate["description"] !=
						pResult.ProductInfo.Description {
						t.Errorf("Expected Description '%s', but got Description '%s'",
							test.FormDataUpdate["description"],
							pResult.ProductInfo.Description)
					}
					if test.StockAfterUpdate != pResult.ProductInfo.Stock {
						t.Errorf("Expected Stock %d, but got Stock %d",
							test.StockAfterUpdate, pResult.ProductInfo.Stock)
					}

					// check product images length
					if len(pResult.ProductImages) != 3 {
						t.Errorf("Expected total image 3, but got %d",
							len(pResult.ProductImages))
					}
				}
			}
		}
	}

	// truncate tables after test
	a, err := GetTestingAPI(middleware.User{})
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestDeleteProductHandler test DeleteProductHandler
func TestDeleteProductHandler(t *testing.T) {
	// get testing API for create products
	u := middleware.User{}
	a, err := GetTestingAPI(u)
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	// create products
	sop := []model.Product{
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT A Path 1.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 2.jpg",
				},
				{
					ImagePath: "PRODUCT A Path 3.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: model.ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []model.ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = model.InsertProductInfo(a.DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = a.DB.Exec(`INSERT INTO 
				product_productimage(image_path, product_productinfo_id)
				VALUES($1,$2)`,
				pImage.ImagePath, sop[i].ProductInfo.ID)
			if err != nil {
				t.Errorf("There's an error when insert data product image => %s",
					err.Error())
			}
		}
	}

	// create testing table
	testTable := []struct {
		TestName       string
		SKU            string
		User           middleware.User
		ExpectedStatus int
		ExpectedData   []model.Product
	}{
		{
			TestName:       "Delete Product Success 1",
			SKU:            sop[0].ProductInfo.SKU,
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{sop[1], sop[2]},
		},
		{
			TestName:       "Delete Product Success 2",
			SKU:            sop[1].ProductInfo.SKU,
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{sop[2]},
		},
		{
			TestName:       "Delete Product Success 3",
			SKU:            sop[2].ProductInfo.SKU,
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusOK,
			ExpectedData:   []model.Product{},
		},
		{
			TestName:       "Delete Product Forbidden",
			SKU:            "test",
			User:           middleware.User{ID: 1, Role: "buyer"},
			ExpectedStatus: http.StatusForbidden,
			ExpectedData:   []model.Product{},
		},
		{
			TestName:       "Delete Product Bad Request",
			SKU:            "",
			User:           middleware.User{ID: 1, Role: "seller"},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedData:   []model.Product{},
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// get testing API
		a, err = GetTestingAPI(test.User)
		if err != nil {
			t.Errorf("There's an error when getting testing API => %s",
				err.Error())
		}

		// get url params for delete product
		params := url.Values{}
		params.Add("sku", test.SKU)

		// create new request for delete product
		req, err := http.NewRequest("DELETE", "/api/product/", nil)
		req.URL.RawQuery = params.Encode()
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API get products => %s",
				test.TestName, err.Error())
		}

		// run request delete product
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response status request delete product
		if response.StatusCode != test.ExpectedStatus {
			t.Errorf("[%s] Expected status %d got %d",
				test.TestName, test.ExpectedStatus, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else if response.StatusCode == http.StatusOK {
			// create new request for get products by user id
			req, err := http.NewRequest("GET", "/api/products/user/", nil)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"request API get products => %s",
					test.TestName, err.Error())
			}

			// run request get products by user id
			response, err := a.FiberApp.Test(req)
			if err != nil {
				t.Errorf("[%s] There's an error serve http testing => %s",
					test.TestName, err.Error())
			}
			defer response.Body.Close()

			// check response status request get products by user id
			if response.StatusCode != http.StatusOK {
				t.Errorf("[%s] There's an error when get products by user id, "+
					"expected status %d got %d",
					test.TestName, http.StatusOK, response.StatusCode)

				var resp map[string]string
				err = json.NewDecoder(response.Body).Decode(&resp)
				if err != nil {
					t.Errorf("[%s] There's an error when unmarshal body response => %s",
						test.TestName, err.Error())
				}

				t.Error(resp)

			} else {
				// get response data (product)
				var result []model.Product
				err = json.NewDecoder(response.Body).Decode(&result)
				if err != nil {
					t.Errorf("[%s] There's an error when unmarshal body response => %s",
						test.TestName, err.Error())
				}

				// check response data length
				if len(test.ExpectedData) != len(result) {
					t.Errorf("[%s] Expected length data %d, but got %d",
						test.TestName, len(test.ExpectedData),
						len(result))
				}

				// check response data
				for _, ep := range test.ExpectedData {
					// check product exist
					pExist := model.Product{}
					for _, p := range result {
						if p.ProductInfo.SKU == ep.ProductInfo.SKU &&
							p.ProductInfo.Name == ep.ProductInfo.Name &&
							p.ProductInfo.Price == ep.ProductInfo.Price &&
							p.ProductInfo.Weight == ep.ProductInfo.Weight &&
							p.ProductInfo.Description == ep.ProductInfo.Description &&
							p.ProductInfo.Stock == ep.ProductInfo.Stock &&
							p.ProductInfo.UserID == ep.ProductInfo.UserID {
							pExist = p
							break
						}
					}

					if pExist.ProductInfo.ID == 0 {
						t.Errorf("[%s] Expected product with SKU %s "+
							"not found", test.TestName, ep.ProductInfo.SKU)
					} else {
						// check all product images exist
						for _, ePImage := range ep.ProductImages {
							ePImageExist := false
							for _, pImage := range pExist.ProductImages {
								if pImage.ImagePath == ePImage.ImagePath {
									ePImageExist = true
									break
								}
							}

							if !ePImageExist {
								t.Errorf("[%s] Expected product with image path %s "+
									"not found", test.TestName, ePImage.ImagePath)
							}
						}
					}
				}
			}
		}
	}

	// truncate tables after test
	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestDecreaseStockHandler test DecreaseStockHandler
//
// Required for the test:
//
// - AddProductHandler
//
// - GetProductHandler
func TestDecreaseStockHandler(t *testing.T) {
	// initialize testing table
	testTable := []struct {
		TestName         string
		User             middleware.User
		FormData         map[string]string
		FormDataUpdate   map[string]string
		StockAfterUpdate int
		ExpectedStatus   int
	}{
		{
			TestName: "Test Decrease Stock Success",
			User: middleware.User{
				ID:   1,
				Role: "seller",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"qty": "10",
			},
			StockAfterUpdate: 90,
			ExpectedStatus:   http.StatusOK,
		},
		{
			TestName: "Test Decrease Stock Forbidden",
			User: middleware.User{
				ID:   1,
				Role: "buyer",
			},
			FormData: map[string]string{
				"name":        "Before Update",
				"price":       "1000000.50",
				"weight":      "1.5",
				"description": "Before Update",
				"stock":       "100",
			},
			FormDataUpdate: map[string]string{
				"qty": "10",
			},
			ExpectedStatus: http.StatusForbidden,
		},
	}

	// loop test in test table
	for _, test := range testTable {
		// initialize testing API for add product
		a, err := GetTestingAPI(middleware.User{ID: 1, Role: "seller"})
		if err != nil {
			t.Errorf("[%s] There's an error when getting testing API => %s",
				test.TestName, err.Error())
		}

		// transform form data to bytes buffer
		var bFormData bytes.Buffer
		w := multipart.NewWriter(&bFormData)
		for key, r := range test.FormData {
			fw, err := w.CreateFormField(key)
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}

			_, err = io.Copy(fw, strings.NewReader(r))
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"bytes buffer form data => %s",
					test.TestName, err.Error())
			}
		}

		w.Close()

		// create new request for add product
		req, err := http.NewRequest("POST", "/api/product/", &bFormData)
		if err != nil {
			t.Errorf("[%s] There's an error when creating "+
				"request API create product => %s",
				test.TestName, err.Error())
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		// run request for add product
		response, err := a.FiberApp.Test(req)
		if err != nil {
			t.Errorf("[%s] There's an error serve http testing => %s",
				test.TestName, err.Error())
		}
		defer response.Body.Close()

		// check response add product
		if response.StatusCode != http.StatusCreated {
			t.Errorf("[%s] There's an error when add product data,"+
				" expected status %d but got %d",
				test.TestName, http.StatusCreated, response.StatusCode)

			var resp map[string]string
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			t.Error(resp)

		} else {
			// get product info data after success add product
			var respPInfo model.ProductInfo
			err = json.NewDecoder(response.Body).Decode(&respPInfo)
			if err != nil {
				t.Errorf("[%s] There's an error when unmarshal body response => %s",
					test.TestName, err.Error())
			}

			// initialize testing API for decrease product stock
			a, err := GetTestingAPI(test.User)
			if err != nil {
				t.Errorf("[%s] There's an error when getting testing API => %s",
					test.TestName, err.Error())
			}

			// transform form data update to bytes buffer
			var bFormData bytes.Buffer
			w := multipart.NewWriter(&bFormData)
			for key, r := range test.FormDataUpdate {
				fw, err := w.CreateFormField(key)
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}

				_, err = io.Copy(fw, strings.NewReader(r))
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"bytes buffer form data => %s",
						test.TestName, err.Error())
				}
			}

			w.Close()

			// get url params for decrease product stock
			params := url.Values{}
			params.Add("sku", respPInfo.SKU)

			// create new request for decrease product stock
			req, err := http.NewRequest("PUT", "/api/product/decrease/stock/", &bFormData)
			req.URL.RawQuery = params.Encode()
			if err != nil {
				t.Errorf("[%s] There's an error when creating "+
					"request API decrease product stock => %s",
					test.TestName, err.Error())
			}
			req.Header.Set("Content-Type", w.FormDataContentType())

			// run request for update product
			response, err := a.FiberApp.Test(req)
			if err != nil {
				t.Errorf("[%s] There's an error serve http testing => %s",
					test.TestName, err.Error())
			}
			defer response.Body.Close()

			// check response update product
			if response.StatusCode != test.ExpectedStatus {
				t.Errorf("[%s] Expected status %d but got %d",
					test.TestName, test.ExpectedStatus, response.StatusCode)

				var resp map[string]string
				err = json.NewDecoder(response.Body).Decode(&resp)
				if err != nil {
					t.Errorf("[%s] There's an error when "+
						"unmarshal body response => %s",
						test.TestName, err.Error())
				}

				t.Error(resp)

			} else if response.StatusCode == http.StatusOK {
				// initialize testing API for get product after update
				a, err := GetTestingAPI(middleware.User{ID: 1, Role: "seller"})
				if err != nil {
					t.Errorf("[%s] There's an error when getting testing API => %s",
						test.TestName, err.Error())
				}

				// get url params for get product after update
				params := url.Values{}
				params.Add("sku", respPInfo.SKU)
				params.Add("testing", "1")

				// create new request get product after update
				req, err := http.NewRequest("GET", "/api/product/", nil)
				req.URL.RawQuery = params.Encode()
				if err != nil {
					t.Errorf("[%s] There's an error when creating "+
						"request API get product => %s",
						test.TestName, err.Error())
				}

				// run request get product after update
				response, err := a.FiberApp.Test(req)
				if err != nil {
					t.Errorf("[%s] There's an error serve "+
						"http testing => %s",
						test.TestName, err.Error())
				}
				defer response.Body.Close()

				// check response status get product after update
				if response.StatusCode != http.StatusOK {
					t.Errorf("[%s] There's an error when get product data,"+
						" expected status %d but got %d",
						test.TestName, http.StatusOK, response.StatusCode)

					var resp map[string]string
					err = json.NewDecoder(response.Body).Decode(&resp)
					if err != nil {
						t.Errorf("[%s] There's an error when "+
							"unmarshal body response => %s",
							test.TestName, err.Error())
					}

					t.Error(resp)

				} else {
					// get response data (product)
					var pResult model.Product
					err = json.NewDecoder(response.Body).Decode(&pResult)
					if err != nil {
						t.Errorf("[%s] There's an error when "+
							"unmarshal body response => %s",
							test.TestName, err.Error())
					}

					// check product info
					if respPInfo.SKU != pResult.ProductInfo.SKU {
						t.Errorf("Expected SKU '%s', but got SKU '%s'",
							respPInfo.SKU, pResult.ProductInfo.SKU)
					}
					if test.StockAfterUpdate != pResult.ProductInfo.Stock {
						t.Errorf("Expected Stock %d, but got Stock %d",
							test.StockAfterUpdate, pResult.ProductInfo.Stock)
					}

				}
			}
		}
	}

	// truncate tables after test
	a, err := GetTestingAPI(middleware.User{})
	if err != nil {
		t.Errorf("There's an error when getting testing API => %s",
			err.Error())
	}

	_, err = a.DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// GetTestingAPI get API for testing
func GetTestingAPI(u middleware.User) (API, error) {
	a := API{}

	// init database
	DBConfig := map[string]string{
		"user":     config.DBUsername,
		"password": config.DBPassword,
		"dbname":   config.DBNameForAPITest,
	}
	err := a.InitDB(DBConfig)
	if err != nil {
		return a, err
	}

	// init router
	a.FiberApp = fiber.New()
	mainRouter := a.FiberApp.Group("")
	mainRouter.Use(AuthorizationMiddlewareForTest(u))
	mainRouter.Post("/api/product/", a.AddProductHandler)
	mainRouter.Get("/api/products/", a.GetProductsHandler)
	mainRouter.Get("/api/products/user/", a.GetProductsByUserIDHandler)
	mainRouter.Get("/api/product/", a.GetProductHandler)
	mainRouter.Put("/api/product/", a.UpdateProductHandler)
	mainRouter.Delete("/api/product/", a.DeleteProductHandler)
	mainRouter.Put("/api/product/decrease/stock/", a.DecreaseStockHandler)

	// change media folder to testing media folder
	config.MediaFolder = "media-test"

	return a, nil
}

// CreateTestImage create testing image
// from: https://yourbasic.org/golang/create-image/
func CreateTestImage() *image.RGBA {
	width := 200
	height := 100

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	return img
}

// AuthorizationMiddlewareForTest middleware authorization for testing
func AuthorizationMiddlewareForTest(u middleware.User) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user", u)
		return c.Next()
	}
}
