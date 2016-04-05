package main

import (
	"fmt"
	"time"

	//logging "gx/ipfs/Qmazh5oNUVsDZTs2g59rq8aYQqwpss8tcUWQzor5sCCEuH/go-log"

	core "github.com/ipfs/go-ipfs/core"
	corenet "github.com/ipfs/go-ipfs/core/corenet"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

func main() {
	//logging.LevelInfo()
	// Basic ipfsnode setup
	r, err := fsrepo.Open("../.ipfs-repo-server")
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
	_, err = corenet.Listen(nd, "/app/lantern")
	if err != nil {
		panic(err)
	}
	msg := fmt.Sprintf("%s: Hello! Lantern welcome you to the wonderland of IPFS\n", time.Now())
	path, _, err := Add(nd, msg, "hello")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hey! I have a message to you at /ipfs/%s\n", path)
	var ch chan struct{}
	<-ch
}
