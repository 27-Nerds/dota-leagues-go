package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"dota_league/db"
	"dota_league/delivery"
	"dota_league/handler"
	"dota_league/repository"
	"dota_league/worker"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

// ENVIRONMENT - Global variable that stores current envorinment
var ENVIRONMENT = "development"

// GetConfigStr add current env to viper config query
func GetConfigStr(key string) string {
	key = fmt.Sprintf("%s.%s", ENVIRONMENT, key)

	return viper.GetString(key)
}

func init() {
	//set default values
	viper.SetDefault("cors.origin", "*")

	//read the cofig file
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func dbConnection() db.Interface {
	log.Printf("cs: %+v, %+v, %+v, %+v",
		GetConfigStr(`database.url`),
		GetConfigStr(`database.user`),
		GetConfigStr(`database.pass`),
		GetConfigStr(`database.name`))

	db, err := db.Connect(context.Background(),
		GetConfigStr(`database.url`),
		GetConfigStr(`database.user`),
		GetConfigStr(`database.pass`),
		GetConfigStr(`database.name`))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return db
}

func main() {

	prodFlag := flag.Bool("production", false, "run in production mode")
	flag.Parse()

	if *prodFlag {
		ENVIRONMENT = "production"
	}

	log.Printf("Current Environment: %s", ENVIRONMENT)

	db := dbConnection()
	leagueRepository := repository.NewLeagueRepository(&db)
	leagueDetailsRepository := repository.NewLeagueDetailsRepository(&db)
	gameRepository := repository.NewGameRepository(&db)
	_ = worker.NewDataLoader(&leagueRepository, &leagueDetailsRepository, &gameRepository)

	//----------------
	//START WEB SERVER
	//----------------

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{GetConfigStr(`cors.origin`)},
		AllowMethods: []string{echo.GET, echo.HEAD},
	}))

	// Routes
	e.Static("/", "./public")

	leaguesHandler := handler.NewLeaguesHandler(&leagueDetailsRepository)
	gamesHandler := handler.NewGameHandler(&gameRepository)
	delivery.NewLeaguesDelivery(e, &leaguesHandler, &gamesHandler)

	// Start server
	e.Logger.Fatal(e.Start(GetConfigStr(`server.address`)))
}
