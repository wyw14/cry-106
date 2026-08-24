package model

import (
	"fmt"
	"time"
)

type GasSample struct {
	ID                  string    `json:"id"`
	SensorID            string    `json:"sensor_id"`
	ZoneID              string    `json:"zone_id"`
	Sequence            uint64    `json:"sequence"`
	CalibrationRevision uint64    `json:"calibration_revision"`
	MethanePercent      float64   `json:"methane_percent"`
	ObservedAt          time.Time `json:"observed_at"`
	ReceivedAt          time.Time `json:"received_at"`
}

func (s GasSample) Validate() error {
	if s.ID == "" || s.SensorID == "" || s.ZoneID == "" {
		return fmt.Errorf("gas sample identity is incomplete")
	}
	if s.Sequence == 0 || s.CalibrationRevision == 0 {
		return fmt.Errorf("gas sample %s has an invalid sequence or calibration", s.ID)
	}
	if s.MethanePercent < 0 || s.MethanePercent > 100 {
		return fmt.Errorf("gas sample %s has invalid methane percent %.3f", s.ID, s.MethanePercent)
	}
	if s.ObservedAt.IsZero() || s.ReceivedAt.IsZero() {
		return fmt.Errorf("gas sample %s is missing timestamps", s.ID)
	}
	return nil
}

type AirClearance struct {
	ID           string    `json:"id"`
	ZoneID       string    `json:"zone_id"`
	EvacuationID string    `json:"evacuation_id"`
	Approved     bool      `json:"approved"`
	MethanePPM   int       `json:"methane_ppm"`
	ObservedAt   time.Time `json:"observed_at"`
}

func (c AirClearance) Validate() error {
	if c.ID == "" || c.ZoneID == "" || c.EvacuationID == "" {
		return fmt.Errorf("air clearance identity is incomplete")
	}
	if c.MethanePPM < 0 || c.ObservedAt.IsZero() {
		return fmt.Errorf("air clearance %s contains invalid measurements", c.ID)
	}
	return nil
}
