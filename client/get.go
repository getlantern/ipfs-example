package main

import (
	"bytes"
	"io"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/path"
	uio "github.com/ipfs/go-ipfs/unixfs/io"
	"golang.org/x/net/context"
)

func Get(node *core.IpfsNode, ctx context.Context, pt string) string {
	p := path.Path(pt)
	dn, err := core.Resolve(ctx, node, p)
	if err != nil {
		panic(err)
	}

	reader, err := uio.NewDagReader(ctx, dn, node.DAG)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
