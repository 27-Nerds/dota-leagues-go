package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"dota_league/api"
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

// GetConfigStr add current env to viper config string query
func GetConfigStr(key string) string {
	key = fmt.Sprintf("%s.%s", ENVIRONMENT, key)

	return viper.GetString(key)
}

// GetConfigInt add current env to viper config int query
func GetConfigInt(key string) int {
	key = fmt.Sprintf("%s.%s", ENVIRONMENT, key)

	return viper.GetInt(key)
}

// GetConfigFloat add current env to viper config float query
func GetConfigFloat(key string) float64 {
	key = fmt.Sprintf("%s.%s", ENVIRONMENT, key)

	if v, ok := viper.Get(key).(float64); ok {
		return v
	}

	return 1.5
}

func init() {
	//set default values
	viper.SetDefault("cors.origin", "*")
	viper.SetDefault("valve.rps", 1.5)

	//read the cofig file
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func dbConnection() *db.ArangoDB {
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
	leagueRepository := repository.NewLeagueRepository(db)
	leagueDetailsRepository := repository.NewLeagueDetailsRepository(db)
	gameRepository := repository.NewGameRepository(db)
	playerRepository := repository.NewPlayerRepository(db)
	teamRepository := repository.NewTeamRepository(db)
	teamRosterRepository := repository.NewTeamRosterRepository(db)
	leagueSeriesRepository := repository.NewLeagueSeriesRepository(db)
	liveGameDetailsRepository := repository.NewLiveGameDetailsRepository(db)

	api.SetValveRateLimit(GetConfigFloat(`valve.rps`))

	_ = worker.NewDataLoader(
		leagueRepository,
		leagueDetailsRepository,
		gameRepository,
		playerRepository,
		teamRepository,
		teamRosterRepository,
		leagueSeriesRepository,
		liveGameDetailsRepository,
	)

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

	e.Use(IPRateLimitWithConfig(GetConfigInt(`throttling.per_min`), GetConfigInt(`throttling.burst`)))

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{GetConfigStr(`cors.origin`)},
		AllowMethods: []string{echo.GET, echo.HEAD},
	}))

	// Routes
	e.Static("/", "./public")

	dpcStandingsRepository := repository.NewDPCStandingsRepository(db)
	dpcResultsRepository := repository.NewDPCResultsRepository(db)
	matchMinimalRepository := repository.NewMatchMinimalRepository(db)

	leaguesHandler := handler.NewLeaguesHandler(leagueDetailsRepository, leagueSeriesRepository)
	gamesHandler := handler.NewGameHandler(gameRepository)
	teamsHandler := handler.NewTeamsHandler(teamRepository)
	dpcHandler := handler.NewDPCHandler(dpcStandingsRepository, api.LoadDPCStandings)
	dpcResultsHandler := handler.NewDPCResultsHandler(dpcResultsRepository, api.LoadDPCLeagueResults)
	matchMinimalHandler := handler.NewMatchMinimalHandler(matchMinimalRepository, api.LoadMatchMinimal)

	delivery.NewLeaguesDelivery(e, leaguesHandler, gamesHandler)
	delivery.NewTeamsDelivery(e, teamsHandler)
	delivery.NewDPCDelivery(e, dpcHandler, dpcResultsHandler, matchMinimalHandler)

	// Start server
	e.Logger.Fatal(e.Start(GetConfigStr(`server.address`)))
}
