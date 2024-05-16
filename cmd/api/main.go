package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/jsonlog"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	openAIApiKey               string
	sentenceTransformersEnable bool
	embeddingsPerBatch         int
	basicAuth                  struct {
		username string
		password string
	}
	scriptsPath struct {
		pythonBinary string
		path         string
	}
	doCronJob bool
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.basicAuth.username, "basic-auth-username", "admin", "Basic Auth username")
	flag.StringVar(&cfg.basicAuth.password, "basic-auth-password", "password", "Basic Auth password")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://whatsapp:password@localhost/whatsapp?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.openAIApiKey, "openai-api-key", "", "OpenAI API key")
	flag.BoolVar(&cfg.sentenceTransformersEnable, "sentence-transformers-enable", false, "Enable Sentence Transformers")
	flag.IntVar(&cfg.embeddingsPerBatch, "embeddings-per-batch", 150, "Embeddings per batch")

	flag.StringVar(&cfg.scriptsPath.pythonBinary, "python-binary", "python3", "Python binary path")
	flag.StringVar(&cfg.scriptsPath.path, "scripts-path", "./scripts", "Python scripts folder path")

	flag.BoolVar(&cfg.doCronJob, "do-cron-job", false, "Run cron job")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("Database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	if cfg.doCronJob {
		logger.PrintInfo("Starting Cron Job", nil)

		app.runCronJob()
	} else {
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.port),
			Handler:      app.routes(),
			ErrorLog:     log.New(logger, "", 0),
			IdleTimeout:  time.Minute,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		logger.PrintInfo("Starting server", map[string]string{
			"addr": srv.Addr,
			"env":  cfg.env,
		})

		err = srv.ListenAndServe()
		logger.PrintFatal(err, nil)
	}
}
