package main

import (
	"context"

	ipld "github.com/ipld/go-ipld-prime"
)

// setup state hamt tree, and return the root link or some error
func setupStateTree(ctx context.Context) (ipld.Link, error) {
	return nil, nil
}

func stateSelector() ipld.Node {
	return nil
}
