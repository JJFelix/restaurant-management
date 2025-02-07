package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/JJFelix/restaurant_management/database"
	"github.com/JJFelix/restaurant_management/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate = validator.New()


func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1{
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1{
			page = 10
		}

		// match and skip
		startIndex := (page-1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		// aggregation pipeline
		// https://chat.deepseek.com/a/chat/s/141f5677-1646-49c9-a5b0-b6543d69c823#:~:text=In%20MongoDB%2C%20aggregation%20pipelines%20are%20used%20to%20process%20and%20transform%20documents%20in%20a%20collection.%20The%20pipeline%20consists%20of%20multiple%20stages%2C%20each%20performing%20a%20specific%20operation%20on%20the%20documents.%20Let%27s%20break%20down%20the%20three%20stages%20you		matchStage

		matchStage := bson.D{{ "$match", bson.D{{}} }}
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}} }}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_count", 1},
					{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
				},
			},
		}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occirred while listning food items"})
			var allFoods []bson.M
			if err = result.All(ctx, &allFoods); err != nil{
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, allFoods[0])
		}
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		food := models.Food{}

		err := foodCollection.FindOne(ctx, bson.M{"food_id":foodId}).Decode(&food)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while fetching the food item"})
		}
		c.JSON(http.StatusOK, food)			
	}
}

func CreateFood() gin.HandlerFunc{
	return func (c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		food := models.Food{}	
		menu := models.Menu{}

		if err := c.BindJSON(&food); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		validationErr := validate.Struct(food)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}

		// find an existing menu
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()
		if err != nil{
			msg := fmt.Sprintf("Menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		num := toFixed(*food.Price, 2)
		food.Price = &num


		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil{
			msg := fmt.Sprintf("Food Item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int{
	return int(num + math.Copysign(0.5, num))
}


func toFixed(num float64, precision int) float64{
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}


func UpdateFood() gin.HandlerFunc{
	return func (c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		menu := models.Menu{}		
		food := models.Food{}

		foodId := c.Param("food_id")

		if err := c.BindJSON(&menu); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		var updateObj primitive.D

		if food.Name != nil{
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}

		if food.Price != nil{
			updateObj = append(updateObj, bson.E{"price", food.Price})
		}

		if food.Food_image != nil{
			updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
		}

		if food.Menu_id != nil{
			err := menuCollection.FindOne(ctx, bson.M{"menu_id":food.Menu_id}).Decode(&menu)
			defer cancel()
			if err != nil{
				msg := fmt.Sprintf("message: Menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"menu_id", food.Price})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"upated_at", food.Updated_at})

		upsert := true
		filter := bson.M{"food_id": foodId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			ctx, 
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil{
			msg := fmt.Sprint("food item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}