package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test/coins/account"
	"test/coins/db"
	"test/coins/transfer"
	"time"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

func main() {
	// Loading database connection string
	godotenv.Load()
	var cs = os.Getenv("DATABASE_CS")

	if cs == "" {
		panic("unable to start service - connection string is not provided")
	}

	// Creating new connection pool
	cnPool, err := db.NewConnectionPool(cs)
	if err != nil {
		panic("Unable to create connection pool: " + err.Error())
	}

	defer cnPool.Close()

	// Initializing logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	var httpLogger = log.With(logger, "component", "http")

	// Initializing services
	var factory = func() (db.DbContext, error) {
		return db.CreateContext(cnPool, time.Second*5)
	}
	var accountService = account.NewAccountService(factory)
	var transferService = transfer.NewTransferService(factory)

	// Registering routes and handles
	var mr = mux.NewRouter()

	account.RegisterHandlers(mr, accountService, httpLogger)
	transfer.RegisterHandlers(mr, transferService, httpLogger)
	http.Handle("/", accessControl(mr))

	// Setting up http server
	var httpAddress = "localhost:" + os.Getenv("PORT")
	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", httpAddress, "msg", "listening")
		httpError := http.ListenAndServe(httpAddress, nil)
		errs <- httpError
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h *mux.Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
