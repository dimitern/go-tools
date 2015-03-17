package main

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/url"
	"time"

	"launchpad.net/gomaasapi"
)

func reserveIP(maasRoot *gomaasapi.MAASObject, netName, ipAddr string) {
	if netName == "" {
		fatalf("network name is required but missing")
	}
	debugf("listing all networks")
	networks := getNetworks(maasRoot)
	nw, ok := networks[netName]
	if !ok {
		fatalf("unknown network %q", netName)
	}
	netIP := net.ParseIP(nw.IP.String())
	if netIP == nil {
		fatalf("unexpected address format %v for network %q", nw.IP, netName)
	}
	ipNet := net.IPNet{IP: netIP, Mask: nw.Netmask}
	debugf("trying to use network %q, finding static range", netName)

	if ipAddr != "" && ipAddr != "random" {
		ip := net.ParseIP(ipAddr)
		if ip == nil {
			fatalf("invalid IP address to reserve on network %q: %v", netName, ipAddr)
		}
		if !ipNet.Contains(ip) {
			fatalf("IP address %q not within network %q range %q", ipAddr, netName, ipNet.String())
		}
	}
	ngUUIDs := getNodeGroupsUUIDs(maasRoot)
	if len(ngUUIDs) == 0 {
		fatalf("no node groups defined")
	}
	debugf("got all node groups UUIDs: %v; matching by network", ngUUIDs)
	var foundNIC Interface
	for _, uuid := range ngUUIDs {
		nics := getNICs(maasRoot, uuid)
		if len(nics) == 0 {
			debugf("skipping node group %q - no interfaces")
			continue
		}
		for _, nic := range nics {
			ip := nic.RouterIP.String()
			nicIP := net.ParseIP(ip)
			if nicIP == nil {
				debugf(
					"skipping interface %q on node group %q - unexpected IP %v",
					nic.Name, uuid, ip,
				)
				continue
			}
			if !ipNet.Contains(nicIP) {
				debugf(
					"skipping interface %q on node group %q - IP %q not within network %q range",
					nic.Name, uuid, nicIP, netName,
				)
				continue
			}
			if !nic.HasStaticRange() {
				fatalf(
					"interface %q on node group %q matches network %q but has no static range",
					nic.Name, uuid, netName,
				)
			}
			// Found it
			debugf("matched network %q to interface %q on node group %q", netName, nic.Name, uuid)
			foundNIC = nic
			break
		}
	}

	if foundNIC.Name == "" {
		fatalf("cannot find any node group interfaces matching network %q", netName)
	}

	var ipArg string
	switch ipAddr {
	case "":
		logf("trying to reserve an IP address on network %q", netName)
	case "random":
		ip := foundNIC.StaticRangeLowIP.IP
		decLow, err := IPv4ToDecimal(ip)
		if err != nil {
			fatalf("cannot convert static range lower bound %q to decimal: %v", ip, err)
		}
		ip = foundNIC.StaticRangeHighIP.IP
		decHigh, err := IPv4ToDecimal(ip)
		if err != nil {
			fatalf("cannot convert static range higher bound %q to decimal: %v", ip, err)
		}
		totalAddressesInRange := decHigh - decLow
		newDecimal := decLow + uint32(random.Intn(int(totalAddressesInRange)))
		newIP := DecimalToIPv4(newDecimal)
		if newIP == nil {
			fatalf("generated random IP %v is invalid", newIP)
		}
		ipArg = newIP.String()
		logf("trying to reserve a random IP address (%q) on network %q", ipArg, netName)
	default:
		ipArg = ipAddr
		logf("trying to reserve IP address %q on network %q", ipAddr, netName)
	}

	ips := maasRoot.GetSubObject("ipaddresses")
	params := make(url.Values)
	params.Set("network", ipNet.String())
	if ipArg != "" {
		params.Set("requested_address", ipArg)
	}
	logf("calling POST %s with op=reserve and params %v", ips.URL(), params)
	result, err := ips.CallPost("reserve", params)
	if err != nil {
		fatalf("MAAS returned: %v", err)
	}
	debugf("result was %v", result)
	data, err := result.MarshalJSON()
	if err != nil {
		fatalf("serializing to JSON failed: %v", err)
	}
	var staticIP StaticIP
	if err := json.Unmarshal(data, &staticIP); err != nil {
		fatalf("deserializing from JSON failed: %v", err)
	}
	if staticIP.IP.String() != ipArg && ipArg != "" {
		fatalf("tried to allocate %q, but MAAS returned %q", ipArg, staticIP.IP)
	}
	logf("allocated IP address %q on network %q successfully.", staticIP, netName)

	listIPs(maasRoot)
}

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().Unix()))
}
