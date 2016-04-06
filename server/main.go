package main

import (
	"fmt"
	"time"

	//logging "gx/ipfs/Qmazh5oNUVsDZTs2g59rq8aYQqwpss8tcUWQzor5sCCEuH/go-log"

	"github.com/getlantern/ipfs-example/ipfs"
)

func main() {
	//logging.LevelInfo()
	node, err := ipfs.Start("../.ipfs-repo-server")
	if err != nil {
		panic(err)
	}
	defer node.Stop()

	interval := 10 * time.Second
	t := time.NewTimer(0)
	for i := 1; ; i++ {
		<-t.C
		msg := fmt.Sprintf("Message %d: Welcome to the wonderland of IPFS\n", i)
		path, _, err := node.Add(msg, "hello")
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d: Hey! I have a message to you at /ipfs/%s\n", i, path)

		ns, err := node.Publish(path)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Also available at permanent link /ipns/%s\n", ns)
		fmt.Println("***")
		t.Reset(interval)
	}
}
