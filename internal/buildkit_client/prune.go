package buildkit_client

import (
	"context"
	"github.com/bradfordwagner/go-util/log"
	"github.com/dustin/go-humanize"
	"github.com/moby/buildkit/client"
	"time"
)

type PruneInterface interface {
	Prune(addr string, keepDuration, requestTimeout time.Duration) error
}

type prune struct {
}

// enforce interface
var _ PruneInterface = prune{}

func NewPrune() PruneInterface {
	return prune{}
}

func (p prune) Prune(addr string, keepDuration, requestTimeout time.Duration) (err error) {
	l := log.Log().With("component", "prune", "addr", addr)
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	buildkitClient, err := client.New(ctx, addr)
	if err != nil {
		l.With("err", err).Error("failed to create buildkit client")
		return err
	}

	usageInfo := make(chan client.UsageInfo)
	go func() {
		var totalBytes uint64
		for {
			select {
			case info, ok := <-usageInfo:
				totalBytes += uint64(info.Size)
				if !ok {
					l.Infof("total_bytes_pruned=%s", humanize.Bytes(totalBytes))
					return
				}
				l.Debugf("ID: %s, Size: %s, Description: %s", info.ID, humanize.Bytes(uint64(info.Size)), info.Description)
			}
		}
	}()
	defer buildkitClient.Close()
	defer close(usageInfo)
	err = buildkitClient.Prune(
		ctx,
		usageInfo,
		client.WithKeepOpt(keepDuration, 0, 0, 0),
	)
	if err != nil {
		l.With("err", err).Error("failed to prune")
		return
	}

	return
}
