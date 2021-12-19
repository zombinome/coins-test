package account

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"

	servErr "test/coins/errors"
)

// Registers http handlers for account servicce
// mr     - Mux router where handlers should be registered
// svc    - service to register
// logger - logger
func RegisterHandlers(mr *mux.Router, svc AccountService, logger kitlog.Logger) {
	var opts = []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	var listAccountsHandler = kithttp.NewServer(
		makeListAccountsEndpoint(svc),
		decodeListAccontsRequest,
		encodeResponse,
		opts...,
	)

	mr.Handle("/api/v1/accounts", listAccountsHandler).Methods("GET")
}

type errorer interface {
	error() error
}

func decodeListAccontsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(ctx context.Context, wr http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), wr)
		return nil
	}
	wr.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(wr).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch svcErr := err.(type) {
	case servErr.ServiceError:
		if svcErr.Kind() == servErr.ErrorKindDB {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
