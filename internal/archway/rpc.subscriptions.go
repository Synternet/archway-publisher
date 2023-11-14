package archway

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

func setMaxValue(a *atomic.Uint64, v uint64) {
	for {
		oldValue := a.Load()
		if oldValue >= v {
			return
		}

		if a.CompareAndSwap(oldValue, v) {
			return
		}
	}
}

func (c *rpc) bufferChannel(events <-chan ctypes.ResultEvent, size int) <-chan ctypes.ResultEvent {
	ch := make(chan ctypes.ResultEvent, size)
	c.group.Go(func() error {
		defer close(ch)
		for {
			select {
			case <-c.ctx.Done():
				log.Println("bufferChannel: Context Done")
				return nil
			case ev, ok := <-events:
				if !ok {
					log.Println("bufferChannel: events closed")
					return nil
				}

				c.evtCounter.Add(1)

				setMaxValue(&c.queueMaxSize, uint64(len(ch)))

				select {
				case ch <- ev:
				default:
					c.evtSkipCounter.Add(1)
					log.Println("bufferChannel: Overflow! Skipping event: ", ev.Query)
				}
			}
		}
	})
	return ch
}

func (c *rpc) subscribeTransactions(publish func(msg any, suffixes ...string) error) error {
	ctx, cancel := context.WithTimeout(c.ctx, time.Second*5)
	events, err := c.tendermint.Subscribe(ctx, subscriberName, fmt.Sprintf("tm.event='%s'", tmtypes.EventTx), 10)
	cancel()
	if err != nil {
		return err
	}
	c.group.Go(func() error {
		return c.handleSubscriptions(publish, c.bufferChannel(events, 2048), time.Minute*10)
	})
	return nil
}

func (c *rpc) subscribeBlocks(publish func(msg any, suffixes ...string) error) error {
	ctx, cancel := context.WithTimeout(c.ctx, time.Second*5)
	events, err := c.tendermint.Subscribe(ctx, subscriberName, fmt.Sprintf("tm.event='%s'", tmtypes.EventNewBlock), 10)
	cancel()
	if err != nil {
		return err
	}

	c.group.Go(func() error {
		return c.handleSubscriptions(publish, c.bufferChannel(events, 2048), time.Minute)
	})
	return nil
}

func (c *rpc) handleSubscriptions(publish func(msg any, suffixes ...string) error, events <-chan ctypes.ResultEvent, timeout time.Duration) error {
	sentinel := time.NewTimer(timeout)
	lastEvent := time.Now()
	c.group.Go(
		func() error {
			defer sentinel.Stop()
			for {
				select {
				case <-c.ctx.Done():
					return nil
				case <-sentinel.C:
					err := fmt.Errorf("event subscription timed out, last seen: %s", time.Since(lastEvent))
					return err
				}
			}
		},
	)

	for {
		select {
		case <-c.ctx.Done():
			log.Println("handleSubscriptions: c.Context Done")
			return nil
		case ev, ok := <-events:
			if !ok {
				log.Println("handleSubscriptions: events closed")
				return nil
			}

			if !sentinel.Stop() {
				return fmt.Errorf("event subscription timed out while resetting, last seen: %s", time.Since(lastEvent))
			}
			sentinel.Reset(timeout)
			lastEvent = time.Now()

			switch data := ev.Data.(type) {
			case tmtypes.EventDataNewBlock:
				c.blockCounter.Add(1)
				publish(
					c.translateBlock(data.Block),
					"block",
				)
			case tmtypes.EventDataTx:
				c.txCounter.Add(1)
				txData := data.GetTx()
				hash := hex.EncodeToString(tmtypes.Tx(txData).Hash())
				tx := c.translateTransaction(txData, hash, fmt.Sprint(c.counter.Add(1)), &data.TxResult, &data.TxResult.Result.Code)
				publish(
					tx,
					"tx",
				)
				log.Println("Transaction: ", tx.TxID, extractTxMessageNames(tx))
			default:
				c.evtOtherCounter.Add(1)
			}
		}
	}
}
