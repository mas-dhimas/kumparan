package main

import (
	"context"
	"flag"
	"fmt"
	"kumparan-test/config"
	"kumparan-test/internal/api"
	"kumparan-test/internal/article"
	"kumparan-test/internal/author"
	"kumparan-test/pkg/database"
	"kumparan-test/pkg/search"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

const (
	DEVELOPER  = "Adhimas W Ramadhana"
	ASCIIImage = `
	
	___  __    ___  ___  _____ ______   ________  ________  ________  ________  ________      	
	|\  \|\  \ |\  \|\  \|\   _ \  _   \|\   __  \|\   __  \|\   __  \|\   __  \|\   ___  \          
	\ \  \/  /|\ \  \\\  \ \  \\\__\ \  \ \  \|\  \ \  \|\  \ \  \|\  \ \  \|\  \ \  \\ \  \         
	 \ \   ___  \ \  \\\  \ \  \\|__| \  \ \   ____\ \   __  \ \   _  _\ \   __  \ \  \\ \  \        
	  \ \  \\ \  \ \  \\\  \ \  \    \ \  \ \  \___|\ \  \ \  \ \  \\  \\ \  \ \  \ \  \\ \  \       
	   \ \__\\ \__\ \_______\ \__\    \ \__\ \__\    \ \__\ \__\ \__\\ _\\ \__\ \__\ \__\\ \__\      
	    \|__| \|__|\|_______|\|__|     \|__|\|__|     \|__|\|__|\|__|\|__|\|__|\|__|\|__| \|__|      
	                           _________  _______   ________  _________                              
	                          |\___   ___|\  ___ \ |\   ____\|\___   ___\                            
	 ____________  ___________\|___ \  \_\ \   __/|\ \  \___|\|___ \  \_|____________  ____________  
	|\____________|\____________\  \ \  \ \ \  \_|/_\ \_____  \   \ \  \|\____________|\____________\
	\|____________\|____________|   \ \  \ \ \  \_|\ \|____|\  \   \ \  \|____________\|____________|
	                                 \ \__\ \ \_______\____\_\  \   \ \__\                           
	                                  \|__|  \|_______|\_________\   \|__|                           
	                                                  \|_________|       
	
	`
	// CHOOSE ONE, EITHER FOR DOCKER OR LOCAL SETUP
	// MIGRATION_PATH = `file:///app/migrations` // DOCKER
	MIGRATION_PATH = `file://./migrations` // LOCAL
)

func main() {
	customFormatter := &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d", f.File, f.Line)
		},
	}
	logrus.SetFormatter(customFormatter)
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.InfoLevel)

	// Default configuration file is empty string, OS ENV variable will be used if config file empty or not found
	configPath := flag.String("config", "", "config file path")
	migrateDB := flag.Bool("migrate", false, "run database migrations and exit")

	flag.Parse()

	// Get Config
	configLoader := config.NewConfig(*configPath)
	serviceConfig, err := configLoader.GetServiceConfig()
	if err != nil {
		logrus.Fatalf("Unable to load configuration: %v", err)
	}

	level, err := logrus.ParseLevel(serviceConfig.ServiceData.LogLevel)
	if err != nil {
		logrus.Fatalf("Unable to read log level : %s", err.Error())
	} else {
		logrus.SetLevel(level)
	}

	// Pre-printed text at startup.
	logrus.Infof("Kumparan Technical Test")
	logrus.Infof("Developed by %v.", DEVELOPER)
	logrus.Infof(ASCIIImage)
	logrus.Infof("Start service...")

	// Initialize PostgreSQL
	dbPool, err := database.NewPostgresDB(&serviceConfig.SourceData)
	if err != nil {
		logrus.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	if *migrateDB {
		logrus.Info("Running database migrations...")
		if err := runMigrations(buildPostgresDSN(&serviceConfig.SourceData)); err != nil {
			logrus.Fatalf("Database migration failed: %v", err)
		}
		logrus.Info("Database migrations completed successfully. Exiting.")
		os.Exit(0)
	}

	// Initialize Elasticsearch
	esClient, err := search.NewElasticsearchClient(serviceConfig.SourceData.ElasticURL)
	if err != nil {
		logrus.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	defer esClient.Stop()

	// Initialize Repositories, Services, and Handlers
	searchService := search.NewSearchService(esClient)
	authorRepo := author.NewPostgresRepository(dbPool)
	authorService := author.NewAuthorService(authorRepo)

	articleRepo := article.NewPostgresRepository(dbPool)
	articleService := article.NewArticleService(articleRepo, authorService, searchService)
	apiHandler := api.NewHandler(articleService)

	// Echo instance
	e := echo.New()
	e.Logger.SetOutput(logrus.StandardLogger().Writer())
	switch logrus.GetLevel() {
	case logrus.DebugLevel:
		e.Logger.SetLevel(log.DEBUG)
	case logrus.InfoLevel:
		e.Logger.SetLevel(log.INFO)
	case logrus.WarnLevel:
		e.Logger.SetLevel(log.WARN)
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		e.Logger.SetLevel(log.ERROR)
	default:
		e.Logger.SetLevel(log.INFO)
	}

	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(serviceConfig.ServiceData.RateLimit))))

	apiHandler.RegisterRoutes(e)

	go func() {
		if err := e.Start(fmt.Sprintf(":%s", serviceConfig.ServiceData.Address)); err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited gracefully")
}

// runMigrations applies database migrations using golang-migrate.
func runMigrations(dsn string) error {
	m, err := migrate.New(
		MIGRATION_PATH,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func buildPostgresDSN(cfg *config.SourceDataConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=%d",
		cfg.PostgresDBUsername,
		cfg.PostgresDBPassword,
		cfg.PostgresDBServer,
		cfg.PostgresDBPort,
		cfg.PostgresDBName,
		cfg.PostgresDBTimeout,
	)
}
