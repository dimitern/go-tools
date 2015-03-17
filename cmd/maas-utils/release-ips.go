package main

import (
	"net/url"

	"launchpad.net/gomaasapi"
)

func releaseIPs(maasRoot *gomaasapi.MAASObject) {
	var released, failed int
	allIPs := getIPs(maasRoot)
	ips := maasRoot.GetSubObject("ipaddresses")
	for _, ip := range allIPs {
		debugf("trying to release %q", ip.IP)

		params := make(url.Values)
		params.Set("ip", ip.IP.String())
		result, err := ips.CallPost("release", params)
		if err != nil {
			logf("cannot release %q: %v", ip.IP, err)
			failed++
			continue
		}
		released++
		debugf("result was %v", result)
		logf("IP %q released.", ip.IP)
	}
	if len(allIPs) > 0 {
		logf("%d IPs successfully released; %d failures", released, failed)
		return
	}
	logf("no allocated IPs to release.")
}
