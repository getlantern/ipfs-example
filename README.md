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
(cd server && go get && go run *.go) # Will print peer id
(cd client && go get && go run *.go <peer id>)
```
