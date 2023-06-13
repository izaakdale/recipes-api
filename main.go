// Recipes API

// Recipes API.
// You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin.
//
//		 Schemes: http
//		 BasePath: /
//		 Version: 1.0.0
//		 Host: api.recipes.io:8080
//		 Contact: izaakdaledev@gmail.com
//
//		 SecurityDefinitions:
//		 api_key:
//	    type: apiKey
//		   name: Authorization
//		   in: Header
//
//		 Consumes:
//		 - application/json
//
//		 Produces:
//		 - application/json
//
//		 Security:
//		 - basic
//
//		SecurityDefinitions:
//		basic:
//		  type: basic
//
// swagger:meta
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/izaakdale/recipes-api/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ()

func init() {
}

func main() {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URI"),
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping(ctx)
	log.Println(status)

	var err error
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to mongo db\n")

	recipeCollection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	usersCollection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	rh, ah := handlers.New(recipeCollection, usersCollection, redisClient)

	store, err := redisStore.NewStore(10, "tcp", os.Getenv("REDIS_URI"), "", []byte("secret"))
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodOptions},
		AllowHeaders:     []string{"ORIGIN"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(sessions.Sessions("recipe_api", store))

	router.GET("/recipes", rh.ListRecipesHandler)
	router.POST("/signin", ah.SignInHandler)
	router.POST("/refresh", ah.RefreshHandler)
	router.POST("/signup", ah.SignUpHandler)

	authorized := router.Group("/")
	authorized.Use(ah.AuthMiddleware())
	authorized.POST("/recipes", rh.NewRecipeHandler)
	authorized.PUT("/recipes/:id", rh.UpdateRecipeHandler)
	authorized.DELETE("/recipes/:id", rh.DeleteRecipeHandler)
	authorized.GET("/recipes/:id", rh.GetRecipeHandler)

	// router.RunTLS(":443", "certs/localhost.crt", "certs/localhost.key")
	router.Run()
}
