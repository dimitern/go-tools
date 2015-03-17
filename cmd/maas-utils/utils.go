package main

import (
	"fmt"
	"net"
	"time"
)

type FieldsMap map[string]interface{}

func (m FieldsMap) StringField(name string, optional bool) (string, error) {
	val, ok := m[name]
	if !ok {
		if optional {
			return "", nil
		}
		return "", fmt.Errorf("required field %q missing", name)
	}
	tVal, ok := val.(string)
	if !ok {
		if optional && val == nil {
			return "", nil
		}
		return "", fmt.Errorf("expected field %q of type string, got %T", name, val)
	}
	return tVal, nil
}

func (m FieldsMap) IntField(name string, optional bool) (int, error) {
	val, ok := m[name]
	if !ok {
		if optional {
			return 0, nil
		}
		return 0, fmt.Errorf("required field %q missing", name)
	}
	// JSON numbers are always float64
	tVal, ok := val.(float64)
	if !ok {
		if optional && val == nil {
			return 0, nil
		}
		return 0, fmt.Errorf("expected field %q of type float64, got %T", name, val)
	}
	return int(tVal), nil
}

func (m FieldsMap) AddressField(name string, optional bool) (Address, error) {
	nothing := Address{}
	addr, err := m.StringField(name, optional)
	if err != nil {
		return nothing, err
	}
	if addr == "" && !optional {
		return nothing, fmt.Errorf("required field %q is empty", name)
	}
	return Address{
		IP:       net.ParseIP(addr),
		Hostname: addr,
	}, nil
}

func (m FieldsMap) TimeField(name string, optional bool) (time.Time, error) {
	nothing := time.Time{}
	val, err := m.StringField(name, optional)
	if err != nil {
		return nothing, err
	}
	if val == "" {
		if !optional {
			return nothing, fmt.Errorf("required field %q is empty", name)
		}
		return nothing, nil
	}
	return time.Parse(time.RFC3339, val+"Z")
}
