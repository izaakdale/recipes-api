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
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func init() {
	ctx := context.Background()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to mongo db\n")

	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
}

func main() {
	router := gin.Default()
	router.GET("/recipes", ListRecipesHandler)
	router.POST("/recipes", NewRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/:id", GetRecipeHandler)
	router.Run()
}

// swagger:route GET /recipes/{id} recipe GetRecipeHandler
// Search for a recipe with given tag
//
// parameters:
//   - +name: id
//     in: path
//     description: recipe id
//     required: true
//     type: string
//
// Consumes:
//   - application/json
//
// Responses:
//
//	200: description:Success
//	404: description:No Records
func GetRecipeHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	res := collection.FindOne(c.Request.Context(), bson.M{
		"_id": objectID,
	})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound,
				gin.H{"message": "no records exist for that ID"})
			return
		}
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	var recipe Recipe
	err = res.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// swagger:route DELETE /recipes/{id} recipe DeleteRecipeHandler
// Delete a recipe
//
// parameters:
//   - +name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// Responses:
//
//	200: description:Success
//	404: description:No Records
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	res, err := collection.DeleteOne(c.Request.Context(), bson.M{
		"_id": objectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no records found for that ID",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "removed",
	})
}

// swagger:route PUT /recipes/{id} recipe UpdateRecipeHandler
// Update a recipe
//
// parameters:
//   - +name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// Consumes:
//   - application/json
//
// Responses:
//
//	200: description:Success
//	400: description:Invalid Input
//	404: description:No Records
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	_, err = collection.UpdateOne(c.Request.Context(),
		bson.M{
			"_id": objectID,
		}, bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: recipe.Name},
				{Key: "instructions", Value: recipe.Instructions},
				{Key: "ingredients", Value: recipe.Ingredients},
				{Key: "tags", Value: recipe.Tags},
			}},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Recipe updated",
	})
}

// swagger:route GET /recipes recipe ListRecipesHandler
// Returns a list of recipes.
//
// produces:
// - application/json
//
// responses:
//
//	200: description:Success
func ListRecipesHandler(c *gin.Context) {
	cur, err := collection.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(c.Request.Context())

	recipes := make([]Recipe, 0)
	for cur.Next(c.Request.Context()) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

// swagger:route POST /recipes recipe NewRecipeHandler
// Saves a new recipe
//
// produces:
// - application/json
//
// responses:
//
//	200: description:Success
//	400: description:Invalid Input
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	res, err := collection.InsertOne(c.Request.Context(), recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error writing to the DB",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": res.InsertedID,
	})
}

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"published_at" bson:"published_at"`
}
