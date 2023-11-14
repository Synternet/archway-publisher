package archway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/syntropynet/archway-publisher/pkg/types"
	"github.com/syntropynet/data-layer-sdk/pkg/options"
	"github.com/syntropynet/data-layer-sdk/pkg/service"
)

type Publisher struct {
	*service.Service
	rpc               *rpc
	chainId           string
	mempoolMessages   atomic.Uint64
	publishedMessages atomic.Uint64
}

func New(opts ...options.Option) *Publisher {
	ret := &Publisher{
		Service: &service.Service{},
	}

	ret.Configure(opts...)

	var err error
	rpc, err := newRpc(ret.Context, ret.Cancel, ret.Group, ret.TendermintApi(), ret.RPCApi(), ret.GRPCApi())
	if err != nil {
		log.Println("Could not connect to Archway: ", err.Error())
		return nil
	}
	ret.rpc = rpc

	id, err := rpc.ChainID()
	if err != nil {
		log.Println("Failed to retrieve chain ID: ", err.Error())
		return nil
	}
	ret.chainId = id
	log.Println("Chain ID:", id)

	ret.AddStatusCallback(ret.getStatus)
	ret.AddStatusCallback(ret.rpc.getStatus)

	return ret
}

func (p *Publisher) Start() context.Context {
	p.mempoolMessages.Store(0)

	err := p.rpc.Subscribe(func(msg any, suffixes ...string) error {
		p.publishedMessages.Add(1)
		return p.Publish(msg, suffixes...)
	})
	if err != nil {
		p.Fail(err)
		return p.Context
	}

	mempoolTicker := time.NewTicker(p.MempoolPeriod())
	p.Group.Go(
		func() error {
			for {
				select {
				case <-p.Context.Done():
					return nil
				case <-mempoolTicker.C:
					if p.rpc == nil {
						continue
					}
					pool, err := p.rpc.Mempool()
					if err != nil {
						log.Println("Mempool failed: ", err.Error())
						continue
					}
					if pool != nil {
						p.mempoolMessages.Add(uint64(len(pool)))
						p.Publish(
							&types.Mempool{
								Transactions: pool,
							},
							"mempool",
						)
					}
				}
			}
		},
	)
	return p.Service.Start()
}

func (p *Publisher) Close() error {
	log.Println("Publisher.Close")
	p.Cancel(nil)

	p.RemoveStatusCallback(p.getStatus)
	p.RemoveStatusCallback(p.rpc.getStatus)

	var err []error
	err = append(err, fmt.Errorf("failure during RPC Close: %w", p.rpc.Close()))

	log.Println("Waiting on publisher group")
	errGr := p.Group.Wait()
	if !errors.Is(errGr, context.Canceled) {
		err = append(err, errGr)
	}
	log.Println("Publisher.Close DONE")
	return errors.Join(err...)
}

func (p *Publisher) getStatus() map[string]any {
	return map[string]any{
		"mempool.txs": p.mempoolMessages.Swap(0),
		"published":   p.publishedMessages.Swap(0),
	}
}
