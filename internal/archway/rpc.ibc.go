package archway

import (
	"context"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/types/query"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

func (c *rpc) GetDenomTrace(denomStr string) (ibctypes.DenomTrace, error) {
	c.ibcMisses.Add(1)
	ctx, cancel := context.WithTimeout(c.ctx, 1*time.Second)
	defer cancel()

	req := &ibctypes.QueryDenomTraceRequest{
		Hash: denomStr,
	}

	res, err := c.ibcQueryClient.DenomTrace(ctx, req)
	if err != nil {
		c.errCounter.Add(1)
		return ibctypes.DenomTrace{}, err
	}

	return *res.DenomTrace, nil
}

func (c *rpc) preHeatDenomTraceCache() {
	var nextPageKey []byte

	for {
		req := &ibctypes.QueryDenomTracesRequest{
			Pagination: &query.PageRequest{
				Key:   nextPageKey,
				Limit: 100, // Adjust the limit as necessary
			},
		}

		ctx, cancel := context.WithTimeout(c.ctx, 1*time.Second)
		defer cancel()

		res, err := c.ibcQueryClient.DenomTraces(ctx, req)
		if err != nil {
			c.errCounter.Add(1)
			log.Printf("Failed to fetch denom traces: %v\n", err)
			return
		}

		for _, trace := range res.DenomTraces {
			c.ibcTraceCache[trace.IBCDenom()] = trace
		}

		nextPageKey = res.Pagination.NextKey
		if nextPageKey == nil {
			break
		}

	}
	log.Printf("IBC Denoms fetched: %v\n", len(c.ibcTraceCache))
}

func (c *rpc) getDenomTraceFromCache(denomStr string) (ibctypes.DenomTrace, error) {
	// Check if the denomStr is in the cache
	if trace, found := (c.ibcTraceCache)[denomStr]; found {
		// If found, return the trace
		return trace, nil
	}

	// If not found in the cache, call GetDenomTrace
	res, err := c.GetDenomTrace(denomStr)
	if err != nil {
		return ibctypes.DenomTrace{}, err
	}

	// Update the cache with the new denom trace
	(c.ibcTraceCache)[denomStr] = res

	return res, nil
}
