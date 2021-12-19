package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"

	servErr "test/coins/errors"
)

// Registers http handlers for transfer servicce
// mr     - Mux router where handlers should be registered
// svc    - service to register
// logger - logger
func RegisterHandlers(mr *mux.Router, svc TransferService, logger kitlog.Logger) {
	var opts = []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}
	var sendPaymentHandler = kithttp.NewServer(
		makeSendPaymentEnpoint(svc),
		decodeSendPaymentRequest,
		encodeResponse,
		opts...,
	)

	mr.Handle("/api/v1/transfers", sendPaymentHandler).Methods("POST")

	var listTransfersHandler = kithttp.NewServer(
		makeListTransfersEndpoint(svc),
		decodeListTransfersRequest,
		encodeResponse,
		opts...,
	)

	mr.Handle("/api/v1/accounts/{account}/transfers", listTransfersHandler).Methods("GET")
}

func RegisterListTranfersHandler(accountHander *mux.Router, svc TransferService, logger kitlog.Logger) http.Handler {
	//var r = mux.NewRouter()

	return accountHander
}

func decodeListTransfersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var vars = mux.Vars(r)
	accountNumber, ok := vars["account"]
	if !ok {
		return nil, errors.New("bad route")
	}

	accNum, err := strconv.ParseUint(accountNumber, 10, 64)
	if err != nil {
		return nil, errors.New("bad route")
	}
	return listTransfersRequest{accNum}, nil
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

	switch svcErr := err.(type) {
	case servErr.ServiceError:
		if svcErr.Kind() == servErr.ErrorKindDB {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		if strings.HasPrefix("json:", err.Error()) {
			// dirty hack, did not found way to determine if specific error is json serialization error
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
