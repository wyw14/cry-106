package model

import (
	"fmt"
	"time"
)

type DeviceKind string

const (
	DeviceFan       DeviceKind = "fan"
	DeviceDoor      DeviceKind = "door"
	DeviceValve     DeviceKind = "valve"
	DeviceFeeder    DeviceKind = "feeder"
	DeviceGasSensor DeviceKind = "gas_sensor"
)

type DeviceCommand struct {
	ID              string     `json:"id"`
	DeviceID        string     `json:"device_id"`
	Kind            DeviceKind `json:"kind"`
	Target          string     `json:"target"`
	OperationID     string     `json:"operation_id"`
	OwnerGeneration uint64     `json:"owner_generation"`
	IssuedAt        time.Time  `json:"issued_at"`
}

func (c DeviceCommand) Validate() error {
	if c.ID == "" || c.DeviceID == "" || c.OperationID == "" {
		return fmt.Errorf("device command identity is incomplete")
	}
	if c.OwnerGeneration == 0 {
		return fmt.Errorf("device command %s has zero owner generation", c.ID)
	}
	switch c.Kind {
	case DeviceFan, DeviceDoor, DeviceValve, DeviceFeeder, DeviceGasSensor:
	default:
		return fmt.Errorf("device command %s has unknown kind %q", c.ID, c.Kind)
	}
	if c.Target == "" {
		return fmt.Errorf("device command %s has an empty target", c.ID)
	}
	return nil
}

type DeviceStatus struct {
	DeviceID        string     `json:"device_id"`
	Kind            DeviceKind `json:"kind"`
	Actual          string     `json:"actual"`
	OwnerGeneration uint64     `json:"owner_generation"`
	Healthy         bool       `json:"healthy"`
	ObservedAt      time.Time  `json:"observed_at"`
}

func (s DeviceStatus) Accepts(command DeviceCommand) bool {
	return s.DeviceID == command.DeviceID && command.OwnerGeneration >= s.OwnerGeneration
}
