package main

import (
	"fmt"
	"os"

	//logging "gx/ipfs/Qmazh5oNUVsDZTs2g59rq8aYQqwpss8tcUWQzor5sCCEuH/go-log"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

func main() {
	//logging.LevelInfo()
	if len(os.Args) < 2 {
		fmt.Println("Please give IPFS path")
		return
	}
	// Basic ipfsnode setup
	r, err := fsrepo.Open("../.ipfs-repo-client")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}

	nd, err := core.NewNode(ctx, cfg)

	if err != nil {
		panic(err)
	}

	path := os.Args[1]
	s := Get(nd, ctx, path)
	fmt.Printf("Hello at %s: %s\n", path, s)
}
