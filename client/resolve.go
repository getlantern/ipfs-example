package main

import (
	"github.com/ipfs/go-ipfs/core"
	"golang.org/x/net/context"
)

func resolve(node *core.IpfsNode, ctx context.Context, name string) (string, error) {
	p, err := node.Namesys.ResolveN(ctx, name, 1)
	if err != nil {
		return "", err
	}

	return p.String(), nil
}
