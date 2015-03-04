package main

import (
	"net/url"

	"launchpad.net/gomaasapi"
)

func releaseIPs() {
	ips := gomaasapi.NewMAAS(*getClient()).GetSubObject("ipaddresses")
	debugf("got ipaddresses endpoint, calling GET")

	var released, failed int
	allIPs := getIPs(ips)
	for i, ip := range allIPs {
		obj, err := ip.GetMAASObject()
		if err != nil {
			fatalf("cannot get IP #%d: %v", i, err)
		}
		sip, err := obj.GetField("ip")
		if err != nil {
			fatalf("cannot get field 'ip' of IP #%d: %v", i, err)
		}
		debugf("trying to release %q", sip)

		params := make(url.Values)
		params.Set("ip", sip)
		result, err := obj.CallPost("release", params)
		if err != nil {
			logf("cannot release %q: %v", sip, err)
			failed++
			continue
		}
		released++
		debugf("result was %v", result)
		logf("IP %q released.", sip)
	}
	if len(allIPs) > 0 {
		logf("%d IPs successfully released; %d failures", released, failed)
		return
	}
	logf("no allocated IPs to release.")
}
