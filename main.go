/*
    miner-resolver / Try to help witness miner to contact the PoC challenger
    Copyright (C) 2022  Sebastien Chaumontet

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main


import (
		"fmt"
		"regexp"
		"strconv"
		"log"
		"time"
		"io/ioutil"
		"net/http"
		"encoding/json"
		"os/exec"
		"github.com/hpcloud/tail"
)
/*

2022-02-26 18:42:00.336 25 [warning] <0.29487.1>@miner_onion_server:send_witness:{243,37} failed to dial challenger "/p2p/112i6wQDX7U2tAMJHFv3KuafE4278ctB2am1aFzdXBY5xX75j2TH": not_found
2022-02-26 18:42:30.338 25 [error] <0.29487.1>@miner_onion_server:send_witness:{207,5} failed to send witness, max retry
2022-02-26 18:38:39.425 25 [info] <0.29528.1>@miner_onion_server:send_witness:{251,37} successfully sent witness to challenger "/p2p/112dEUibhM6b9SYVEF1Nke7Ez41WZ4i7Tf5x23TD5GAWiAZvJU67" with RSSI: -98, Frequency: 867.1, SNR: -19.8

balena exec --interactive --tty $(balena ps --filter name=^helium-miner --format "{{.ID}}") miner peer ping
*/

const softwareName = "miner-resolver"
const softwareVersion = "0.0.0b3"

type WitnessStatus struct {
	address string
	date string
	count int
	actionFired bool
}


var ongoingWitness = map[string]WitnessStatus{}

func doMinerPing(addr string) ([]byte) {
	cmd := `balena exec --interactive $(balena ps --filter name=^helium-miner --format "{{.ID}}") miner peer ping `+addr+` 2>&1`
	fmt.Printf("exec command: %s\n",cmd)

	out, err := exec.Command(`bash`,`-c`,cmd).Output()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("exec command output: %s",out)
	return out
}

func getP2pAddrFromJson(HSjson []byte) (data []interface{},ok bool) {
	defer func() { // Prevent panic
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
			data = nil
			ok = false
		}
	}()

	var result map[string]interface{}

	json.Unmarshal(HSjson, &result)

	data, ok = ((result["data"].(map[string]interface{}))["status"].(map[string]interface{}))["listen_addrs"].([]interface{})

	return data,ok
}

func doApiRequest(address string) ([]byte) {
	//url := "https://api.helium.io/v1/hotspots/"
	url := "https://helium-api.stakejoy.com/v1/hotspots/"

	apiClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 5 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url + address, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "miner-resolver")

	res, getErr := apiClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return body
}

func main() {
	fmt.Println(softwareName+" "+softwareVersion+" Copyright (C) 2022 Sebastien Chaumontet - https://github.com/snip/miner-resolver")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("This is free software, and you are welcome to redistribute it under GPL v3.0 conditions.\n")

	//t, err := tail.TailFile("/mnt/data/docker/volumes/*_miner-log/_data/console.log", tail.Config{Follow: true})
	t, err := tail.TailFile("/mnt/data/docker/volumes/1804676_miner-log/_data/console.log", tail.Config{ReOpen:true, MustExist: false,Follow: true})

	reSendWitnessFailedDial := regexp.MustCompile(`\[warning\] <(.+)>@miner_onion_server:send_witness:.*failed to dial.*"/p2p/(\w+)".*`)
	reSendWitnessFailedSend := regexp.MustCompile(`\[error\] <(.+)>@miner_onion_server:send_witness:.* failed to send witness`)
	reSendWitnessSuccess := regexp.MustCompile(`\[info\] <(.+)>@miner_onion_server:send_witness:.* successfully sent witness to challenger`)
	reSuccessfully := regexp.MustCompile(`successfully`)
	// /p2p/112b9HMf5YSFeBnFJkv2mQnFECn91Q4xmWpsJt62WMXoqwbu1ucM/p2p-circuit/p2p/112Wnq8peTKWfQVei6GTkCGPmF2CuVhtMjgyBmz2aJizeJE9dgC6
	reRelayedAddress := regexp.MustCompile(`/p2p/(.+)/p2p-circuit/p2p/`)

	for line := range t.Lines {
		matches := reSendWitnessFailedDial.FindStringSubmatch(line.Text)
		if ( matches != nil ) { // failed to dial challenger
			fmt.Println(line.Text)
			if val, ok := ongoingWitness[matches[1]]; ok {
				fmt.Println("Already existing. Count: " + strconv.Itoa(val.count))
				val.count++
				if (val.count >= 2) && (val.actionFired == false) { // More than 2 retries
					fmt.Println("Do action with: " + val.address)
					json := doApiRequest(val.address)
					adresses, ok := getP2pAddrFromJson(json)
					if ok {
						for _, address := range adresses {
							fmt.Println("p2p addr from API: "+address.(string))
							matches := reRelayedAddress.FindStringSubmatch(address.(string))
							if (matches != nil) {
								fmt.Println("Relayed address detecter. We need to ping this relay first ("+matches[1]+")")
								json := doApiRequest(matches[1])
								adresses, ok := getP2pAddrFromJson(json)
								if ok {
									for _, address := range adresses { // try to ping all relay address
										fmt.Println("p2p relay addr from API: "+address.(string))
										resultmessage := string(doMinerPing(address.(string)))
										if (reSuccessfully.FindStringSubmatch(resultmessage) != nil) {
											fmt.Println("Successfull ping peer => end of ping relay work")
											break
										}
									}
								}
							}
							resultmessage := string(doMinerPing(address.(string)))
							if (reSuccessfully.FindStringSubmatch(resultmessage) != nil) {
								fmt.Println("Successfull ping peer => end of ping work")
								break
							}
						}
					} else {
						fmt.Println("No address found from API => Ignoring")
					}
					val.actionFired = true
				}
				ongoingWitness[matches[1]] = val;
			} else {
				fmt.Println("New witness")
				ongoingWitness[matches[1]] = WitnessStatus{
					matches[2]," ",1,false,
				}
			}
		}
		matches = reSendWitnessFailedSend.FindStringSubmatch(line.Text)
		if ( matches != nil ) {
			fmt.Println(line.Text)
			if _, ok := ongoingWitness[matches[1]]; ok {
				fmt.Println("Deleting entry")
				delete(ongoingWitness,matches[1])
			}
		}
		matches = reSendWitnessSuccess.FindStringSubmatch(line.Text)
		if ( matches != nil ) {
			fmt.Println(line.Text)
			if _, ok := ongoingWitness[matches[1]]; ok {
				fmt.Println("Deleting entry")
				delete(ongoingWitness,matches[1])
			}
		}
	}
	fmt.Println(err)
}
