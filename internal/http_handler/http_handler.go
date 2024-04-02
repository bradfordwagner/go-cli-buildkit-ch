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

func (h *Handler) handler(mode cache.HashMode, w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]
	l := log.Log().With("mode", mode, "hash", hash)
	l.Debug("received request")

	res, err := h.c.Get().ConsistentHash(cache.HashModeAPIGateway, hash, w)
	l.With("result", res, "error", err).Info("consistent hash result")

	fmt.Fprintf(w, res)
}

func (h *Handler) apiGateway(w http.ResponseWriter, r *http.Request) {
	h.handler(cache.HashModeAPIGateway, w, r)
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
	r.HandleFunc("/api-gateway/{hash}", h.apiGateway).Methods("GET")
	// r.HandleFunc("/api-gateway", h.apiGateway).Methods("GET")
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
