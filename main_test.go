package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
	get  = "GET"
	post = "POST"
	del  = "DELETE"
	put  = "PUT"
)

func TestMain(m *testing.M) {
	// Run test suite
	os.Exit(RunTests(m))
}

func RunTests(m *testing.M) int {
	setupConfig()

	if _, modeisset := os.LookupEnv("GIN_MODE"); !modeisset {
		gin.SetMode(gin.TestMode)
	}

	dbPool = setupDBPool()
	router = setupRouter()

	defer dbPool.Close()

	validate = validator.New()

	return m.Run()
}

func TestPingRoute(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGetCommunities(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/communities", nil)
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

func TestGetProducts(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/products", nil)
	router.ServeHTTP(w, req)

	var productarray []Product

	err := json.Unmarshal(w.Body.Bytes(), &productarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all Community structs in the array communityarray
	for _, product := range productarray {
		err = validate.Struct(product)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUser(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1", nil)
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
	req, _ = http.NewRequest(get, "/users/99999", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserCommunities(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/communities", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/communities", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test with valid user ID and URL paramater joined=false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(get, "/users/1/communities?joined=false", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &communityarray)
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

	// Test with invalid user ID and URL paramater joined=false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(get, "/users/99999/communities?joined=false", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserFollowers(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/followers", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/followers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProduct(t *testing.T) {
	// Test with valid product ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/products/1", nil)
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
	req, _ = http.NewRequest(get, "/products/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserReviews(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/reviews", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/reviews", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserProducts(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/products", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/products", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateProduct(t *testing.T) {
	// Test with valid JSON body and valid user ID
	endpoint := "/users/1/products"
	reqBody := `{"name": "Test Product", "category": "Test Category", "service": true, "price": 2, "description": "Test Description"}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Product{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Test with invalid JSON body and valid user ID
	endpoint = "/users/1/products"
	reqBody = `{"wrong-field-name": "Test Product", "wrong-field-name2": false, "price": "abc", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = Product{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with valid JSON body and invalid user ID
	endpoint = "/users/99999/products"
	reqBody = `{"name": "Test Product", "category": "Test Category", "service": true, "price": 5, "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Product{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid JSON body and invalid user ID
	endpoint = "/users/99999/products"
	reqBody = `{"invalid-field-name": "Test Product", "wrong-field-name2": false, "price": "abc", "description": "Test Description"}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Product{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}
func TestCreateReview(t *testing.T) {
	c := context.Background()
	// delete review to test
	query := "DELETE FROM Review where fk_owner_id = $1 and fk_reviewer_id = $2"
	_, err := dbPool.Exec(c, query, 1, 2)

	if err != nil {
		fmt.Println(err)
	}
	// Test with valid JSON body and valid user ID
	endpoint := "/users/1/reviews"
	reqBody := `{"rating": 5, "description": "Test Description", "reviewer_id": 2}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Review{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err = json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Test with invalid JSON body and valid user ID
	endpoint = "/users/1/reviews"
	reqBody = `{"rating": "this should be a number", "description": "Test Description", "reviewer_id": 2}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = Review{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with valid JSON body and invalid user ID

	endpoint = "/users/99999/reviews"
	reqBody = `{"rating": 1, "description": "Test Description", "reviewer_id": 2}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Review{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid JSON body and invalid user ID
	endpoint = "/users/99999/reviews"
	reqBody = `{"rating": "this should be a number", "description": "Test Description", "reviewer_id": 2}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Review{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestCreateUser(t *testing.T) {
	// Test with valid JSON body
	endpoint := "/users" //nolint:goconst // No const is better for readability
	reqBody := `{"name": "Test User", "phone_number": "+12027485281", "password": "a nice password", "Business" : false}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := User{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Delete the created test user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(del, "/users/"+strconv.Itoa(expectedResponseStruct.UserID), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		fmt.Println("Notice: the created test user could not be deleted.")
		fmt.Println(w.Body.String())
	}

	// Test with invalid JSON body
	endpoint = "/users"
	reqBody = `{"invalid-field-name": "Test User", "phone_number": "this should be a number", "password": "a nice password", "business": false}`
	expectedHTTPStatusCode = http.StatusBadRequest
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestJoinCommunity(t *testing.T) {
	// Test with valid JSON body and valid user ID
	endpoint := "/users/1/communities"
	reqBody := `{"community_id": 2}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Community{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Test with invalid JSON body and valid user ID
	endpoint = "/users/1/communities"
	reqBody = `{"invalid-field-name": 1}`
	expectedHTTPStatusCode = http.StatusBadRequest
	expectedResponseStruct = Community{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with valid JSON body and invalid user ID
	endpoint = "/users/99999/communities"
	reqBody = `{"community_id": 3}`
	expectedHTTPStatusCode = http.StatusNotFound
	expectedResponseStruct = Community{}

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestDeleteUser(t *testing.T) {
	// Create user to delete
	reqBody := `{"name": "Test User", "phone_number": "+12999999999", "password": "a nice password", "business": true}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(post, "/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		fmt.Println(w.Body.String())
		fmt.Println(w.Code)
		t.Error("Failed to create test user to delete")

		return
	}

	var testUser User

	err := json.Unmarshal(w.Body.Bytes(), &testUser)
	if err != nil {
		t.Errorf("Error unmarshalling json of test user to be deleted: %v", err)
		return
	}

	// Test with valid user ID
	endpoint := "/users/" + strconv.Itoa(testUser.UserID)
	expectedHTTPStatusCode := http.StatusNoContent
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with invalid user ID
	endpoint = "/users/99999"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)
}

func TestGetPinnedProducts(t *testing.T) {
	// Test with valid user ID
	endpoint := "/users/1/pinned" //nolint:goconst // No const is better for readability
	expectedHTTPStatusCode := http.StatusOK
	expectedResponseStructSlice := []Product{}
	bodyBytes := reqTester(t, get, endpoint, "", expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStructSlice)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all product structs in the slice
	for _, product := range expectedResponseStructSlice {
		err = validate.Struct(product)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	// Test with invalid user ID
	endpoint = "/users/99999/pinned"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, get, endpoint, "", expectedHTTPStatusCode)
}

func TestAddPinnedProduct(t *testing.T) {
	// Test with valid user ID and valid product ID
	endpoint := "/users/1/pinned"
	reqBody := `{"product_id": 2}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Product{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Delete the created test pinned product
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(del, "/users/1/pinned/2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		fmt.Println("Notice: the created test pinned product could not be deleted.")
		fmt.Println(w.Body.String())
	}

	// Test with invalid user ID
	endpoint = "/users/99999/pinned"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid product ID
	endpoint = "/users/1/pinned"
	reqBody = `{"product_id": 99999}`
	expectedHTTPStatusCode = http.StatusBadRequest

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestDeletePinnedProduct(t *testing.T) {
	// Add pinned product to be deleted
	reqBody := `{"product_id": 2}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(post, "/users/1/pinned", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		fmt.Println(w.Body.String())
		fmt.Println(w.Code)
		t.Error("Failed to create test pinned product to delete")

		return
	}

	// Test with valid but wrong user ID and valid product ID
	endpoint := "/users/2/pinned/2"
	expectedHTTPStatusCode := http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with valid user ID and valid product ID
	endpoint = "/users/1/pinned/2"
	expectedHTTPStatusCode = http.StatusNoContent
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with invalid user ID
	endpoint = "/users/99999/pinned/2"
	expectedHTTPStatusCode = http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with invalid product ID
	endpoint = "/users/1/pinned/99999"
	expectedHTTPStatusCode = http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)
}

func TestGetUserFollowing(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/following", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/following", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserFollowingProducts(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/following/products", nil)
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
	req, _ = http.NewRequest(get, "/users/99999/following/products", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// reqTester is a helper function for request testing
func reqTester(t *testing.T, httpMethod string, endpoint string, reqBody string, expectedHTTPStatusCode int) []byte {
	t.Helper()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(httpMethod, endpoint, strings.NewReader(reqBody))

	if httpMethod == post || httpMethod == "PUT" || httpMethod == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	router.ServeHTTP(w, req)

	if !assert.Equal(t, expectedHTTPStatusCode, w.Code) {
		fmt.Println("Returned response: " + w.Body.String())
	}

	return w.Body.Bytes()
}

func TestUpdateUser(t *testing.T) {
	// Test with valid JSON body
	endpoint := "/users/2"
	reqBody := `{"name": "Test User", "phone_number": "+12027485281", "password": "a nice password", "business": false}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := User{}
	bodyBytes := reqTester(t, put, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Test with invalid JSON body
	endpoint = "/users"
	reqBody = `{"invalid-field-name": "Test User", "phone_number": "this should be a number", "password": "a nice password"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Change back the original username
	endpoint = "/users/2"
	reqBody = `{"name": "Victor", "phone_number": "+12027455483", "password": "lorem ipsum", "rating": 4, "business": false}`
	expectedHTTPStatusCode = http.StatusCreated
	expectedResponseStruct = User{}

	reqTester(t, put, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(put, "/users/99999", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLogin(t *testing.T) {
	// Test with valid credentials
	endpoint := "/login" //nolint:goconst // No const is better for readability
	reqBody := `{"phone_number": "+12027455483", "password": "lorem ipsum"}`
	expectedHTTPStatusCode := http.StatusCreated
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid phone number
	endpoint = "/login"
	reqBody = `{"phone_number": "+99999999999", "password": "lorem ipsum"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid password
	endpoint = "/login"
	reqBody = `{"phone_number": "+12027455483", "password": "wrong password"}`
	expectedHTTPStatusCode = http.StatusBadRequest
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestGetBuyingProducts(t *testing.T) {
	// Test with valid user ID
	endpoint := "/users/2/pinned" //nolint:goconst // No const is better for readability
	expectedHTTPStatusCode := http.StatusOK
	expectedResponseStructSlice := []Product{}
	bodyBytes := reqTester(t, get, endpoint, "", expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStructSlice)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all product structs in the slice
	for _, product := range expectedResponseStructSlice {
		err = validate.Struct(product)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	// Test with invalid user ID
	endpoint = "/users/99999/pinned"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, get, endpoint, "", expectedHTTPStatusCode)
}

func TestAddBuyingProduct(t *testing.T) {
	// Test with valid user ID and valid product ID
	endpoint := "/users/2/buying"
	reqBody := `{"product_id": 2}`
	expectedHTTPStatusCode := http.StatusCreated
	expectedResponseStruct := Product{}
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Delete the created test buying product
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(del, "/users/2/buying/2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		fmt.Println("Notice: the created test buying product could not be deleted.")
		fmt.Println(w.Body.String())
	}

	// Test with invalid user ID
	endpoint = "/users/99999/buying"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid product ID
	endpoint = "/users/2/buying"
	reqBody = `{"product_id": 99999}`
	expectedHTTPStatusCode = http.StatusBadRequest

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}

func TestDeleteBuyingProduct(t *testing.T) {
	// Add buying product to be deleted
	reqBody := `{"product_id": 2}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(post, "/users/2/buying", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		fmt.Println(w.Body.String())
		fmt.Println(w.Code)
		t.Error("Failed to create test buying product to delete")

		return
	}

	// Test with valid but wrong user ID and valid product ID
	endpoint := "/users/1/buying/2"
	expectedHTTPStatusCode := http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with valid user ID and valid product ID
	endpoint = "/users/2/buying/2"
	expectedHTTPStatusCode = http.StatusNoContent
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with invalid user ID
	endpoint = "/users/99999/buying/2"
	expectedHTTPStatusCode = http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)

	// Test with invalid product ID
	endpoint = "/users/2/buying/99999"
	expectedHTTPStatusCode = http.StatusNotFound
	reqTester(t, del, endpoint, "", expectedHTTPStatusCode)
}

func TestGetUserChats(t *testing.T) {
	// Test with valid user ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(get, "/users/1/chats", nil)
	router.ServeHTTP(w, req)

	var chatarray []Chat

	err := json.Unmarshal(w.Body.Bytes(), &chatarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all User structs in the array userarray
	for _, user := range chatarray {
		err = validate.Struct(user)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with valid alt user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(get, "/users/2/chats", nil)
	router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &chatarray)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate all User structs in the array userarray
	for _, user := range chatarray {
		err = validate.Struct(user)
		if err != nil {
			t.Errorf("Error validating struct: %v", err)
		}
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid user ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(get, "/users/99999/chats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateAndDeleteChat(t *testing.T) {
	// Test delete chat
	endpoint := "/users/1/chats/2"
	reqBody := ``
	expectedHTTPStatusCode := http.StatusNoContent
	expectedResponseStruct := Chat{}
	reqTester(t, del, endpoint, reqBody, expectedHTTPStatusCode)

	// Test create same chat
	endpoint = "/users/2/chats"
	reqBody = `{"user_id": 1}`
	expectedHTTPStatusCode = http.StatusCreated
	bodyBytes := reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test decoding of JSON response body
	err := json.Unmarshal(bodyBytes, &expectedResponseStruct)
	if err != nil {
		t.Errorf("Error unmarshalling json: %v", err)
	}

	// Validate struct
	err = validate.Struct(expectedResponseStruct)
	if err != nil {
		t.Errorf("Error validating struct: %v", err)
	}

	// Test delete and create again
	endpoint = "/users/2/chats/1"
	expectedHTTPStatusCode = http.StatusNoContent
	reqTester(t, del, endpoint, reqBody, expectedHTTPStatusCode)

	endpoint = "/users/1/chats"
	reqBody = `{"user_id": 2}`
	expectedHTTPStatusCode = http.StatusCreated
	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid user ID
	endpoint = "/users/99999/chats"
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)

	// Test with invalid product ID
	endpoint = "/users/2/chats"
	reqBody = `{"user_id": 99999}`
	expectedHTTPStatusCode = http.StatusNotFound

	reqTester(t, post, endpoint, reqBody, expectedHTTPStatusCode)
}
