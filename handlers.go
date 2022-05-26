package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// ping returns a simple string to test the server is running.
func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func addPinnedProduct(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	type pinnedProduct struct {
		ProductID int `json:"product_id" binding:"required" db:"fk_product_id"`
	}

	var product pinnedProduct

	err := c.Bind(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkIfProductExist(c, strconv.Itoa(product.ProductID)) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product does not exist"})
		return
	}

	query := "INSERT INTO Pinned_Product (fk_product_id, fk_user_id) VALUES($1,$2) RETURNING fk_product_id"
	err = pgxscan.Get(c, dbPool, &product, query, product.ProductID, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, product)
}

// Get the products that userid has pinned
func getPinnedProducts(c *gin.Context) {
	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var pinnedProducts []*Product

	query := "SELECT * from Product WHERE product_id IN (SELECT fk_product_id FROM Pinned_Product WHERE fk_user_id = $1)"
	err := pgxscan.Select(c, dbPool, &pinnedProducts, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, pinnedProducts)
}

func deletePinnedProduct(c *gin.Context) {
	userID := c.Param("user_id")
	productID := c.Param("product_id")

	if checkIfUserExist(c, userID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	if checkIfProductExist(c, productID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product does not exist"})
		return
	}

	// Check if user has pinned product
	var pinnedProductID int

	query := "SELECT fk_product_id FROM Pinned_Product WHERE fk_user_id = $1 AND fk_product_id = $2"
	err := pgxscan.Get(c, dbPool, &pinnedProductID, query, userID, productID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User has not pinned the product"})
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	query = "DELETE FROM Pinned_Product WHERE fk_user_id = $1 AND fk_product_id = $2"
	_, err = dbPool.Exec(c, query, userID, productID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusNoContent, gin.H{"deleted": productID})
}

// Gives you all products that are owned by userId
func getUserProducts(c *gin.Context) {
	user := c.Param("user_id")
	owned := c.DefaultQuery("owned", "true")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var products []*Product

	var query string
	if owned == "false" {
		query = "SELECT * from Product WHERE fk_user_id != $1"
	} else {
		query = "SELECT * from Product WHERE fk_user_id = $1"
	}

	err := pgxscan.Select(c, dbPool, &products, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, products)
}

// Adds a product to the userID
func createProduct(c *gin.Context) {
	var product Product

	userID := c.Param("user_id")
	if checkIfUserExist(c, userID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	err := c.Bind(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encode picture to base64
	product.Picture = []byte(base64.StdEncoding.EncodeToString(product.Picture))

	query := "INSERT INTO Product(name,service,price,description,picture,fk_user_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING *"
	err = pgxscan.Get(c, dbPool, &product, query, product.Name, product.Service, product.Price, product.Description, product.Picture, userID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, product)
}

func createReview(c *gin.Context) {
	var review Review

	owner := c.Param("user_id")
	if checkIfUserExist(c, owner) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	err := c.Bind(&review)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkIfUserExist(c, strconv.Itoa(review.ReviewerID)) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	if checkForDupReview(c, review.ReviewerID, owner) == true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already left a review on this user"})

		return
	}

	query := "INSERT INTO Review(rating,content, fk_reviewer_id, fk_owner_id) VALUES($1,$2, $3, $4) RETURNING *"

	err = pgxscan.Get(c, dbPool, &review, query, review.Rating, review.Content, review.ReviewerID, owner)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, review)

	query = "UPDATE Users SET rating = (SELECT AVG(rating) FROM Review WHERE fk_owner_id = $1) WHERE user_id = $1"

	_, er := dbPool.Exec(c, query, review.OwnerID)
	if er != nil {
		fmt.Println(er)
		c.JSON(http.StatusInternalServerError, review)
	}
}

func joinCommunity(c *gin.Context) {
	var userCommunity UserCommunity

	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	err := c.Bind(&userCommunity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO User_Community(fk_user_id, fk_community_id) VALUES($1, $2) RETURNING fk_user_id, fk_community_id"
	err = pgxscan.Get(c, dbPool, &userCommunity, query, user, userCommunity.CommunityID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, userCommunity)
}

func getUserReviews(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	query := "SELECT * from Review WHERE fk_owner_id = $1"

	var reviews []*Review

	err := pgxscan.Select(c, dbPool, &reviews, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, reviews)
}

func getCommunities(c *gin.Context) {
	var communities []*Community

	query := "SELECT * FROM Community"
	err := pgxscan.Select(c, dbPool, &communities, query)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, communities)
}

func getProducts(c *gin.Context) {
	var products []*Product

	query := "SELECT * FROM Product"
	err := pgxscan.Select(c, dbPool, &products, query)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, products)
}

// getUserCommunities returns all communities the user is in.

func getUserCommunities(c *gin.Context) {
	user := c.Param("user_id")
	joined := c.DefaultQuery("joined", "true")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var query string

	var communities []*Community

	if joined == "false" {
		query = " SELECT * from Community WHERE community_id NOT IN (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	} else {
		query = " SELECT * from Community WHERE community_id IN (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	}

	err := pgxscan.Select(c, dbPool, &communities, query, user)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, communities)
}

// getUser returns the user with the given id.
func getUser(c *gin.Context) {
	var result User

	user := c.Param("user_id")
	query := "SELECT * from Users WHERE user_id = $1"

	err := pgxscan.Get(c, dbPool, &result, query, user)
	if err != nil {
		if err.Error() == ErrNoRows {
			c.Status(http.StatusNotFound)
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, result)
}

// getProduct returns the product with the given id.
func getProduct(c *gin.Context) {
	var result Product

	productID := c.Param("product_id")
	query := "SELECT * FROM Product WHERE product_id = $1"

	err := pgxscan.Get(c, dbPool, &result, query, productID)
	if err != nil {
		if err.Error() == ErrNoRows {
			c.Status(http.StatusNotFound)
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, result)
}

// getUserFollowers returns all users that follow the user with the given id.
func getUserFollowers(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var followers []*User

	query := ` SELECT * FROM Users WHERE user_id IN (SELECT fk_followed_id FROM User_Followers WHERE fk_user_id=$1)`

	err := pgxscan.Select(c, dbPool, &followers, query, user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, followers)
}

// getUserIsFollowing returns all users that the user with the given id is following.
func getUserIsFollowing(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var followers []*User

	query := `SELECT * FROM Users WHERE user_id IN (SELECT fk_user_id FROM User_Followers WHERE fk_followed_id=$1)`
	err := pgxscan.Select(c, dbPool, &followers, query, user)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, followers)
}

func getFollowingUsersProducts(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var products []*Product

	query := `SELECT * FROM Product WHERE fk_user_id in (SELECT user_id FROM Users WHERE user_id IN (SELECT fk_user_id FROM User_Followers WHERE fk_followed_id=$1))`
	err := pgxscan.Select(c, dbPool, &products, query, user)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, products)
}

// createUser creates a new user.
func createUser(c *gin.Context) {
	var user User

	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encode picture to base64
	user.Picture = []byte(base64.StdEncoding.EncodeToString(user.Picture))

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = string(hashedPassword)
	query := "INSERT INTO Users(name, phone_number, password, picture, business) VALUES($1,$2, $3, $4, $5) RETURNING *"
	err = pgxscan.Get(c, dbPool, &user, query, user.Name, user.PhoneNumber, user.Password, user.Picture, user.Business)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, user)
}

