package main

import (
	"encoding/json"
	"fmt"

	"launchpad.net/gomaasapi"
)

func getIPs(maasRoot *gomaasapi.MAASObject) []StaticIP {
	ipaddrs := maasRoot.GetSubObject("ipaddresses")
	result, err := ipaddrs.CallGet("", nil)
	if err != nil {
		fatalf("cannot get IPs: %v", err)
	}

	list, err := result.GetArray()
	if err != nil {
		fatalf("cannot list IPs: %v", err)
	}
	debugf("GetArray returned %d results", len(list))
	ips := make([]StaticIP, len(list))
	for i, ip := range list {
		data, err := ip.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		var staticIP StaticIP
		if err := json.Unmarshal(data, &staticIP); err != nil {
			fatalf("deserializing from JSON failed: %v", err)
		}
		ips[i] = staticIP
	}
	return ips
}

func listIPs(maasRoot *gomaasapi.MAASObject) {
	allIPs := getIPs(maasRoot)
	logf("listing %d static IPs in MAAS:\n", len(allIPs))
	for _, ip := range allIPs {
		fmt.Println(ip.GoString(), "\n")
	}
}
