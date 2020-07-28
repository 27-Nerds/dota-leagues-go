package main

import (
	"context"
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

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func dbConnection() db.Interface {
	log.Printf("cs: %+v, %+v, %+v, %+v", viper.GetString(`database.url`), viper.GetString(`database.user`), viper.GetString(`database.pass`), viper.GetString(`database.name`))
	db, err := db.Connect(context.Background(), viper.GetString(`database.url`), viper.GetString(`database.user`), viper.GetString(`database.pass`), viper.GetString(`database.name`))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return db
}

func main() {
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
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD},
	}))

	// Routes
	e.Static("/", "./public")

	leaguesHandler := handler.NewLeaguesHandler(&leagueDetailsRepository)
	gamesHandler := handler.NewGameHandler(&gameRepository)
	delivery.NewLeaguesDelivery(e, &leaguesHandler, &gamesHandler)

	// Start server
	e.Logger.Fatal(e.Start(viper.GetString(`server.address`)))
}
