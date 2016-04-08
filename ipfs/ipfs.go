package ipfs

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/ipfs/go-ipfs/blocks/key"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/corenet"
	"github.com/ipfs/go-ipfs/core/coreunix"
	dag "github.com/ipfs/go-ipfs/merkledag"
	"github.com/ipfs/go-ipfs/path"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	uio "github.com/ipfs/go-ipfs/unixfs/io"
	"golang.org/x/net/context"
	crypto "gx/ipfs/QmSN2ELGRp4T9kjqiSsSNJRUeR9JKXzQEgwe1HH3tdSGbC/go-libp2p/p2p/crypto"
)

type IpfsNode struct {
	node   *core.IpfsNode
	pk     crypto.PrivKey
	ctx    context.Context
	cancel context.CancelFunc
}

func Start(repoDir string, pkfile string) (*IpfsNode, error) {
	r, err := fsrepo.Open(repoDir)
	if err != nil {
		return nil, err
	}

	var pk crypto.PrivKey
	if pkfile != "" {
		pk, err = GenKeyIfNotExists(pkfile)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}

	nd, err := core.NewNode(ctx, cfg)

	if err != nil {
		return nil, err
	}
	_, err = corenet.Listen(nd, "/app/lantern")
	if err != nil {
		return nil, err
	}
	return &IpfsNode{nd, pk, ctx, cancel}, nil
}

func (node *IpfsNode) Stop() {
	node.cancel()
}

func (node *IpfsNode) Add(content string, name string) (path string, dNode *dag.Node, err error) {
	return coreunix.AddWrapped(node.node, strings.NewReader(content), name)
}

func (node *IpfsNode) AddFile(fileName string, name string) (path string, dNode *dag.Node, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", nil, err
	}
	return coreunix.AddWrapped(node.node, file, name)
}

func (node *IpfsNode) Get(pt string) (string, error) {
	p := path.Path(pt)
	dn, err := core.Resolve(node.ctx, node.node, p)
	if err != nil {
		return "", err
	}

	reader, err := uio.NewDagReader(node.ctx, dn, node.node.DAG)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (node *IpfsNode) Publish(p string) (string, error) {
	ref := path.Path(p)
	k := node.node.PrivateKey
	if node.pk != nil {
		k = node.pk
	}
	err := node.node.Namesys.Publish(node.ctx, k, ref)
	if err != nil {
		return "", err
	}

	hash, err := k.GetPublic().Hash()
	if err != nil {
		return "", err
	}

	return key.Key(hash).String(), nil
}

func (node *IpfsNode) Resolve(name string) (string, error) {
	p, err := node.node.Namesys.ResolveN(node.ctx, name, 1)
	if err != nil {
		return "", err
	}

	return p.String(), nil
}
