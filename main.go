package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"mp2720/subscriptions/sqlgen"

	_ "mp2720/subscriptions/docs"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/lib/pq"

	"github.com/pressly/goose/v3"
)

type config struct {
	DB         string `env:"DB,required"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBPort     uint16 `env:"DB_PORT,required"`
	DBHost     string `env:"DB_HOST,required"`
	DBSSLMode  string `env:"DB_SSL_MODE,required"`
	Verbose    bool   `env:"SERVICE_VERBOSE"`
	Port       uint16 `env:"SERVICE_PORT,required"`
}

//go:embed sql/migrations/*.sql
var embedMigrations embed.FS

func initQueries(cfg *config) (*sqlgen.Queries, error) {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"dbname=%s user=%s password=%s port=%d host=%s sslmode=%s",
		cfg.DB,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBPort,
		cfg.DBHost,
		cfg.DBSSLMode,
	))
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, "sql/migrations"); err != nil {
		return nil, err
	}

	return sqlgen.New(db), nil
}

//	@version	1.0
//	@host		localhost:8080
//	@BasePath	/api/v1

func main() {
	cfg, err := env.ParseAs[config]()
	if err != nil {
		log.Fatalf("failed to parse env: %s", err)
	}

	queries, err := initQueries(&cfg)
	if err != nil {
		log.Fatalf("failed to init db queries: %s", err)
	}

	logger := NewLoger(cfg.Verbose)

	r := gin.Default()
	r.Use(ErrorHandler(logger))

	api := API{
		Queries: queries,
		Log:     NewLoger(cfg.Verbose),
	}
	api.RegisterHandlers(r.Group("/api/v1"))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
