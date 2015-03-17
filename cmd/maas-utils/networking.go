package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/juju/errors"
)

// Address describes an IP address or hostname.
type Address struct {
	IP       net.IP
	Hostname string
}

func (a Address) String() string {
	if a.IP != nil {
		return a.IP.String()
	}
	return a.Hostname
}

func (a Address) IsEmpty() bool {
	return a.IP == nil && a.Hostname == ""
}

// Addresses defines a list of Address entries.
type Addresses []Address

func (a Addresses) String() string {
	addrs := make([]string, len(a))
	for i, addr := range a {
		addrs[i] = fmt.Sprintf("%q", addr)
	}
	return fmt.Sprintf("[%s]", strings.Join(addrs, ", "))
}

// Network describes a MAAS network.
type Network struct {
	Name        string
	Description string
	Netmask     net.IPMask
	VLANTag     int
	DNSServers  Addresses
	IP          Address
	Gateway     Address
}

func (n *Network) UnmarshalJSON(data []byte) error {
	fields := make(FieldsMap)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var err error
	n.Name, err = fields.StringField("name", false)
	if err != nil {
		return err
	}
	n.Description, err = fields.StringField("description", true)
	if err != nil {
		return err
	}
	n.VLANTag, err = fields.IntField("vlan_tag", true)
	if err != nil {
		return err
	}
	dnsServers, err := fields.StringField("dns_servers", true)
	if err != nil {
		return err
	}
	if dnsServers != "" {
		for _, srv := range strings.Split(dnsServers, " ") {
			srv = strings.TrimSpace(srv)
			ip := net.ParseIP(srv)
			if ip != nil {
				n.DNSServers = append(n.DNSServers, Address{IP: ip})
			} else {
				n.DNSServers = append(n.DNSServers, Address{Hostname: srv})
			}
		}
	}
	n.Gateway, err = fields.AddressField("default_gateway", true)
	if err != nil {
		return err
	}
	n.Netmask, err = fields.NetmaskField("netmask", false)
	if err != nil {
		return err
	}
	n.IP, err = fields.AddressField("ip", false)
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) GoString() string {
	return fmt.Sprintf(
		"Network{Name: %q, Description: %q, IP: %q, Netmask: %q, DNSServers: %s, Gateway: %q, VLANTag: %v}",
		n.Name, n.Description, n.IP, n.Netmask, n.DNSServers, n.Gateway, n.VLANTag,
	)
}

func (n *Network) String() string {
	return fmt.Sprintf("network %q (%s/%s)", n.Name, n.IP, n.Netmask)
}

// ManagementType describes the way MAAS manages an interface.
type ManagementType int

const (
	Unmanaged ManagementType = iota
	ManageDHCPOnly
	ManageDNSAndDHCP
)

func (m ManagementType) String() string {
	switch m {
	case Unmanaged:
		return "Unmanaged"
	case ManageDHCPOnly:
		return "ManageDHCPOnly"
	case ManageDNSAndDHCP:
		return "ManageDNSAndDHCP"
	}
	return fmt.Sprintf("<unknown: %d>", m)
}

// Interface describes a MAAS node group interface.
type Interface struct {
	ClusterID         string
	Name              string
	Interface         string
	RouterIP          Address
	BroadcastIP       Address
	Netmask           net.IPMask
	DHCPRangeLowIP    Address
	DHCPRangeHighIP   Address
	StaticRangeLowIP  Address
	StaticRangeHighIP Address
	Management        ManagementType
}

func (i *Interface) UnmarshalJSON(data []byte) error {
	fields := make(FieldsMap)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var err error
	i.Name, err = fields.StringField("name", false)
	if err != nil {
		return err
	}
	i.Interface, err = fields.StringField("interface", false)
	if err != nil {
		return err
	}
	i.RouterIP, err = fields.AddressField("ip", false)
	if err != nil {
		return err
	}
	i.BroadcastIP, err = fields.AddressField("broadcast_ip", true)
	if err != nil {
		return err
	}
	i.Netmask, err = fields.NetmaskField("subnet_mask", true)
	if err != nil {
		return err
	}
	i.DHCPRangeLowIP, err = fields.AddressField("ip_range_low", true)
	if err != nil {
		return err
	}
	i.DHCPRangeHighIP, err = fields.AddressField("ip_range_high", true)
	if err != nil {
		return err
	}
	i.StaticRangeLowIP, err = fields.AddressField("static_ip_range_low", true)
	if err != nil {
		return err
	}
	i.StaticRangeHighIP, err = fields.AddressField("static_ip_range_high", true)
	if err != nil {
		return err
	}
	mgmt, err := fields.IntField("management", false)
	if err != nil {
		return err
	}
	i.Management = ManagementType(mgmt)
	return nil
}

