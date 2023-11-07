package main

import (
	"fmt"
	"log"
	controller "web-scrap/Controller"
	"web-scrap/model"

	_ "github.com/lib/pq"
)

func main() {

	//load env
	err := model.LoadEnv()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := model.ConnectToDB()
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	scraperController := controller.NewScraperController(db)

	// Start scraping with initial counter value of 0
	message := scraperController.Scrape(db, 0)

	fmt.Println(message)
}
