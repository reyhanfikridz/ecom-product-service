/*
Package model containing structs and functions for
database transaction
*/
package model

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"time"

	"github.com/reyhanfikridz/ecom-product-service/internal/config"
	"github.com/reyhanfikridz/ecom-product-service/internal/utils"
)

// ProductInfo contain basic information of a product
type ProductInfo struct {
	ID          int     `json:"id" form:"id"`
	SKU         string  `json:"sku" form:"sku"`
	Name        string  `json:"name" form:"name"`
	Price       float64 `json:"price" form:"price"`
	Weight      float32 `json:"weight" form:"weight"`
	Description string  `json:"description" form:"description"`
	Stock       int     `json:"stock" form:"stock"`
	UserID      int     `json:"user_id" form:"user_id"`
}

// ProductImage contain image of a product
type ProductImage struct {
	ID          int         `json:"id" form:"id"`
	ImagePath   string      `json:"image_path" form:"image_path"`
	ProductInfo ProductInfo `json:"product_info" form:"product_info"`
}

// Product contain product info, product images, and seller info
type Product struct {
	ProductInfo   ProductInfo    `json:"product_info"`
	ProductImages []ProductImage `json:"product_images"`
	SellerInfo    struct {
		Email       string `json:"email"`
		FullName    string `json:"full_name"`
		Address     string `json:"address"`
		PhoneNumber string `json:"phone_number"`
	} `json:"seller_info"`
}

// InsertProductInfo insert a product info into database
func InsertProductInfo(DB *sql.DB, pInfo ProductInfo) (ProductInfo, error) {
	// begin transaction
	tx, err := DB.Begin()
	if err != nil {
		return pInfo, err
	}
	defer tx.Rollback()

	// get unique random SKU
	var SKU string
	for {
		SKU = utils.GetRandomSKU()

		var tmpID int
		err := tx.QueryRow(`
			SELECT id 
			FROM product_productinfo
			WHERE sku = $1
		`, SKU).Scan(&tmpID)
		if err != nil && err != sql.ErrNoRows {
			return pInfo, err
		} else if err != nil && err == sql.ErrNoRows {
			break
		}
	}

	// insert product info, returning product info ID and SKU
	row := tx.QueryRow(`INSERT INTO 
		product_productinfo(
			sku, name, weight, price, description, stock, account_user_id) 
		VALUES($1,$2,$3,$4,$5,$6,$7) returning id, sku`,
		SKU, pInfo.Name, pInfo.Weight, pInfo.Price,
		pInfo.Description, pInfo.Stock, pInfo.UserID)

	if row.Err() != nil {
		return pInfo, row.Err()
	}

	err = row.Scan(&pInfo.ID, &pInfo.SKU)
	if err != nil {
		return pInfo, err
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return pInfo, err
	}

	return pInfo, nil
}

