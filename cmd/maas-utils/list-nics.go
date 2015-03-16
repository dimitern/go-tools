package main

import (
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

func getNICs(nodeGroups map[string]gomaasapi.MAASObject, uuidNG string) []gomaasapi.JSONObject {
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
	return list
}

func listNICs() {
	client := getClient()
	maas := gomaasapi.NewMAAS(*client)
	ng := maas.GetSubObject("nodegroups")
	debugf("got nodegroups and nodegroupinterfaces endpoints, calling GET")

	var uuids []string
	fmt.Printf("listing all node groups\n\n")
	nodeGroups := getNodeGroups(ng)
	for uuid, ngObj := range nodeGroups {
		uuids = append(uuids, uuid)
		data, err := ngObj.MarshalJSON()
		if err != nil {
			fatalf("serializing to JSON failed: %v", err)
		}
		fmt.Println(string(data))
	}
	fmt.Printf("\nlisting all NICs for all node groups %v\n", uuids)
	for _, uuid := range uuids {
		fmt.Printf("\nnode group %q NICs:\n\n", uuid)
		for i, nic := range getNICs(nodeGroups, uuid) {
			data, err := nic.MarshalJSON()
			if err != nil {
				fatalf("serializing NIC #%d to JSON failed: %v", i, err)
			}
			fmt.Println(string(data))
		}
	}
}
