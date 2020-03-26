package main

import (
	ma "github.com/multiformats/go-multiaddr"
)

func isRelay(addrs []ma.Multiaddr) bool {
	for _, addr := range addrs {
		// split /a/p2p-circuit/b into (/a, /p2p-circuit/b)
		_, destaddr := ma.SplitFunc(addr, func(c ma.Component) bool {
			return c.Protocol().Code == ma.P_CIRCUIT
		})

		if destaddr != nil {
			return true
		}
	}

	return false
}
