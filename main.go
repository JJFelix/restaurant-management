package main

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/JJFelix/restaurant_management/database"
	routes "github.com/JJFelix/restaurant_management/routes"
	middleware "github.com/JJFelix/restaurant_management/middleware"
	"go.mongodb.org/mongo-driver/mongo"

)


// mongodb
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main(){
	port := os.Getenv("PORT")

	if port ==""{
		port = "8000"
	}

	// using gin router
	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)


}