// InsertProductImages insert product images into database
// and save the product image files into media folder
func InsertProductImages(DB *sql.DB, fileHeaders []*multipart.FileHeader,
	pInfo ProductInfo) error {
	// begin transaction
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // rollback transaction if fail

	// delete existed images first
	_, err = tx.Exec(`DELETE FROM product_productimage 
		WHERE product_productinfo_id = $1`,
		pInfo.ID)
	if err != nil {
		return err
	}

	// loop the image file headers
	for _, fileHeader := range fileHeaders {
		// save the image file into media folder
		imagePath, err := SaveProductImage(fileHeader)
		if err != nil {
			return err
		}

		// insert product image into database
		_, err = tx.Exec(`INSERT INTO 
			product_productimage(image_path, product_productinfo_id)
			VALUES($1,$2)`,
			imagePath, pInfo.ID)
		if err != nil {
			return err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// SaveProductImage save product image file into media folder
func SaveProductImage(fileHeader *multipart.FileHeader) (string, error) {
	// open the file
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// create the product image folder first if not exist
	err = os.MkdirAll(fmt.Sprintf("./../%s/product-image/", config.MediaFolder),
		os.ModePerm)
	if err != nil {
		return "", err
	}

	// create a new image file in the product image directory
	image_path := fmt.Sprintf("product-image/%d-%s",
		time.Now().UnixNano(),
		fileHeader.Filename)
	dst, err := os.Create(fmt.Sprintf("./../%s/%s", config.MediaFolder, image_path))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// copy the uploaded image file to the image file
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return image_path, nil
}

// GetProducts get products from database by key filter and/or search
func GetProducts(DB *sql.DB, filter ProductInfo, search string) ([]Product, error) {
	sop := []Product{}

	// get query string
	q := `
		SELECT 
			id, sku, name, price, weight, description, stock, account_user_id 
		FROM product_productinfo
	`

	if filter.UserID != 0 {
		q += ` WHERE account_user_id = ` + strconv.Itoa(filter.UserID)
		if search != "" {
			q += ` AND (name ILIKE '%` + search + `%'
					OR description ILIKE '%` + search + `%')`
		}
	} else {
		if search != "" {
			q += ` WHERE (name ILIKE '%` + search + `%'
				OR description ILIKE '%` + search + `%')`
		}
	}

	q += ` ORDER BY id`

	// get product info rows
	rows, err := DB.Query(q)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	// loop product info rows
	for rows.Next() {
		p := Product{}

		// scan product info row
		err = rows.Scan(
			&p.ProductInfo.ID, &p.ProductInfo.SKU,
			&p.ProductInfo.Name, &p.ProductInfo.Price,
			&p.ProductInfo.Weight, &p.ProductInfo.Description,
			&p.ProductInfo.Stock, &p.ProductInfo.UserID)
		if err != nil {
			return []Product{}, err
		}

		// get product image rows
		imageRows, err := DB.Query(`
			SELECT 
				id, image_path
			FROM product_productimage
			WHERE product_productinfo_id = $1`,
			p.ProductInfo.ID)
		if err != nil {
			return []Product{}, err
		}

		// loop product image rows
		for imageRows.Next() {
			// scan product image row
			pImage := ProductImage{}
			err = imageRows.Scan(&pImage.ID, &pImage.ImagePath)
			if err != nil {
				return []Product{}, err
			}

			p.ProductImages = append(p.ProductImages, pImage)
		}
		imageRows.Close()

		// put product info and product images into product
		sop = append(sop, p)
	}

	return sop, nil
}

// GetProductBySKU get one product from database by key SKU
func GetProductBySKU(DB *sql.DB, SKU string) (Product, error) {
	p := Product{}

	// get product info
	row := DB.QueryRow(`
		SELECT
			id, sku, name, price, weight, description, stock, account_user_id
		FROM product_productinfo
		WHERE sku = $1
	`, SKU)
	if row.Err() != nil {
		return p, row.Err()
	}

	err := row.Scan(
		&p.ProductInfo.ID, &p.ProductInfo.SKU, &p.ProductInfo.Name,
		&p.ProductInfo.Price, &p.ProductInfo.Weight,
		&p.ProductInfo.Description, &p.ProductInfo.Stock, &p.ProductInfo.UserID)
	if err != nil {
		return p, err
	}

	// get product images
	imageRows, err := DB.Query(`
		SELECT 
			id, image_path
		FROM product_productimage
		WHERE product_productinfo_id = $1`,
		p.ProductInfo.ID)
	if err != nil {
		return Product{}, err
	}

	for imageRows.Next() {
		pImage := ProductImage{}
		err = imageRows.Scan(&pImage.ID, &pImage.ImagePath)
		if err != nil {
			return Product{}, err
		}

		p.ProductImages = append(p.ProductImages, pImage)
	}
	imageRows.Close()

	return p, nil
}

// UpdateProductInfoBySKU update product info in database by key SKU
func UpdateProductInfoBySKU(DB *sql.DB, pInfo ProductInfo) (ProductInfo, error) {
	// begin transaction
	tx, err := DB.Begin()
	if err != nil {
		return pInfo, err
	}
	defer tx.Rollback() // rollback transaction if fail

	// execute query update, returning product info ID
	row := tx.QueryRow(`
		UPDATE product_productinfo 
		SET name = $1, price = $2, weight = $3, description = $4, 
			stock = $5, account_user_id = $6 WHERE sku = $7
		RETURNING id`,
		pInfo.Name, pInfo.Price, pInfo.Weight, pInfo.Description,
		pInfo.Stock, pInfo.UserID, pInfo.SKU)
	if row.Err() != nil {
		return pInfo, row.Err()
	}

	err = row.Scan(&pInfo.ID)
	if err != nil {
		return pInfo, err
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return pInfo, err
	}

	return pInfo, nil
}

// DeleteProductBySKU delete product in database with key SKU
func DeleteProductBySKU(DB *sql.DB, SKU string) error {
	// begin transaction
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // rollback transaction if fail

	// delete product
	_, err = tx.Exec(`DELETE FROM product_productinfo WHERE sku = $1`, SKU)
	if err != nil {
		return err
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