func (i *Interface) HasStaticRange() bool {
	return !i.StaticRangeLowIP.IsEmpty() && !i.StaticRangeHighIP.IsEmpty()
}

func (i *Interface) GoString() string {
	return fmt.Sprintf(
		"Interface{ClusterID: %q, Name: %q, Interface: %q, RouterIP: %q, BroadcastIP: %q, Netmask: %q, Management: %q, DHCPRangeLowIP: %q, DHCPRangeHighIP: %q, StaticRangeLowIP: %q, StaticRangeHighIP: %q}",
		i.ClusterID, i.Name, i.Interface, i.RouterIP, i.BroadcastIP, i.Netmask, i.Management,
		i.DHCPRangeLowIP, i.DHCPRangeHighIP, i.StaticRangeLowIP, i.StaticRangeHighIP,
	)
}

func (i *Interface) String() string {
	return fmt.Sprintf("interface %q (%s/%s)", i.Interface, i.RouterIP, i.Netmask)
}

// AllocationType describes a StaticIP allocation type used by MAAS.
type AllocationType int

const (
	AllocAuto         AllocationType = 0
	AllocSticky       AllocationType = 1
	AllocUserReserved AllocationType = 4
)

func (a AllocationType) String() string {
	switch a {
	case AllocAuto:
		return "Auto"
	case AllocSticky:
		return "Sticky"
	case AllocUserReserved:
		return "UserReserved"
	}
	return fmt.Sprintf("<unknown: %d>", a)
}

// StaticIP describes a static IP address in MAAS.
type StaticIP struct {
	AllocType AllocationType
	Created   time.Time
	IP        Address
}

func (s *StaticIP) UnmarshalJSON(data []byte) error {
	fields := make(FieldsMap)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	allocType, err := fields.IntField("alloc_type", false)
	s.AllocType = AllocationType(allocType)
	s.IP, err = fields.AddressField("ip", false)
	if err != nil {
		return err
	}
	s.Created, err = fields.TimeField("created", false)
	if err != nil {
		return err
	}
	return nil
}

func (s *StaticIP) GoString() string {
	return fmt.Sprintf(
		"StaticIP{AllocType: %q, Created: %q, IP: %q}",
		s.AllocType, s.Created, s.IP,
	)
}

func (s *StaticIP) String() string {
	return fmt.Sprintf("static IP address %q", s.IP)
}

// ParseIPv4Mask parses a given IPv4 netmask in dotted quad format (e.g. "255.255.240.0").
func ParseIPv4Mask(mask string) (net.IPMask, error) {
	parts := strings.Split(strings.TrimSpace(mask), ".")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid IPv4 netmask: %v", mask)
	}
	bytes := make([]byte, len(parts))
	for i, part := range parts {
		npart, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid IPv4 netmask %v: %v", mask, err)
		}
		bytes[i] = byte(npart)
	}
	return net.IPMask(bytes), nil
}

// DecimalToIPv4 converts a decimal to the dotted quad IP address format.
func DecimalToIPv4(addr uint32) net.IP {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, addr)
	return net.IP(bytes)
}

// IPv4ToDecimal converts a dotted quad IP address to its decimal equivalent.
func IPv4ToDecimal(ipv4Addr net.IP) (uint32, error) {
	ip := ipv4Addr.To4()
	if ip == nil {
		return 0, errors.Errorf("%q is not a valid IPv4 address", ipv4Addr.String())
	}
	return binary.BigEndian.Uint32([]byte(ip)), nil
}
