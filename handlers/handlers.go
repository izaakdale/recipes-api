package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/izaakdale/recipes-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipeHandler struct {
	collection  *mongo.Collection
	redisClient *redis.Client
}

func New(recipeCollection *mongo.Collection, usersCollection *mongo.Collection, rc *redis.Client) (*RecipeHandler, *AuthHandler) {
	return &RecipeHandler{recipeCollection, rc}, &AuthHandler{usersCollection}
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
func (h *RecipeHandler) GetRecipeHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	res := h.collection.FindOne(c.Request.Context(), bson.M{
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
	var recipe models.Recipe
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
func (h *RecipeHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	res, err := h.collection.DeleteOne(c.Request.Context(), bson.M{
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
func (h *RecipeHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe models.Recipe
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
	_, err = h.collection.UpdateOne(c.Request.Context(),
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
func (h *RecipeHandler) ListRecipesHandler(c *gin.Context) {
	res, err := h.redisClient.Get(c.Request.Context(), "recipes").Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("requesting data from mongo\n")
			cur, err := h.collection.Find(c.Request.Context(), bson.M{})
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					gin.H{"error": err.Error()})
				return
			}
			defer cur.Close(c.Request.Context())

			recipes := make([]models.Recipe, 0)
			for cur.Next(c.Request.Context()) {
				var recipe models.Recipe
				cur.Decode(&recipe)
				recipes = append(recipes, recipe)
			}

			data, _ := json.Marshal(recipes)
			h.redisClient.Set(c.Request.Context(), "recipes", string(data), 0)
			c.JSON(http.StatusOK, recipes)
			return
		} else {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(res), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
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
func (h *RecipeHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	res, err := h.collection.InsertOne(c.Request.Context(), recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error writing to the DB",
		})
		return
	}

	h.redisClient.Del(c.Request.Context(), "recipes")

	c.JSON(http.StatusCreated, gin.H{
		"id": res.InsertedID,
	})
}
