package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/JohnKucharsky/go-kit/account"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	log2 "github.com/go-kit/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var connDB = "postgresql://postgres:pass@localhost:5432/data?sslmode=disable"

func main() {
	var httpAddr = flag.String("http", ":8080", "http listen address")
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger, "srv", "account", "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	_ = level.Info(logger).Log("msg", "service started")
	defer func(info log2.Logger, msg string) {
		_ = info.Log(msg)
	}(level.Info(logger), "service ended")

	var db *sql.DB
	{
		var err error

		db, err = sql.Open("postgres", connDB)
		if err != nil {
			_ = level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			fmt.Println("error running migrations:", err.Error())
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file:///Projects/go-kit/migrations",
			"postgres", driver)
		if err != nil {
			fmt.Println("error running migrations:", err.Error())
		}
		if m != nil {
			err = m.Up()
			if err != nil {
				fmt.Println("error running migrations:", err.Error())
			}
		}

	}

	flag.Parse()
	ctx := context.Background()
	var srv account.Service
	{
		repository := account.NewRepo(db, logger)

		srv = account.NewService(repository, logger)
	}

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	endpoints := account.MakeEndpoints(srv)

	go func() {
		fmt.Println("listening on", *httpAddr)
		handler := account.NewHTTPServer(ctx, endpoints)
		errs <- http.ListenAndServe(*httpAddr, handler)
	}()

	_ = level.Error(logger).Log("exit", <-errs)
}
