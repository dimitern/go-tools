package main

import (
	"encoding/json"
	"fmt"

	"launchpad.net/gomaasapi"
)

func getNodeGroupsUUIDs(maasRoot *gomaasapi.MAASObject) []string {
	ng := maasRoot.GetSubObject("nodegroups")
	result, err := ng.CallGet("list", nil)
	if err != nil {
		fatalf("cannot get node groups: %v", err)
	}
	nodeGroups, err := result.GetArray()
	if err != nil {
		fatalf("cannot list node groups: %v", err)
	}
	debugf("GetArray returned %d results", len(nodeGroups))
	uuids := make([]string, len(nodeGroups))
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
		uuids[i] = sUUID
	}
	return uuids
}

func getNICs(maasRoot *gomaasapi.MAASObject, uuidNG string) []Interface {
	ngi := maasRoot.GetSubObject("nodegroups").GetSubObject(uuidNG).GetSubObject("interfaces")
	result, err := ngi.CallGet("list", nil)
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

func listNICs(maasRoot *gomaasapi.MAASObject) {
	debugf("getting all node groups UUIDs")
	uuids := getNodeGroupsUUIDs(maasRoot)
	logf("listing all NICs for node groups: %v\n", uuids)
	for _, uuid := range uuids {
		for _, nic := range getNICs(maasRoot, uuid) {
			fmt.Println(nic.GoString(), "\n")
		}
	}
}
