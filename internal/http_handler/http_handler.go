package http_handler

import (
	"bkch/internal/args"
	"bkch/internal/cache"
	"context"
	"fmt"
	"net/http"

	bwutil "github.com/bradfordwagner/go-util"
	"github.com/bradfordwagner/go-util/log"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	a      args.ServerArgs
	c      *bwutil.Lockable[*cache.Cache]
	ctx    context.Context
	cancel context.CancelFunc
	l      *zap.SugaredLogger
}

func (h *Handler) handler(w http.ResponseWriter, r *http.Request) {
	l := log.Log()
	l.Debug("received request")

	i := "hello"
	res, err := h.c.Get().ConsistentHash(cache.HashModeAPIGateway, i, w)
	l.With("input", i, "result", res, "error", err).Info("consistent hash result")

	fmt.Fprintf(w, res)
}

// NewHandler creates a new Handler
func NewHandler(
	ctx context.Context,
	cancel context.CancelFunc,
	a args.ServerArgs,
	c *bwutil.Lockable[*cache.Cache],
) *Handler {
	l := log.Log().With(
		"module", "http_handler",
		"port", a.Port,
	)
	return &Handler{
		a:      a,
		c:      c,
		cancel: cancel,
		ctx:    ctx,
		l:      l,
	}
}

// Start starts the http server
// will invoke cancel if it fails to start
func (h *Handler) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", h.handler)
	http.Handle("/", r)
	address := fmt.Sprintf(":%d", h.a.Port)
	server := &http.Server{Addr: address}

	// start the server
	go func() {
		h.l.Info("starting http server")
		if err := server.ListenAndServe(); err != nil {
			h.l.With("error", err).Error("failed to start http server, shutting down")
			h.cancel()
		}
	}()

	// listen for shutdown signal
	go func() {
		h.l.Info("listening for shutdown signal")
		<-h.ctx.Done()
		h.l.Info("shutting down http server")
		if err := server.Shutdown(h.ctx); err != nil {
			h.l.With("error", err).Error("failed to cleanly shutdown http server")
		}
	}()
}
