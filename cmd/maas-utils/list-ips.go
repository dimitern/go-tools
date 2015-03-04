package main

import (
	"fmt"

	"launchpad.net/gomaasapi"
)

func getIPs(ips gomaasapi.MAASObject) []gomaasapi.JSONObject {
	result, err := ips.CallGet("", nil)
	if err != nil {
		fatalf("cannot get IPs: %v", err)
	}

	list, err := result.GetArray()
	if err != nil {
		fatalf("cannot list IPs: %v", err)
	}
	debugf("GetArray returned %d results", len(list))
	return list
}

func listIPs() {
	ips := gomaasapi.NewMAAS(*getClient()).GetSubObject("ipaddresses")
	debugf("got ipaddresses endpoint, calling GET")

	for i, ip := range getIPs(ips) {
		obj, err := ip.GetMAASObject()
		if err != nil {
			fatalf("cannot get IP #%d: %v", i, err)
		}
		data, err := obj.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		fmt.Println(string(data))
	}
}
