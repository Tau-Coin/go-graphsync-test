package main

import (
	"context"
)

type GraphsyncContext struct {
	ctx	context.Context
}

func setupGSContext(ctx context.Context) (*GraphsyncContext, error) {
	gsCtx := &GraphsyncContext{
		ctx: ctx,
	}

	return gsCtx, nil
}