func deleteUser(c *gin.Context) {
	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	query := "UPDATE Review SET fk_reviewer_id = 0 WHERE fk_reviewer_id = $1"
	_, err := dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	query = "DELETE FROM Users where user_id = $1"
	_, err = dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusNoContent, gin.H{"deleted": user})
}

// login logs in the user with the given credentials.
func login(c *gin.Context) {
	type LoginUser struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	var loginUser LoginUser

	var response struct {
		ID    int
		Token string
	}

	var user User

	if err := c.Bind(&loginUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "SELECT password, user_id FROM Users where phone_number = $1"
	err := pgxscan.Get(c, dbPool, &user, query, loginUser.PhoneNumber)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect phone number"})
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	// Check if password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	response.ID = user.UserID
	response.Token = tokenString

	c.JSON(http.StatusCreated, response)
}

func updateUser(c *gin.Context) {
	var user User

	userid := c.Param("user_id")

	if checkIfUserExist(c, userid) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	if err := c.Bind(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = string(hashedPassword)

	query := "UPDATE Users SET name = $2, phone_number = $3, password = $4, picture = $5, rating = $6 WHERE user_id = $1 RETURNING *"
	err = pgxscan.Get(c, dbPool, &user, query, userid, user.Name, user.PhoneNumber, user.Password, user.Picture, user.Rating)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		fmt.Println(err)

		return
	}

	c.JSON(http.StatusCreated, user)
}

// checkIfUserExist is a helper function that checks if a user with the given ID exists in the database.
func checkIfUserExist(c *gin.Context, userID string) bool {
	query := "SELECT user_id from Users WHERE user_id = $1"

	var result User

	err := pgxscan.Get(c, dbPool, &result, query, userID)
	if err != nil {
		return false
	}

	return true
}

// checkIfProductExist is a helper function that checks if a product with the given ID exists in the database.
func checkIfProductExist(c *gin.Context, productID string) bool {
	query := "SELECT product_id from Product WHERE product_id = $1"

	var result Product

	err := pgxscan.Get(c, dbPool, &result, query, productID)
	if err != nil {
		return false
	}

	return true
}

// checkForDupReview is a helper function that checks if the given reviewer already has a review for the given owner
func checkForDupReview(c *gin.Context, reviewer int, owner string) bool {
	query := "SELECT review_id FROM Review WHERE fk_reviewer_id = $1 AND fk_owner_id = $2"

	var result Review

	err := pgxscan.Get(c, dbPool, &result, query, reviewer, owner)
	if err != nil {
		return false
	}

	return true
}

func checkForDupFollow(c *gin.Context, followed int, follower string) bool {
	query := "SELECT user_followers_id FROM User_Followers WHERE fk_user_id = $1 AND fk_followed_id = $2"

	var result Follow
	err := pgxscan.Get(c, dbPool, &result, query, follower, followed)

	if err != nil {
		return false
	}

	return true
}

func createFollow(c *gin.Context) {
	follower := c.Param("user_id")

	var follow Follow

	if err := c.BindJSON(&follow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkForDupFollow(c, follow.Followed, follower) == true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already follow this person"})

		return
	}

	query := "INSERT INTO User_Followers(fk_user_id, fk_followed_id) VALUES($1,$2)"
	_, err := dbPool.Exec(c, query, follower, follow.Followed)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, true)
}
