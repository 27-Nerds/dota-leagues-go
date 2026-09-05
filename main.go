package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	if viper.IsSet(key) {
		return viper.GetFloat64(key)
	}

	return 1.5
}

func init() {
	viper.SetEnvPrefix("DOTA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
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

func dbConnection(ctx context.Context) (*db.ArangoDB, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return db.Connect(ctx,
		GetConfigStr(`database.url`),
		GetConfigStr(`database.user`),
		GetConfigStr(`database.pass`),
		GetConfigStr(`database.name`))
}

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	prodFlag := flag.Bool("production", false, "run in production mode")
	indexesFlag := flag.Bool("create-indexes", false, "create missing database indexes and exit without starting collectors or HTTP")
	flag.Parse()

	if *prodFlag {
		ENVIRONMENT = "production"
	}

	slog.Info("starting application", "environment", ENVIRONMENT)

	db, err := dbConnection(ctx)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	if err := db.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("initialize database indexes: %w", err)
	}
	if *indexesFlag {
		return nil
	}
	if err := db.EnsureCoreCollections(ctx); err != nil {
		return fmt.Errorf("initialize database collections: %w", err)
	}
	leagueRepository := repository.NewLeagueRepository(db)
	leagueDetailsRepository := repository.NewLeagueDetailsRepository(db)
	gameRepository := repository.NewGameRepository(db)
	playerRepository := repository.NewPlayerRepository(db)
	teamRepository := repository.NewTeamRepository(db)
	teamRosterRepository := repository.NewTeamRosterRepository(db)
	leagueSeriesRepository := repository.NewLeagueSeriesRepository(db)
	liveGameDetailsRepository := repository.NewLiveGameDetailsRepository(db)
	updatesRepository := repository.NewUpdatesRepository(db)
	if err := db.EnsureUpdatesCollection(ctx); err != nil {
		return fmt.Errorf("initialize updates feed: %w", err)
	}

	sourcesRepository := repository.NewSourceRepository(db)
	if err := db.EnsureSourceCollections(ctx); err != nil {
		return fmt.Errorf("initialize source archive: %w", err)
	}
	api.SetSourceRecorder(sourcesRepository)
	api.SetValveRateLimit(GetConfigFloat(`valve.rps`))

	loader := worker.NewDataLoader(
		ctx,
		leagueRepository,
		leagueDetailsRepository,
		gameRepository,
		playerRepository,
		teamRepository,
		teamRosterRepository,
		leagueSeriesRepository,
		liveGameDetailsRepository,
		updatesRepository,
	)
	defer loader.Stop()

	//----------------
	//START WEB SERVER
	//----------------

	// Echo instance
	e := echo.New()
	e.Server.BaseContext = func(net.Listener) context.Context { return ctx }

	// Middleware
	e.Use(middleware.RequestLogger())
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
	delivery.NewStaticDelivery(e, "./public", os.Getenv("ASSET_DIR"))
	e.GET("/healthz", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	dpcStandingsRepository := repository.NewDPCStandingsRepository(db)
	dpcResultsRepository := repository.NewDPCResultsRepository(db)
	matchMinimalRepository := repository.NewMatchMinimalRepository(db)

	leaguesHandler := handler.NewLeaguesHandler(leagueDetailsRepository, leagueSeriesRepository)
	gamesHandler := handler.NewGameHandler(gameRepository)
	teamsHandler := handler.NewTeamsHandler(teamRepository)
	dpcHandler := handler.NewDPCHandler(dpcStandingsRepository, api.LoadDPCStandings)
	dpcResultsHandler := handler.NewDPCResultsHandler(dpcResultsRepository, api.LoadDPCLeagueResults)
	matchMinimalHandler := handler.NewMatchMinimalHandler(matchMinimalRepository, playerRepository, api.LoadMatchMinimal)

	delivery.NewLeaguesDelivery(e, leaguesHandler, gamesHandler)
	delivery.NewTeamsDelivery(e, teamsHandler)
	updatesHandler := handler.NewUpdatesHandler(updatesRepository)
	delivery.NewUpdatesDelivery(e, updatesHandler)
	delivery.NewSourceDelivery(e, sourcesRepository)
	delivery.NewDPCDelivery(e, dpcHandler, dpcResultsHandler, matchMinimalHandler)

	siteURL := os.Getenv("SITE_URL")
	if siteURL == "" {
		siteURL = "https://dota-leagues.27n.gg"
	}
	if err := delivery.NewPagesDelivery(e, leaguesHandler, teamsHandler, matchMinimalHandler, dpcHandler, updatesHandler, leagueSeriesRepository, "./public/index.html", siteURL, os.Getenv("GA_MEASUREMENT_ID")); err != nil {
		return fmt.Errorf("configure page routes: %w", err)
	}

	serverErrors := make(chan error, 1)
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", GetConfigStr(`server.address`))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	e.Server.Handler = e
	go func() { serverErrors <- e.Server.Serve(listener) }()
	serverStopped := false
	select {
	case err = <-serverErrors:
		serverStopped = true
		cancel()
	case <-ctx.Done():
	}
	// Shutdown gets its own deadline because the application context is canceled.
	shutdownCtx, stopShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopShutdown()
	shutdownErr := e.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, e.Close())
	}
	if !serverStopped {
		err = <-serverErrors
	}
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return errors.Join(err, shutdownErr)
}
