package main

import (
	"flag"

	"github.com/fydmer/fileserver/internal/app"
	"github.com/fydmer/fileserver/internal/repositories/infra"
	"github.com/fydmer/fileserver/internal/repositories/storage"
	"github.com/fydmer/fileserver/internal/schema/database"
	"github.com/fydmer/fileserver/internal/servers/httpserver"
	"github.com/fydmer/fileserver/internal/services/controller"
	"github.com/fydmer/fileserver/pkg/pgconn"
)

type Config struct {
	Port     int
	Postgres pgconn.Config
}

func main() {
	a := app.NewApp()

	config := &Config{}
	{
		flag.IntVar(&config.Port, "port", 8080, "API server port")
		flag.StringVar(&config.Postgres.Host, "postgres.host", "postgres", "Postgres hostname")
		flag.UintVar(&config.Postgres.Port, "postgres.port", 5432, "Postgres port")
		flag.StringVar(&config.Postgres.DB, "postgres.db", "postgres", "Postgres database")
		flag.StringVar(&config.Postgres.Username, "postgres.username", "user", "Postgres username")
		flag.StringVar(&config.Postgres.Password, "postgres.password", "password", "Postgres password")
		flag.IntVar(&config.Postgres.MaxIdleConns, "postgres.max_idle_conns", 10, "Postgres Max idle connections")
		flag.IntVar(&config.Postgres.MaxOpenConns, "postgres.max_open_conns", 30, "Postgres Max open connections")
		flag.Parse()
	}

	pgConn, err := pgconn.NewDB(a.Context(), &config.Postgres)
	if err != nil {
		a.Panic(err)
	}
	a.AddStopFn(func() {
		_ = pgConn.Close()
	})

	if err = database.Apply(a.Context(), pgConn); err != nil {
		a.Panic(err)
	}

	infraRepo, err := infra.NewRepository(pgConn)
	if err != nil {
		a.Panic(err)
	}

	storageRepo, err := storage.NewRepository(pgConn)
	if err != nil {
		a.Panic(err)
	}

	controllerService := controller.NewController(infraRepo, storageRepo)

	server, err := httpserver.RunControllerServer(a.Context(), config.Port, controllerService)
	if err != nil {
		a.Panic(err)
	}
	a.AddStopFn(func() {
		server.Close()
	})

	a.Keep()
}
