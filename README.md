# ecom-product-service

### ECOM summary:
ECOM is a simple E-Commerce website builded with Go backend microservices and Django frontend. Disclaimer! I have zero real experience in building E-Commerce system, so if the system is really bad, I apologized in advance. This is just my personal project using Go microservices. You can use all the code of this project as a template for real E-Commerce in the future if you like it. Disclaimer again! I also not a frontend specialist, so I just use a free template I found in the internet and an original bootstrap template.

### Repository summary:
This is a microservice for ECOM that related to customer products CRUD.

### Requirements:
1. go (recommended: v1.18.4)
2. postgresql (recommended: v13.4)

### Microservice requirements:
1. ecom-account-service (must: https://github.com/reyhanfikridz/ecom-account-service/tree/release-1)

### Steps to run the server:
1. install all requirements
2. install and run all microservice requirements
3. clone repository at directory `$GOPATH/src/github.com/`
4. install required go library with `go mod download` then `go mod vendor` at repository root directory (same level as README.md)
5. create file .env at repository root directory (same level as README.md) with contents:

```
ECOM_PRODUCT_SERVICE_DB_NAME=<database name, example:ecom_product_service>
ECOM_PRODUCT_SERVICE_DB_NAME_FOR_API_TEST=<database name for overall api testing, example: ecom_product_service_api_test>
ECOM_PRODUCT_SERVICE_DB_NAME_FOR_MODEL_TEST=<database name for model crud testing, example: ecom_product_service_model_test>
ECOM_PRODUCT_SERVICE_DB_USERNAME=<postgresql username>
ECOM_PRODUCT_SERVICE_DB_PASSWORD=<postgresql password>

ECOM_PRODUCT_SERVICE_JWT_SECRET_KEY=<ecom account service jwt secret key>

ECOM_PRODUCT_SERVICE_URL=<this service url, example: :8020>
ECOM_PRODUCT_SERVICE_FRONTEND_URL=<ecom frontend url, example: http://127.0.0.1:8000>
ECOM_PRODUCT_SERVICE_ACCOUNT_SERVICE_URL=<ecom account service url, example: http://127.0.0.1:8010>
```

6. create postgresql databases with name same as in .env file
7. test server first with `go test ./...` to make sure server works fine
8. run server with `go run ./...`
