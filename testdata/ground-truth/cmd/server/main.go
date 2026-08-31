package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"stripe-lite-mock/mock"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	store := mock.NewStore()

	r := gin.Default()
	mock.RegisterRoutes(r, store)

	log.Printf("Mock server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
