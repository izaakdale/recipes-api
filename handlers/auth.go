package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"hash"
	"net/http"
	"os"
	"time"

	"github.com/auth0-community/go-auth0"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/izaakdale/recipes-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/square/go-jose.v2"
)

var h hash.Hash

type AuthHandler struct {
	collection *mongo.Collection
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func init() {
	h = sha256.New()
}

// swagger:route POST /signin auth SignInHandler
// Login with username and password
//
// Produces:
//   - application/json
//
// Consumes:
//   - application/json
//
// Responses:
//
//	200: description:Success
//	401: description:Invalid credentials
func (a *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	cur := a.collection.FindOne(c.Request.Context(), bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})
	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	url := "https://" + os.Getenv("AUTH0_DOMAIN") + "/oauth/token"
	var payload = struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Audience     string `json:"audience"`
		GrantType    string `json:"grant_type"`
	}{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Audience:     os.Getenv("AUTH0_API_IDENTIFIER"),
		GrantType:    "client_credentials",
	}

	payBytes, _ := json.Marshal(payload)
	buf := bytes.NewBuffer(payBytes)
	req, _ := http.NewRequest(http.MethodPost, url, buf)
	req.Header.Add("content-type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	respPayload := make(map[string]any)
	json.NewDecoder(res.Body).Decode(&respPayload)

	c.JSON(http.StatusOK, respPayload)
}

// swagger:route POST /refresh auth RefreshHandler
// Get new token in exchange for an old one
//
// Produces:
//   - application/json
//
// Responses:
//
//		200: description:Success
//	    400: description:Token does not need refreshing
//		401: description:Invalid credentials
func (a *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	out := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}

	c.JSON(http.StatusOK, out)
}

func (a *AuthHandler) SignUpHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res, err := a.collection.InsertOne(c.Request.Context(), bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"user_id": res.InsertedID,
	})
}

func (a *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		var authDomain = "https://" + os.Getenv("AUTH0_DOMAIN") + "/"
		client := auth0.NewJWKClient(auth0.JWKClientOptions{
			URI: authDomain + ".well-known/jwks.json",
		}, nil)

		configuration := auth0.NewConfiguration(client, []string{os.Getenv("AUTH0_API_IDENTIFIER")}, authDomain, jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		_, err := validator.ValidateRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized,
				gin.H{"message": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}
