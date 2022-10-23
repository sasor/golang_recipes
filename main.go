package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sasor/golang_recipes/handlers"
	"github.com/sasor/golang_recipes/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

const (
	MongoUri        = "mongodb://recipes:recipes@localhost:27017/demo?authSource=admin"
	MongoDb         = "demo"
	MongoCollection = "recipes"
)

var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	collection := client.Database(MongoDb).Collection(MongoCollection)
	recipesHandler = handlers.NewRecipesHandler(ctx, collection)
	log.Printf("Mongo database connected.")
}

// https://stackoverflow.com/questions/71907261/try-to-convert-json-to-map-for-golang-web-application
// https://golangdocs.com/json-with-golang
func ImportRecipesHandler(c *gin.Context) {
	recipes := make([]models.Recipe, 0)
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
		recipe.ID = primitive.NewObjectID()
		items = append(items, recipe)
	}

	result, err := recipesHandler.MongoCollection().InsertMany(recipesHandler.SharedContext(), items)
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

func main() {
	router := gin.Default()
	router.GET("/import", ImportRecipesHandler)
	//
	router.POST("/recipes", recipesHandler.NewRecipesHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	log.Fatal(router.Run())
}
