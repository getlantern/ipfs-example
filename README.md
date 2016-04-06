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

### Run server
```
(cd server && go get && go run *.go)
```

It will add a new message every 10 seconds. Note that the IPFS link changed for each message but IPNS link didn't.

```
1: Hey! I have a message to you at /ipfs/QmaEtVYrLa3GGA3VkTek9ABH2mZxuTQ2TdAGhbgLC3gAqc/hello
Also available at permanent link /ipns/QmPifCgEXvisnQ8vBXF4MKvkqsJ5H8bP9NZVpLpL4VaGS9
***
2: Hey! I have a message to you at /ipfs/QmR74gHBS8yRfhv6y1JdLCzWevScaFB5hGyLiVHBw7mgM3/hello
Also available at permanent link /ipns/QmPifCgEXvisnQ8vBXF4MKvkqsJ5H8bP9NZVpLpL4VaGS9
***
...
```
### Run client
```
(cd client && go get && go run *.go <link>)
```

The link can be IPFS link or IPNS link. If you supply IPNS link, the retrieved message will vary gradually.

```
Real path for /ipns/QmPifCgEXvisnQ8vBXF4MKvkqsJ5H8bP9NZVpLpL4VaGS9: /ipfs/QmZYcPJUDShoFcyef6LDfTxcQ6Y9m1rWT43u8VRMXXthnG/hello
Message 295: Welcome to the wonderland of IPFS

Real path for /ipns/QmPifCgEXvisnQ8vBXF4MKvkqsJ5H8bP9NZVpLpL4VaGS9: /ipfs/Qmb9LYjQjwCY33zgwPJ3XyGH48GbLZpvT2YAxy2RaffvwV/hello
Message 299: Welcome to the wonderland of IPFS
...
```
