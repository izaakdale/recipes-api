// Recipes API

// Recipes API.
// You can find out more about the API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin.
//
//	 Schemes: http
//	 BasePath: /
//	 Version: 1.0.0
//	 Host: localhost:8080
//	 Contact: izaakdaledev@gmail.com
//
//	 Consumes:
//	 - application/json
//
//	 Produces:
//	 - application/json
//
//	 Security:
//	 - basic
//
//	SecurityDefinitions:
//	basic:
//	  type: basic
//
// swagger:meta
package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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
	var err error
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to mongo db\n")

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	rh := handlers.New(context.Background(), collection)

	router := gin.Default()
	router.GET("/recipes", rh.ListRecipesHandler)
	router.POST("/recipes", rh.NewRecipeHandler)
	router.PUT("/recipes/:id", rh.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", rh.DeleteRecipeHandler)
	router.GET("/recipes/:id", rh.GetRecipeHandler)

	router.Run()
}
