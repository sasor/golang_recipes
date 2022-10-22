package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	recipes []Recipe

	ctx    context.Context
	err    error
	client *mongo.Client
)

const (
	MongoUri        = "mongodb://recipes:recipes@localhost:27017/demo?authSource=admin"
	MongoDb         = "demo"
	MongoCollection = "recipes"
)

func init() {
	recipes = make([]Recipe, 0)
	//
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}
	log.Printf("Mongo database connected.")
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"published_at"`
}

// https://stackoverflow.com/questions/71907261/try-to-convert-json-to-map-for-golang-web-application
// https://golangdocs.com/json-with-golang
func ImportRecipesHandler(c *gin.Context) {
	recipes := make([]Recipe, 0)
	content, err := os.ReadFile("recipes.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = json.Unmarshal(content, &recipes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var items []interface{}
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	collection := client.Database(MongoDb).Collection(MongoCollection)
	result, err := collection.InsertMany(ctx, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"detail":  result.InsertedIDs,
	})
}

func NewRecipesHandler(c *gin.Context) {
	var recipe Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)

	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := -1

	for idx, r := range recipes {
		if r.ID == id {
			index = idx
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe Not Found",
		})
		return
	}

	recipe.ID = id // sin esto el ID del recipe se pierde al actualizar
	recipes[index] = recipe

	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1

	for idx, r := range recipes {
		if r.ID == id {
			index = idx
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}

	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusNoContent, nil)
}

func main() {
	router := gin.Default()
	router.GET("/import", ImportRecipesHandler)
	//
	router.POST("/recipes", NewRecipesHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	log.Fatal(router.Run())
}
