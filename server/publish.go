package main

import (
	"golang.org/x/net/context"

	"github.com/ipfs/go-ipfs/blocks/key"
	core "github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/path"
)

func publish(ctx context.Context, n *core.IpfsNode, p string) (string, error) {
	ref := path.Path(p)
	k := n.PrivateKey
	err := n.Namesys.Publish(ctx, k, ref)
	if err != nil {
		return "", err
	}

	hash, err := k.GetPublic().Hash()
	if err != nil {
		return "", err
	}

	return key.Key(hash).String(), nil
}
