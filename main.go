package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.Default().Run()
}

type Recipe struct {
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"published_at"`
}
