package http_handler

import (
	"bkch/internal/args"
	"context"
	"fmt"
	"net/http"

	"github.com/bradfordwagner/go-util/log"
	"github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
	l := log.Log()
	l.Info("received request")
}

// Start starts the http server
// will invoke cancel if it fails to start
func Start(ctx context.Context, cancel context.CancelFunc, a args.ServerArgs) {
	l := log.Log().With(
		"module", "http_handler",
		"port", a.Port,
	)
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	http.Handle("/", r)
	address := fmt.Sprintf(":%d", a.Port)
	server := &http.Server{Addr: address}

	// start the server
	go func() {
		l.Info("starting http server")
		if err := server.ListenAndServe(); err != nil {
			l.With("error", err).Error("failed to start http server, shutting down")
			cancel()
		}
	}()

	// listen for shutdown signal
	go func() {
		l.Info("listening for shutdown signal")
		<-ctx.Done()
		l.Info("shutting down http server")
		if err := server.Shutdown(ctx); err != nil {
			l.With("error", err).Error("failed to cleanly shutdown http server")
		}
	}()
}
