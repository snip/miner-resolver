# miner-resolver
Try to help Helium witness miner to contact the PoC challenger

## Concept
This tool is checking from an Helium full miner if it got issues contacting challenger.
In this case, after few miner normal try, this tool is going to ask Helium public API the known p2p adress for this challenger.
And then will try to ping this challenger with this p2p address.
If ping is success, it will be added to local miner peer table. And then the miner will be able to send the witness report to the PoC challenger

## Installation
From the miner get relased binary:
```
wget xxxx -O /tmp/miner-resolver_arm64
chmod +x /tmp/miner-resolver_arm64
```

## Usage
From the miner:
```
/tmp/miner-resolver_arm64
```


## Building

Normal build:
```
go build -o miner-resolver
```

Cross compiling for miner:
```
env GOOS=linux GOARCH=arm64 go build -o miner-resolver_arm64
```
