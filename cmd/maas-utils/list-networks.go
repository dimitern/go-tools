package main

import (
	"fmt"

	"launchpad.net/gomaasapi"
)

func getNetworks(nets gomaasapi.MAASObject) []gomaasapi.JSONObject {
	result, err := nets.CallGet("", nil)
	if err != nil {
		fatalf("cannot get networks: %v", err)
	}

	list, err := result.GetArray()
	if err != nil {
		fatalf("cannot list networks: %v", err)
	}
	debugf("GetArray returned %d results", len(list))
	return list
}

func listNetworks() {
	ips := gomaasapi.NewMAAS(*getClient()).GetSubObject("networks")
	debugf("got networks endpoint, calling GET")

	for i, nw := range getNetworks(ips) {
		obj, err := nw.GetMAASObject()
		if err != nil {
			fatalf("cannot get network #%d: %v", i, err)
		}
		data, err := obj.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		fmt.Println(string(data))
	}
}
