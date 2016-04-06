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
	interval := 10 * time.Second
	t := time.NewTimer(0)
	for i := 1; ; i++ {
		<-t.C
		msg := fmt.Sprintf("Message %d: Welcome to the wonderland of IPFS\n", i)
		path, _, err := Add(nd, msg, "hello")
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d: Hey! I have a message to you at /ipfs/%s\n", i, path)

		ns, err := publish(ctx, nd, path)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Also available at permanent link /ipns/%s\n", ns)
		fmt.Println("***")
		t.Reset(interval)
	}
}
