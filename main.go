package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}
	log.Printf("Mongo database connected.")
}

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"published_at" bson:"published_at"`
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
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	collection := client.Database(MongoDb).Collection(MongoCollection)
	_, err = collection.InsertOne(ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error on storing a recipe",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	collection := client.Database(MongoDb).Collection(MongoCollection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	defer cursor.Close(ctx)

	recipes := make([]Recipe, 0)
	for cursor.Next(ctx) {
		var recipe Recipe
		err := cursor.Decode(&recipe)
		if err != nil {
			log.Println(err.Error())
		}
		recipes = append(recipes, recipe)
	}

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

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid param",
			"details": err.Error(),
		})
		return
	}

	filter := bson.D{{"_id", objectId}}
	update := bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}}
	collection := client.Database(MongoDb).Collection(MongoCollection)
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Recipe not updated",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid param",
			"details": err.Error(),
		})
		return
	}

	collection := client.Database(MongoDb).Collection(MongoCollection)
	filter := bson.D{{"_id", objectId}}
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Some error",
			"details": err.Error(),
		})
		return
	}

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
