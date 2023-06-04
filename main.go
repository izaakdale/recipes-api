package main

import (
	"encoding/json"
	"net/http"
	"os"
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
	router.Run()
}

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

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

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
