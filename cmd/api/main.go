package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/jsonlog"
	"github.com/jempe/whatsapp_backup/internal/mailer"
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
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	embeddings struct {
		defaultProvider               string
		sentenceTransformersServerURL string
		embeddingsPerBatch            int
		maxTokens                     int
	}
	apiKeys struct {
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	site struct {
		baseURL string
	}
	doCronJob bool
	//custom_config_variables
}

type application struct {
	config        config
	logger        *jsonlog.Logger
	templateCache map[string]*template.Template
	models        data.Models
	mailer        mailer.Mailer
	wg            sync.WaitGroup
}

func main() {
	var cfg config

	// API Environment Settings
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// API Web Server Settings
	flag.IntVar(&cfg.port, "port", 8001, "API server port")

	// API Database Settings
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://whatsapp:password@localhost/whatsapp?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// API Rate Limiter Settings
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")

	// API SMTP Settings
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Jempe <no-reply@jempe.org>", "SMTP sender")

	// API Embeddings Settings
	flag.IntVar(&cfg.embeddings.embeddingsPerBatch, "embeddings-per-batch", 150, "Embeddings per batch")
	flag.IntVar(&cfg.embeddings.maxTokens, "max-tokens", 2500, "Max tokens per document")

	// API Embeddings URL for Sentence Transformers Server
	flag.StringVar(&cfg.embeddings.sentenceTransformersServerURL, "sentence-transformers-server-url", "", "Sentence Transformers Server URL. Example: http://localhost:5000")

	// Set Default Embeddings Provider
	flag.StringVar(&cfg.embeddings.defaultProvider, "default-embeddings-provider", "sentence-transformers", "Default embeddings provider (google|openai|sentence-transformers)")

	// API Behavior Flags
	flag.BoolVar(&cfg.doCronJob, "do-cron-job", false, "Run cron job")

	//Site Setup
	flag.StringVar(&cfg.site.baseURL, "site-base-url", "http://localhost:8001/", "Base URL for the site")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	//custom_config_flags

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Binary Version:\t%s\nDB Version:\t%d\n", version, dbVersion)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	logger.PrintInfo("Starting Template API", map[string]string{
		"version":   version,
		"env":       cfg.env,
		"dbversion": fmt.Sprintf("%d", dbVersion),
		"limiter":   fmt.Sprintf("%v", cfg.limiter.enabled),
		"rps":       fmt.Sprintf("%f", cfg.limiter.rps),
		"burst":     fmt.Sprintf("%d", cfg.limiter.burst),
	})

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("Database connection pool established", nil)

	app := &application{
		config:        cfg,
		logger:        logger,
		templateCache: templateCache,
		models:        data.NewModels(db),
		mailer:        mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	//custom_code

	if cfg.doCronJob {
		logger.PrintInfo("Starting Cron Job", nil)

		cronjobErr := app.runCronJob()

		if cronjobErr != nil {
			logger.PrintFatal(cronjobErr, nil)
		}
	} else {

		err = app.serve()
		if err != nil {
			logger.PrintFatal(err, nil)
		}

	}
}
