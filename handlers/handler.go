package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/sasor/golang_recipes/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

type RecipesHandler struct {
	context    context.Context
	collection *mongo.Collection
	redis      *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redis *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		context:    ctx,
		collection: collection,
		redis:      redis,
	}
}

func (h *RecipesHandler) MongoCollection() *mongo.Collection {
	return h.collection
}

func (h *RecipesHandler) SharedContext() context.Context {
	return h.context
}

func (h *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	recipes := make([]models.Recipe, 0)
	result, err := h.redis.Get(h.context, "recipes").Result()
	if err == redis.Nil {
		log.Println("MONGO request")
		cursor, err := h.collection.Find(h.context, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		defer cursor.Close(h.context)

		for cursor.Next(h.context) {
			var recipe models.Recipe
			err := cursor.Decode(&recipe)
			if err != nil {
				log.Println(err.Error())
			}
			recipes = append(recipes, recipe)
		}
		marshal, _ := json.Marshal(recipes)
		h.redis.Set(h.context, "recipes", string(marshal), 0)
		c.JSON(http.StatusOK, recipes)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	} else {
		log.Println("REDIS request")
		err := json.Unmarshal([]byte(result), &recipes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, recipes)
}

func (h *RecipesHandler) RetrieveARecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var recipe models.Recipe
	redisKey := fmt.Sprintf("%s/%s", "recipe", id)
	result, err := h.redis.Get(h.context, redisKey).Result()
	if err == redis.Nil {
		log.Println("MONGO request")
		filter := bson.D{{"_id", objectID}}
		err = h.collection.FindOne(h.context, filter).Decode(&recipe)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		marshal, _ := json.Marshal(recipe)
		h.redis.Set(h.context, redisKey, string(marshal), 0)
		c.JSON(http.StatusOK, recipe)
		return
	} else if err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	} else {
		log.Println("REDIS request")
		err := json.Unmarshal([]byte(result), &recipe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, recipe)
}

func (h *RecipesHandler) NewRecipesHandler(c *gin.Context) {
	var recipe models.Recipe
	err := c.ShouldBindJSON(&recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()

	_, err = h.collection.InsertOne(h.context, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error on storing a recipe",
			"details": err.Error(),
		})
		return
	}

	h.redis.Del(h.context, "recipes")
	c.JSON(http.StatusOK, recipe)
}

func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid param",
			"details": err.Error(),
		})
		return
	}

	filter := bson.D{{"_id", objectId}}
	_, err = h.collection.DeleteOne(h.context, filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Some error",
			"details": err.Error(),
		})
		return
	}

	h.redis.Del(h.context, "recipes")
	c.JSON(http.StatusNoContent, nil)
}

func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
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

	_, err = h.collection.UpdateOne(h.context, filter, update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Recipe not updated",
			"details": err.Error(),
		})
		return
	}

	h.redis.Del(h.context, "recipes")
	c.JSON(http.StatusOK, recipe)
}
