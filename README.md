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
wget https://github.com/snip/miner-resolver/releases/latest/download/miner-resolver_arm64 -O /tmp/miner-resolver_arm64
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


## Example of success execution
```
2022-02-27 12:36:05.459 25 [warning] <0.28578.3>@miner_onion_server:send_witness:{243,37} failed to dial challenger "/p2p/112NTMRPGUYdiYMyVGwX1QWNJspjLmGwZLzJDRpQdLjNoEEMLZCS": not_found
New witness
2022-02-27 12:36:35.591 25 [warning] <0.28578.3>@miner_onion_server:send_witness:{243,37} failed to dial challenger "/p2p/112NTMRPGUYdiYMyVGwX1QWNJspjLmGwZLzJDRpQdLjNoEEMLZCS": not_found
Already existing. Count: 1
Do action with: 112NTMRPGUYdiYMyVGwX1QWNJspjLmGwZLzJDRpQdLjNoEEMLZCS
p2p addr from API: /ip4/173.176.240.170/tcp/44158
exec command: balena exec --interactive $(balena ps --filter name=^helium-miner --format "{{.ID}}") miner peer ping /ip4/173.176.240.170/tcp/44158 2>&1
exec command output: Pinged "/ip4/173.176.240.170/tcp/44158" successfully with roundtrip time: 144 ms

2022-02-27 12:37:06.186 25 [info] <0.28578.3>@miner_onion_server:send_witness:{251,37} successfully sent witness to challenger "/p2p/112NTMRPGUYdiYMyVGwX1QWNJspjLmGwZLzJDRpQdLjNoEEMLZCS" with RSSI: -92, Frequency: 867.1, SNR: -15.2
Deleting entry
```
