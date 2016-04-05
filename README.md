##ipfs-example

It's a slightly modified version of [Making your own ipfs service](https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/QmQwAP9vFjbCtKvD8RkJdCvPHqLQjZfW7Mqbbqx18zd8j7/api/service/readme.md) that can have server and client work on same machine.

### [Install IPFS from source](https://github.com/ipfs/go-ipfs/#download--compile-ipfs)

### Initialize IPFS repo for both server and client
```
IPFS_PATH=.ipfs-repo-server ipfs init
IPFS_PATH=.ipfs-repo-client ipfs init

IPFS_PATH=.ipfs-repo-client ipfs config edit
```
Then change the addresses in `Swarm` section to a different port.

### Run server and client
```
(cd server && go get && go run *.go)
# A moment after the IPFS link get printed
(cd client && go get && go run *.go <IPFS link>)
```
