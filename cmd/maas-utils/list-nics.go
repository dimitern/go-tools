package main

import (
	"encoding/json"
	"fmt"

	"launchpad.net/gomaasapi"
)

func getNodeGroups(ng gomaasapi.MAASObject) map[string]gomaasapi.MAASObject {
	result, err := ng.CallGet("list", nil)
	if err != nil {
		fatalf("cannot get node groups: %v", err)
	}

	nodeGroups, err := result.GetArray()
	if err != nil {
		fatalf("cannot list node groups: %v", err)
	}
	debugf("GetArray returned %d results", len(nodeGroups))
	ng2obj := make(map[string]gomaasapi.MAASObject, len(nodeGroups))
	for i, ngroup := range nodeGroups {
		objMap, err := ngroup.GetMap()
		if err != nil {
			fatalf("cannot get node group #%d object map: %v", i, err)
		}
		uuid, ok := objMap["uuid"]
		if !ok {
			fatalf("cannot get node group #%d UUID from %v", i, objMap)
		}
		sUUID, err := uuid.GetString()
		if err != nil {
			fatalf("cannot get node group #%d UUID as string: %v", i, err)
		}
		ng2obj[sUUID] = ng.GetSubObject(sUUID)
	}
	return ng2obj
}

func getNICs(nodeGroups map[string]gomaasapi.MAASObject, uuidNG string) []Interface {
	nodeGroup, ok := nodeGroups[uuidNG]
	if !ok {
		fatalf("cannot find node group %q in %v", uuidNG, nodeGroups)
	}
	result, err := nodeGroup.GetSubObject("interfaces").CallGet("list", nil)
	if err != nil {
		fatalf("cannot get node group %q interfaces: %v", uuidNG, err)
	}

	list, err := result.GetArray()
	if err != nil {
		fatalf("cannot list node group %q interfaces: %v", uuidNG, err)
	}
	debugf("GetArray returned %d results", len(list))
	nics := make([]Interface, len(list))
	for i, nic := range list {
		data, err := nic.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		var iface Interface
		if err := json.Unmarshal(data, &iface); err != nil {
			fatalf("deserializing from JSON failed: %v", err)
		}
		iface.ClusterID = uuidNG
		nics[i] = iface
	}
	return nics
}

func listNICs() {
	client := getClient()
	maas := gomaasapi.NewMAAS(*client)
	ng := maas.GetSubObject("nodegroups")
	debugf("got nodegroups and nodegroupinterfaces endpoints, calling GET")

	var uuids []string
	logf("listing all node groups\n\n")
	nodeGroups := getNodeGroups(ng)
	for uuid, ngObj := range nodeGroups {
		uuids = append(uuids, uuid)
		data, err := ngObj.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		fmt.Println(string(data))
	}
	logf("\nlisting all NICs for all node groups %v\n", uuids)
	for _, uuid := range uuids {
		logf("\nnode group %q NICs:\n\n", uuid)
		for _, nic := range getNICs(nodeGroups, uuid) {
			fmt.Println(nic.GoString(), "\n")
		}
	}
}
