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
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, err := os.ReadFile("recipes.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(file), &recipes)
	if err != nil {
		panic(err)
	}
}

func main() {
	router := gin.Default()
	router.GET("/recipes", ListRecipesHandler)
	router.POST("/recipes", NewRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.Run()
}

// swagger:route GET /recipes/search recipe SearchRecipesHandler
// Search for a recipe with given tag
//
// parameters:
//   - +name: tag
//     in: query
//     description: recipe tag
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
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	var recipeList []Recipe

	for _, r := range recipes {
		for _, t := range r.Tags {
			if strings.EqualFold(t, tag) {
				recipeList = append(recipeList, r)
			}
		}
	}

	if len(recipeList) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no records found for that tag",
		})
		return
	}

	c.JSON(http.StatusOK, recipeList)
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

	idx := -1
	for i, r := range recipes {
		if r.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no records exist for that id",
		})
		return
	}
	recipes = append(recipes[:idx], recipes[idx+1:]...)
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

	idx := -1
	for i, r := range recipes {
		if r.ID == id {
			recipe.ID = r.ID
			idx = i
			break
		}
	}
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no records exist for that id",
		})
		return
	}

	recipes[idx] = recipe
	c.JSON(http.StatusAccepted, recipe)
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
	recipe.ID = uuid.NewString()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusCreated, recipe)
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"published_at"`
}
