package kube_watcher

import (
	"bkch/internal/args"
	"context"

	"github.com/bradfordwagner/go-util/log"

	"go.uber.org/zap"
)

type Watcher struct {
	a   args.ServerArgs
	ctx context.Context
	l   *zap.SugaredLogger
}

// NewWatcher creates a new Watcher
func NewWatcher(ctx context.Context, a args.ServerArgs) *Watcher {
	l := log.Log().With("module", "kube_watcher")
	return &Watcher{
		l:   l,
		a:   a,
		ctx: ctx,
	}
}

func (w *Watcher) Start() {
	w.l.Info("starting")
	// watch kubernetes
}
