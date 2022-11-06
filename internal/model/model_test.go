/*
Package model containing structs and functions for
database transaction
*/
package model

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/reyhanfikridz/ecom-product-service/internal/config"
)

// TestMain do some test before and after all testing in the package
func TestMain(m *testing.M) {
	// init all config before can be used
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("There's an error when initialize config => %s",
			err.Error())
	}

	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		log.Fatalf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// truncate product tables before all test
	_, err = DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo before running all test => %s",
			err.Error())
	}

	// run all testing
	m.Run()

	// truncate product tables after all test
	_, err = DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo before running all test => %s",
			err.Error())
	}
}

// TestInsertProductInfo test for InsertProductInfo
func TestInsertProductInfo(t *testing.T) {
	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		t.Errorf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// create product info
	pInfo := ProductInfo{
		Name:        "ABC Product",
		Price:       1200000.55,
		Weight:      1.5,
		Description: "Decription 123",
		Stock:       100,
		UserID:      1,
	}

	// insert product info into database
	pInfo, err = InsertProductInfo(DB, pInfo)

	// check result
	if err != nil {
		t.Errorf("Expected error nil, but got error not nil => %s", err.Error())
	}

	if pInfo.ID == 0 {
		t.Errorf("Expected ID not zero, but got zero")
	}
	if pInfo.SKU == "" {
		t.Errorf("Expected SKU not empty string, but got empty string")
	}

	// truncate tables after test
	_, err = DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestGetProducts test for GetProducts
//
// Required for the test: InsertProductInfo
func TestGetProducts(t *testing.T) {
	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		t.Errorf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// create products
	sop := []Product{
		{
			ProductInfo: ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []ProductImage{
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
			ProductInfo: ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      2,
			},
			ProductImages: []ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = InsertProductInfo(DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = DB.Exec(`INSERT INTO 
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
		Filter         ProductInfo
		Search         string
		ExpectedResult []Product
	}{
		{
			TestName:       "Get All",
			Filter:         ProductInfo{},
			Search:         "",
			ExpectedResult: sop,
		},
		{
			TestName: "Get By User ID <1>",
			Filter: ProductInfo{
				UserID: 1,
			},
			Search:         "",
			ExpectedResult: []Product{sop[0], sop[2]},
		},
		{
			TestName:       "Get By Search <product b>",
			Filter:         ProductInfo{},
			Search:         "product b",
			ExpectedResult: []Product{sop[1], sop[2]},
		},
		{
			TestName: "Get By UserID <1> and Search <product b>",
			Filter: ProductInfo{
				UserID: 1,
			},
			Search:         "product b",
			ExpectedResult: []Product{sop[2]},
		},
	}

	// loop test in test table
	for _, test := range testTable {
		result, err := GetProducts(DB, test.Filter, test.Search)

		// check err result
		if err != nil {
			t.Errorf("Expected error nil, but got not nil => %s", err.Error())
		}

		// check result
		for _, ep := range test.ExpectedResult {
			// check product exist
			pExist := Product{}
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

	// truncate tables after test
	_, err = DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestGetProductBySKU test GetProductBySKU
//
// Required for the test: InsertProductInfo
func TestGetProductBySKU(t *testing.T) {
	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		t.Errorf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// create products
	sop := []Product{
		{
			ProductInfo: ProductInfo{
				Name:        "PRODUCT A",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT A",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []ProductImage{
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
			ProductInfo: ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      2,
			},
			ProductImages: []ProductImage{
				{
					ImagePath: "PRODUCT B Path 1.jpg",
				},
			},
		},
		{
			ProductInfo: ProductInfo{
				Name:        "PRODUCT B",
				Price:       1200000.55,
				Weight:      1.5,
				Description: "Description PRODUCT B",
				Stock:       100,
				UserID:      1,
			},
			ProductImages: []ProductImage{},
		},
	}

	// loop products
	for i := range sop {
		// insert product info into database
		sop[i].ProductInfo, err = InsertProductInfo(DB, sop[i].ProductInfo)
		if err != nil {
			t.Errorf("There's an error when insert data product info => %s",
				err.Error())
		}

		// insert product images into database without save image files
		for _, pImage := range sop[i].ProductImages {
			_, err = DB.Exec(`INSERT INTO 
				product_productimage(image_path, product_productinfo_id)
				VALUES($1,$2)`,
				pImage.ImagePath, sop[i].ProductInfo.ID)
			if err != nil {
				t.Errorf("There's an error when insert data product image => %s",
					err.Error())
			}
		}

		// test get product by SKU
		pResult, err := GetProductBySKU(DB, sop[i].ProductInfo.SKU)

		// check err result
		if err != nil {
			t.Errorf("Expected error nil, but got not nil => %s", err.Error())
		}

		// check product info
		if sop[i].ProductInfo.SKU != pResult.ProductInfo.SKU {
			t.Errorf("Expected SKU '%s', but got SKU '%s'",
				sop[i].ProductInfo.SKU, pResult.ProductInfo.SKU)
		}
		if sop[i].ProductInfo.Name != pResult.ProductInfo.Name {
			t.Errorf("Expected Name '%s', but got Name '%s'",
				sop[i].ProductInfo.Name, pResult.ProductInfo.Name)
		}
		if sop[i].ProductInfo.Price != pResult.ProductInfo.Price {
			t.Errorf("Expected Price %f, but got Price %f",
				sop[i].ProductInfo.Price, pResult.ProductInfo.Price)
		}
		if sop[i].ProductInfo.Weight != pResult.ProductInfo.Weight {
			t.Errorf("Expected Weight %f, but got Weight %f",
				sop[i].ProductInfo.Weight, pResult.ProductInfo.Weight)
		}
		if sop[i].ProductInfo.Description != pResult.ProductInfo.Description {
			t.Errorf("Expected Description '%s', but got Description '%s'",
				sop[i].ProductInfo.Description, pResult.ProductInfo.Description)
		}
		if sop[i].ProductInfo.Stock != pResult.ProductInfo.Stock {
			t.Errorf("Expected Stock %d, but got Stock %d",
				sop[i].ProductInfo.Stock, pResult.ProductInfo.Stock)
		}
		if sop[i].ProductInfo.UserID != pResult.ProductInfo.UserID {
			t.Errorf("Expected UserID %d, but got UserID %d",
				sop[i].ProductInfo.UserID, pResult.ProductInfo.UserID)
		}

		// check all product images
		for _, ePImage := range sop[i].ProductImages {
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

	// truncate tables after test
	_, err = DB.Exec("TRUNCATE product_productinfo RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("There's an error when truncating "+
			"table product_productinfo => %s",
			err.Error())
	}
}

// TestUpdateProductInfoBySKU test UpdateProductInfoBySKU
//
// Required for the test:
//
// - InsertProductInfo
//
// - GetProductBySKU
func TestUpdateProductInfoBySKU(t *testing.T) {
	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		t.Errorf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// create test table
	testTable := []struct {
		TestName                string
		ProductInfoBeforeUpdate ProductInfo
		ExpectedResult          Product
	}{
		{
			TestName: "Test Update ProductInfo 1",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "After update",
					Price:       1,
					Weight:      1,
					Description: "Before update",
					Stock:       1,
					UserID:      1,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 2",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       333333.3333,
					Weight:      1,
					Description: "Before update",
					Stock:       1,
					UserID:      1,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 3",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      3.51,
					Description: "Before update",
					Stock:       1,
					UserID:      1,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 4",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      1,
					Description: "After update",
					Stock:       1,
					UserID:      1,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 5",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      1,
					Description: "Before update",
					Stock:       111,
					UserID:      1,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 6",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      1,
					Description: "Before update",
					Stock:       1,
					UserID:      2,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 7",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "After update",
					Price:       1111111112.231344,
					Weight:      11.1231313131,
					Description: "After update",
					Stock:       1000,
					UserID:      2,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 8",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1111111112.231344,
					Weight:      11.1231313131,
					Description: "After update",
					Stock:       1000,
					UserID:      2,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 9",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      11.1231313131,
					Description: "After update",
					Stock:       1000,
					UserID:      2,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 10",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      1,
					Description: "After update",
					Stock:       1000,
					UserID:      2,
				},
			},
		},
		{
			TestName: "Test Update ProductInfo 11",
			ProductInfoBeforeUpdate: ProductInfo{
				Name:        "Before update",
				Price:       1,
				Weight:      1,
				Description: "Before update",
				Stock:       1,
				UserID:      1,
			},
			ExpectedResult: Product{
				ProductInfo: ProductInfo{
					Name:        "Before update",
					Price:       1,
					Weight:      1,
					Description: "Before update",
					Stock:       1000,
					UserID:      2,
				},
			},
		},
	}

	// do the test
	for _, test := range testTable {
		// insert product info to database
		test.ProductInfoBeforeUpdate, err = InsertProductInfo(DB, test.ProductInfoBeforeUpdate)
		if err != nil {
			t.Errorf("[%s] There's an error "+
				"when creating product data => %s",
				test.TestName, err.Error())
		}

		// update product info by SKU
		test.ExpectedResult.ProductInfo.SKU = test.ProductInfoBeforeUpdate.SKU
		_, err = UpdateProductInfoBySKU(DB, test.ExpectedResult.ProductInfo)
		if err != nil {
			t.Errorf("[%s] There's an error "+
				"when updating product data => %s",
				test.TestName, err.Error())
		}

		// get product info by SKU and check the result
		result, err := GetProductBySKU(DB, test.ProductInfoBeforeUpdate.SKU)

		if err != nil {
			t.Errorf("[%s] Expected error nil, but got not nil => %s",
				test.TestName, err.Error())
		}

		if test.ExpectedResult.ProductInfo.SKU != result.ProductInfo.SKU {
			t.Errorf("[%s] Expected SKU '%s', but got SKU '%s'",
				test.TestName, test.ExpectedResult.ProductInfo.SKU,
				result.ProductInfo.SKU)
		}
		if test.ExpectedResult.ProductInfo.Name != result.ProductInfo.Name {
			t.Errorf("[%s] Expected Name '%s', but got Name '%s'",
				test.TestName, test.ExpectedResult.ProductInfo.Name,
				result.ProductInfo.Name)
		}
		if test.ExpectedResult.ProductInfo.Price != result.ProductInfo.Price {
			t.Errorf("[%s] Expected Price '%f', but got Price '%f'",
				test.TestName, test.ExpectedResult.ProductInfo.Price,
				result.ProductInfo.Price)
		}
		if test.ExpectedResult.ProductInfo.Weight != result.ProductInfo.Weight {
			t.Errorf("[%s] Expected Weight '%f', but got Weight '%f'",
				test.TestName, test.ExpectedResult.ProductInfo.Weight,
				result.ProductInfo.Weight)
		}
		if test.ExpectedResult.ProductInfo.Description != result.ProductInfo.Description {
			t.Errorf("[%s] Expected Description '%s', but got Description '%s'",
				test.TestName, test.ExpectedResult.ProductInfo.Description,
				result.ProductInfo.Description)
		}
		if test.ExpectedResult.ProductInfo.Stock != result.ProductInfo.Stock {
			t.Errorf("[%s] Expected Stock '%d', but got Stock '%d'",
				test.TestName, test.ExpectedResult.ProductInfo.Stock,
				result.ProductInfo.Stock)
		}
		if test.ExpectedResult.ProductInfo.UserID != result.ProductInfo.UserID {
			t.Errorf("[%s] Expected UserID '%d', but got UserID '%d'",
				test.TestName, test.ExpectedResult.ProductInfo.UserID,
				result.ProductInfo.UserID)
		}

		// delete data after test
		_, err = DB.Exec(`DELETE FROM product_productinfo WHERE sku = $1`,
			test.ExpectedResult.ProductInfo.SKU)
		if err != nil {
			t.Errorf("[%s] There's an error "+
				"when deleting previous product data => %s",
				test.TestName, err.Error())
		}
	}
}

// TestDeleteProductBySKU test DeleteProductBySKU
//
// Required for the test:
//
// - InsertProductInfo
//
// - GetProductBySKU
func TestDeleteProductBySKU(t *testing.T) {
	// get testing DB connection
	DB, err := getTestDBConnection()
	if err != nil {
		t.Errorf("There's an error when initialize "+
			"testing database connection => %s", err.Error())
	}

	// insert product info into database
	p := ProductInfo{
		Name:        "AAA",
		Price:       100000.00,
		Weight:      1.5,
		Description: "BBB",
		Stock:       100,
		UserID:      1,
	}

	p, err = InsertProductInfo(DB, p)
	if err != nil {
		t.Errorf("There's an error "+
			"when creating product data => %s",
			err.Error())
	}

	// delete product by key SKU
	err = DeleteProductBySKU(DB, p.SKU)
	if err != nil {
		t.Errorf("Expected error nil when deleting data, "+
			"but got error => %s", err.Error())
	}

	// get product by SKU and check the result
	_, err = GetProductBySKU(DB, p.SKU)
	if err == nil {
		t.Errorf("Expected an error when getting data, but got no error")
	}
}

// getTestDBConnection get testing DB connection for package model testing
func getTestDBConnection() (*sql.DB, error) {
	// connect to DB
	connString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		config.DBUsername, config.DBPassword, config.DBNameForModelTest)
	DB, err := sql.Open("postgres", connString)
	if err != nil {
		return DB, err
	}

	// create tables if not exists
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
	_, err = DB.Exec(tableCreationQuery)
	if err != nil {
		return DB, err
	}

	return DB, nil
}
