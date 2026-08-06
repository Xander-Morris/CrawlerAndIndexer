package server

import "net/http"

func New(addr string) *http.Server {
	mux := http.NewServeMux()
	registerRoutes(mux)

	return &http.Server{
		Addr:    addr,
		Handler: withCORS(mux),
	}
}
