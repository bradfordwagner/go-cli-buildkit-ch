package server_cmd

import (
	"bkch/internal/args"
	"bkch/internal/cache"
	"bkch/internal/http_handler"
	"bkch/internal/kube_watcher"
	"context"

	bwutil "github.com/bradfordwagner/go-util"
	"github.com/bradfordwagner/go-util/log"
	"github.com/bradfordwagner/go-util/shutdown"
)

// Run is the main function for the serverCmd command
func Run(a args.ServerArgs) (err error) {
	l := log.Log().With("args", a)

	// setup shutdown listener
	ctx := context.Background() // base level context
	ctx, cancel := shutdown.Listen(ctx, func() {
		l.Info("shutdown listener invoked")
	})
	l.Info("initialized shutdown listener")

	// create cache
	c := bwutil.NewLockableWithValue[*cache.Cache](cache.NewCache())

	// watch kubernetes
	kube_watcher.NewWatcher(ctx, cancel, a, c).Start()

	// start http server on another routine
	http_handler.NewHandler(ctx, cancel, a, c).Start()

	// wait for shutdown
	<-ctx.Done()
	return
}
