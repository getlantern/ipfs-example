package main

import (
	"strings"

	core "github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	dag "github.com/ipfs/go-ipfs/merkledag"
)

func Add(node *core.IpfsNode, content string, file string) (path string, dNode *dag.Node, err error) {
	return coreunix.AddWrapped(node, strings.NewReader(content), file)
}
