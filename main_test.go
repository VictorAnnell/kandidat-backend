package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

// Use a single instance of Validate, it caches struct info.
var validate *validator.Validate

// HTTP method constants
const (
	get = "GET"
	post = "POST"
	delete = "DELETE"
)

func TestMain(m *testing.M) {
	// Run test suite
	os.Exit(RunTests(m))
}

func RunTests(m *testing.M) int {
	setupConfig()
	gin.SetMode(gin.TestMode)

	dbPool = setupDBPool()
	router = setupRouter()

	defer dbPool.Close()

	validate = validator.New()

	return m.Run()
}

func TestPingRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGetCommunities(t *testing.T) {
	HTTPMethod := "GET"
	endpoint := "/communities"
	reqBody := ``
	expectedHTTPStatusCode := http.StatusOK
	expectedResponseStruct := []Community{}
	reqTester(t, HTTPMethod, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/communities", nil)
	router.ServeHTTP(w, req)

	var communityarray []Community

	err := json.Unmarshal(w.Body.Bytes(), &communityarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all Community structs in the array communityarray
	for _, community := range communityarray {
		err = validate.Struct(community)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUser(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1", nil)
	router.ServeHTTP(w, req)

	var user User

	err := json.Unmarshal(w.Body.Bytes(), &user)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate User struct
	err = validate.Struct(user)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/99999", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserCommunities(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1/communities", nil)
	router.ServeHTTP(w, req)

	var communityarray []Community

	err := json.Unmarshal(w.Body.Bytes(), &communityarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all Community structs in the array communityarray
	for _, community := range communityarray {
		err = validate.Struct(community)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/99999/communities", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserFollowers(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1/followers", nil)
	router.ServeHTTP(w, req)

	var userarray []User

	err := json.Unmarshal(w.Body.Bytes(), &userarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all User structs in the array userarray
	for _, user := range userarray {
		err = validate.Struct(user)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/99999/followers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProduct(t *testing.T) {
	// Test with valid product ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/products/1", nil)
	router.ServeHTTP(w, req)

	var product Product

	err := json.Unmarshal(w.Body.Bytes(), &product)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate Product struct
	err = validate.Struct(product)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid product ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/products/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserGetReviews(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1/reviews", nil)
	router.ServeHTTP(w, req)

	var reviewarray []Review

	err := json.Unmarshal(w.Body.Bytes(), &reviewarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all Review structs in the array reviewarray
	for _, review := range reviewarray {
		err = validate.Struct(review)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/99999/reviews", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserProducts(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/1/products", nil)
	router.ServeHTTP(w, req)

	var productarray []Product

	err := json.Unmarshal(w.Body.Bytes(), &productarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all Product structs in the array productarray
	for _, product := range productarray {
		err = validate.Struct(product)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/99999/products", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateProduct(t *testing.T) {
	// Test with valid JSON body and valid user ID
	endpoint := "/users/1/products"
	reqBody := `{"name": "Test Product", "service": false, "price": "2.00", "description": "Test Description"}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Product{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with invalid JSON body and valid user ID
	endpoint = "/users/1/products"
	reqBody = `{"wrong-field-name": "Test Product", "wrong-field-name2": false, "price": "abc", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = Product{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with valid JSON body and invalid user ID
	endpoint = "/users/99999/products"
	reqBody = `{"name": "Test Product", "service": false, "price": "1.00", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Product{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with invalid JSON body and invalid user ID
	endpoint = "/users/99999/products"
	reqBody = `{"invalid-field-name": "Test Product", "wrong-field-name2": false, "price": "abc", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Product{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)
}

func TestCreateReview(t *testing.T) {
	// Test with valid JSON body and valid user ID
	endpoint := "/users/1/reviews"
	reqBody := `{"rating": "5", "description": "Test Description", "reviewerid": "2"}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Review{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with invalid JSON body and valid user ID
	endpoint = "/users/1/reviews"
	reqBody = `{"rating": "this should be a number", "description": "Test Description", "reviewerid": "2"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = Review{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with valid JSON body and invalid user ID
	endpoint = "/users/99999/reviews"
	reqBody = `{"name": "Test Product", "service": false, "price": "1.00", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Review{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with invalid JSON body and invalid user ID
	endpoint = "/users/99999/reviews"
	reqBody = `{"rating": "this should be a number", "description": "Test Description", "reviewerid": "2"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Review{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)
}

func TestCreateUser(t *testing.T) {
	// Test with valid JSON body
	endpoint := "/users"
	reqBody := `{"name": "Test User", "phonenumber": "123456", "password": "a nice password"}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := User{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)

	// Test with invalid JSON body
	endpoint = "/users"
	reqBody = `{"invalid-field-name": "Test User", "phonenumber": "this should be a number", "password": "a nice password"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = User{}
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode, expectedResponseStruct)
}

// reqTester is a helper function for request testing
func reqTester(t *testing.T, httpMethod string, endpoint string, reqBody string, expectedHTTPStatusCode int, expectedResponseStruct any) {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(httpMethod, endpoint, strings.NewReader(reqBody))

	if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	router.ServeHTTP(w, req)

	err := json.Unmarshal(w.Body.Bytes(), &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	assert.Equal(t, expectedHTTPStatusCode, w.Code)
}
