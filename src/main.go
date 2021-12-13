package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test/coins/account"
	"test/coins/payment"

	"github.com/go-kit/log"
)

func main() {
	var httpAddress = "http://localhost:8080"

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	var httpLogger = log.With(logger, "component", "http")

	var accountService = account.NewAccountService()
	var paymentService = payment.NewPaymentService()

	var mux = http.NewServeMux()
	mux.Handle("/api/v1/accounts", account.MakeHandler(accountService, httpLogger))
	mux.Handle("/api/v1/payments", payment.MakeHandler(paymentService, httpLogger))

	http.Handle("/", accessControl(mux))

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", httpAddress, "msg", "listening")
		errs <- http.ListenAndServe(httpAddress, nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
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
