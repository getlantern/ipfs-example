package main

import (
	"fmt"
	"os"
	"time"

	//logging "gx/ipfs/Qmazh5oNUVsDZTs2g59rq8aYQqwpss8tcUWQzor5sCCEuH/go-log"

	"github.com/getlantern/ipfs-example/ipfs"
)

func main() {
	//logging.LevelInfo()
	if len(os.Args) < 2 {
		fmt.Println("Please give IPFS path")
		return
	}
	node, err := ipfs.Start("../.ipfs-repo-client", "")
	if err != nil {
		panic(err)
	}
	defer node.Stop()

	path := os.Args[1]
	interval := 10 * time.Second
	t := time.NewTimer(0)
	for {
		<-t.C
		t.Reset(interval)
		realPath, err := node.Resolve(path)
		if err != nil {
			fmt.Printf("resolve: %s\n", err)
			continue
		}
		fmt.Printf("Real path for %s: %s\n", path, realPath)
		s, err := node.Get(realPath)
		if err != nil {
			fmt.Printf("get: %s\n", err)
			continue
		}
		fmt.Println(s)
	}
}
