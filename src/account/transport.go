package account

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
)

func MakeHandler(ts AccountService, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	listAccountsHandler := kithttp.NewServer(
		makeListAccountsEndpoint(ts),
		decodeListAccontsRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/accounts", listAccountsHandler).Methods("GET")

	return r
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
	switch err {
	// case cargo.ErrUnknown:
	// 	w.WriteHeader(http.StatusNotFound)
	// case ErrInvalidArgument:
	// 	w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
