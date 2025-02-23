package tfetest

import (
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc"
)

type Middleware func(http.Handler) http.Handler

func Setup(middlewares ...Middleware) (*tfc.Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()

	// Apply middlewares in reverse order so they execute in the order they
	// were passed
	var handler http.Handler = mux
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	server := httptest.NewServer(handler)

	client, err := tfe.NewClient(&tfe.Config{
		Address: server.URL,
		Token:   "token",
	})
	if err != nil {
		panic(err)
	}

	return tfc.NewClient(client), mux, server.Close
}
