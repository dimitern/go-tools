package main

import (
	"encoding/json"
	"fmt"

	"launchpad.net/gomaasapi"
)

func getNetworks(maasRoot *gomaasapi.MAASObject) map[string]Network {
	nets := maasRoot.GetSubObject("networks")
	result, err := nets.CallGet("", nil)
	if err != nil {
		fatalf("cannot get networks: %v", err)
	}

	list, err := result.GetArray()
	if err != nil {
		fatalf("cannot list networks: %v", err)
	}
	debugf("GetArray returned %d results", len(list))
	networks := make(map[string]Network, len(list))
	for i, nw := range list {
		obj, err := nw.GetMAASObject()
		if err != nil {
			fatalf("cannot get network #%d: %v", i, err)
		}
		data, err := obj.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		var network Network
		if err := json.Unmarshal(data, &network); err != nil {
			fatalf("deserializing from JSON failed: %v", err)
		}
		networks[network.Name] = network
	}
	return networks
}

func listNetworks(maasRoot *gomaasapi.MAASObject) {
	nws := getNetworks(maasRoot)
	logf("listing %d networks in MAAS:\n", len(nws))
	for _, nw := range nws {
		fmt.Println(nw.GoString(), "\n")
	}
}
