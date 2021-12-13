package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
)

func MakeHandler(svc PaymentService, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	listPaymentsHandler := kithttp.NewServer(
		makeListPaymentsEndpoint(svc),
		decodeListPaymentsRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/account/{id}/payments", listPaymentsHandler).Methods("GET")

	sendPaymentHandler := kithttp.NewServer(
		makeSendPaymentEnpoint(svc),
		decodeSendPaymentRequest,
		encodeResponse,
		opts...,
	)

	r.Handle("/api/v1/payments", sendPaymentHandler).Methods("POST")

	return r
}

func decodeListPaymentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errors.New("bad route")
	}

	guid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("bad route")
	}
	return listPaymentsRequest{guid}, nil
}

func decodeSendPaymentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body sendPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}

type errorer interface {
	error() error
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
