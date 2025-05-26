package main

import (
	"context"
	"github.com/gburgyan/go-quickgraph"
	"github.com/patrickmn/go-cache"
	"time"
)

type SimpleGraphRequestCache struct {
	cache *cache.Cache
}

type simpleGraphRequestCacheEntry struct {
	request string
	stub    *quickgraph.RequestStub
	err     error
}

func (d *SimpleGraphRequestCache) SetRequestStub(ctx context.Context, request string, stub *quickgraph.RequestStub, err error) {
	setErr := d.cache.Add(request, &simpleGraphRequestCacheEntry{
		request: request,
		stub:    stub,
		err:     err,
	}, time.Hour)
	if setErr != nil {
		// Log this error, but don't return it.
		// Potentially disable the cache if this recurs continuously.
	}
}

func (d *SimpleGraphRequestCache) GetRequestStub(ctx context.Context, request string) (*quickgraph.RequestStub, error) {
	value, found := d.cache.Get(request)
	if !found {
		return nil, nil
	}
	entry, ok := value.(*simpleGraphRequestCacheEntry)
	if !ok {
		return nil, nil
	}
	return entry.stub, entry.err
}
