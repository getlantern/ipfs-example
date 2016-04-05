##ipfs-example

It's a slightly modified version of [Making your own ipfs service](https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/QmQwAP9vFjbCtKvD8RkJdCvPHqLQjZfW7Mqbbqx18zd8j7/api/service/readme.md) that can have server and client work on same machine.

### Install prerequisites

* [Install IPFS from source](https://github.com/ipfs/go-ipfs/#download--compile-ipfs)
* Install repo migration tool
```
go get -u github.com/ipfs/fs-repo-migrations
```

### Initialize IPFS repo for both server and client
```
IPFS_PATH=.ipfs-repo-server ipfs init && fs-repo-migrations
IPFS_PATH=.ipfs-repo-client ipfs init && fs-repo-migrations
sed -i .prev "s/4001/4002/" .ipfs-repo-client/config # client should listen on different port than server
```

### Run server and client
```
(cd server && go get && go run *.go)
# A moment after the peer id get printed
(cd client && go get && go run *.go <peer id>)
```
