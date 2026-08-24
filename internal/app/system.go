package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-106/internal/alarm"
	"github.com/wyw14/cry-106/internal/door"
	"github.com/wyw14/cry-106/internal/evacuation"
	"github.com/wyw14/cry-106/internal/extraction"
	"github.com/wyw14/cry-106/internal/interlock"
	"github.com/wyw14/cry-106/internal/journal"
	"github.com/wyw14/cry-106/internal/model"
	"github.com/wyw14/cry-106/internal/monitor"
	"github.com/wyw14/cry-106/internal/operation"
	"github.com/wyw14/cry-106/internal/power"
	"github.com/wyw14/cry-106/internal/sensor"
	"github.com/wyw14/cry-106/internal/topology"
	"github.com/wyw14/cry-106/internal/ventilation"
)

type System struct {
	mu                sync.Mutex
	events            *journal.Store
	cursor            *journal.Cursor
	topologyStore     *topology.SnapshotStore
	topology          *topology.Repository
	doorGraph         *door.Graph
	doors             *door.Controller
	access            *door.AccessController
	owners            *operation.Registry
	commands          *operation.Commander
	fanArbiter        *interlock.FanArbiter
	gasGuard          *interlock.GasGuard
	powerBarrier      *interlock.PowerBarrier
	zoneRelease       *interlock.ZoneRelease
	ventilation       *ventilation.Service
	fanStops          *ventilation.StopController
	planner           *ventilation.Planner
	calibrations      *sensor.Calibrations
	gasEvaluator      *alarm.Evaluator
	monitor           *monitor.SnapshotReader
	alarmListener     *alarm.Listener
	zoneState         *sensor.ZoneState
	samples           *journal.SampleStore
	receiver          *sensor.Receiver
	alarms            *alarm.Store
	extraction        *extraction.Service
	handover          *extraction.Handover
	cleanup           *extraction.CleanupRunner
	power             *power.Service
	isolator          *power.Isolator
	evacuations       *evacuation.Coordinator
	evacuationStore   *evacuation.Store
	clearances        *evacuation.ClearanceService
	clearanceDispatch *sensor.ClearanceDispatcher
	health            *monitor.HealthMonitor
	activity          *monitor.ActivityLog
	closed            bool
}

func Open(dataDir string) (*System, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory is required")
	}
	events, err := journal.Open(filepath.Join(dataDir, "events.jsonl"))
	if err != nil {
		return nil, err
	}
	cursor, err := journal.OpenCursor(filepath.Join(dataDir, "gas-cursor.json"))
	if err != nil {
		events.Close()
		return nil, err
	}
	topologyStore := topology.NewSnapshotStore(dataDir)
	snapshot, err := topologyStore.Load(topology.DefaultSnapshot())
	if err != nil {
		events.Close()
		return nil, err
	}
	topologyRepository, err := topology.NewRepository(snapshot)
	if err != nil {
		events.Close()
		return nil, err
	}
	zones := defaultZones()
	owners := operation.NewRegistry()
	commands := operation.NewCommander()
	fanArbiter := interlock.NewFanArbiter()
	gasGuard := interlock.NewGasGuard(1.5)
	powerBarrier := interlock.NewPowerBarrier()
	zoneRelease := interlock.NewZoneRelease()
	doorGraph := door.NewGraph()
	doorGraph.Apply(snapshot)
	topologyRepository.Subscribe(doorGraph.Apply)
	doors := door.NewController(map[string][]string{"east": {"D1", "D2"}, "north": {"D2", "D3"}})
	access := door.NewAccessController()
	gasEvaluator := alarm.NewEvaluator(gasGuard)
	alarmListener := alarm.NewListener(nil)
	zoneState := sensor.NewZoneState(zones, gasEvaluator, alarmListener)
	alarmListener.SetReader(zoneState)
	monitorReader := monitor.NewSnapshotReader(zones)
	samples := journal.NewSampleStore(events)
	receiver := sensor.NewReceiver(cursor, samples)
	powerService := power.NewService(events, owners, commands)
	evacuations := evacuation.NewCoordinator()
	system := &System{
		events: events, cursor: cursor, topologyStore: topologyStore,
		topology: topologyRepository, doorGraph: doorGraph, doors: doors, access: access,
		owners: owners, commands: commands, fanArbiter: fanArbiter, gasGuard: gasGuard,
		powerBarrier: powerBarrier, zoneRelease: zoneRelease,
		ventilation: ventilation.NewService(fanArbiter, owners, commands),
		fanStops:    ventilation.NewStopController(), planner: ventilation.NewPlanner(),
		calibrations: sensor.NewCalibrations(), gasEvaluator: gasEvaluator, monitor: monitorReader,
		alarmListener: alarmListener, zoneState: zoneState, samples: samples, receiver: receiver,
		alarms: alarm.NewStore(events), extraction: extraction.NewService(events, owners, commands),
		handover: extraction.NewHandover(owners), cleanup: extraction.NewCleanupRunner(owners, commands),
		power: powerService, isolator: power.NewIsolator(powerService, powerBarrier),
		evacuations: evacuations, evacuationStore: evacuation.NewStore(events),
		clearances:        evacuation.NewClearanceService(evacuations, zoneRelease, access),
		clearanceDispatch: sensor.NewClearanceDispatcher(zoneRelease), health: monitor.NewHealthMonitor(),
		activity: monitor.NewActivityLog(200),
	}
	if _, err := system.calibrations.Publish("east-gas-1", 0); err != nil {
		system.Close()
		return nil, err
	}
	if _, err := system.calibrations.Publish("north-gas-1", 0); err != nil {
		system.Close()
		return nil, err
	}
	if _, err := system.planner.Build(snapshot); err != nil {
		system.Close()
		return nil, err
	}
	if err := system.recover(); err != nil {
		system.Close()
		return nil, err
	}
	return system, nil
}

func defaultZones() []model.ZoneState {
	now := time.Now().UTC()
	return []model.ZoneState{
		{ID: "east", Name: "East wing", Phase: model.ZoneNormal, Generation: 1, UpdatedAt: now},
		{ID: "north", Name: "North wing", Phase: model.ZoneNormal, Generation: 1, UpdatedAt: now},
	}
}

func (s *System) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.events.Close()
}

func (s *System) Health() monitor.Health {
	return s.health.Snapshot()
}

func (s *System) RecoverNow() error {
	return s.recover()
}

func (s *System) recover() error {
	events, err := journal.ReadAll(s.events.Path())
	if err != nil {
		return err
	}
	if _, err := journal.Replay(events); err != nil {
		return err
	}
	if _, err := extraction.Recover(events); err != nil {
		return err
	}
	if _, err := evacuation.Recover(events); err != nil {
		return err
	}
	if _, err := power.Recover(events, s.powerBarrier); err != nil {
		return err
	}
	if _, err := journal.LoadSamples(s.events.Path(), ""); err != nil {
		s.health.Set("samples", "degraded")
		return err
	}
	s.health.Set("samples", "ok")
	return nil
}

func (s *System) StartFanOperation(ctx context.Context, request ventilation.TransferRequest) (string, error) {
	operationID, err := ventilation.Execute(ctx, s.ventilation, request)
	if err == nil {
		s.recordActivity("ventilation", "fan operation accepted", "", map[string]any{"group_id": request.GroupID, "mode": request.Mode, "operation_id": operationID})
	}
	return operationID, err
}

func (s *System) FanTargets() map[string]string {
	return s.ventilation.Targets()
}

func (s *System) ActiveFanOperations() []operation.Owner {
	return s.ventilation.ActiveOperations()
}

func (s *System) CompleteFanOperation(operationID string) error {
	if err := s.ventilation.CompleteOperation(operationID); err != nil {
		return err
	}
	s.recordActivity("ventilation", "fan operation completed", "", map[string]any{"operation_id": operationID})
	return nil
}

func (s *System) Commands() []model.DeviceCommand {
	commands := s.commands.Commands()
	sort.Slice(commands, func(i, j int) bool { return commands[i].IssuedAt.Before(commands[j].IssuedAt) })
	return commands
}

func (s *System) DeviceStatuses() []model.DeviceStatus {
	statuses := s.commands.Statuses()
	for _, status := range statuses {
		s.monitor.UpdateDevice(status)
	}
	return s.monitor.Devices()
}

func (s *System) SubmitGas(ctx context.Context, sample model.GasSample) (alarm.Alarm, error) {
	if sample.ID == "" {
		sample.ID = uuid.NewString()
	}
	if sample.ReceivedAt.IsZero() {
		sample.ReceivedAt = time.Now().UTC()
	}
	if sample.ObservedAt.IsZero() {
		sample.ObservedAt = sample.ReceivedAt
	}
	calibration, ok := s.calibrations.Current(sample.SensorID)
	if !ok {
		return alarm.Alarm{}, fmt.Errorf("sensor %s has no calibration", sample.SensorID)
	}
	if sample.CalibrationRevision == 0 {
		sample.CalibrationRevision = calibration.Revision
	}
	if err := s.receiver.Accept(ctx, sample); err != nil {
		return alarm.Alarm{}, err
	}
	value, changed, err := s.zoneState.ApplySample(sample)
	if err != nil {
		return alarm.Alarm{}, err
	}
	if changed {
		if _, err := s.alarms.Record(ctx, value); err != nil {
			return alarm.Alarm{}, err
		}
	}
	for _, zone := range s.zoneState.Zones() {
		s.monitor.UpdateZone(zone)
	}
	if changed && value.Severity == alarm.SeverityCritical {
		if _, err := s.isolator.Trip(ctx, sample.ZoneID); err != nil {
			return alarm.Alarm{}, err
		}
	}
	s.recordActivity("gas", "gas sample accepted", sample.ZoneID, map[string]any{"sensor_id": sample.SensorID, "sequence": sample.Sequence, "methane_percent": sample.MethanePercent})
	return value, nil
}

func (s *System) PublishCalibration(sensorID string, offset float64) (sensor.Calibration, error) {
	return s.calibrations.Publish(sensorID, offset)
}

func (s *System) GasZones() []model.ZoneState {
	for _, zone := range s.zoneState.Zones() {
		s.monitor.UpdateZone(zone)
	}
	return s.monitor.Zones()
}

func (s *System) GasEpochs() map[string]alarm.GasEpoch {
	result := make(map[string]alarm.GasEpoch)
	for _, zone := range s.GasZones() {
		if epoch, ok := s.gasEvaluator.Epoch(zone.ID); ok {
			result[zone.ID] = epoch
		}
	}
	return result
}

func (s *System) GasEpochHistory() map[string][]alarm.GasEpoch {
	result := make(map[string][]alarm.GasEpoch)
	for _, zone := range s.GasZones() {
		result[zone.ID] = s.gasEvaluator.EpochHistory(zone.ID)
	}
	return result
}

func (s *System) AlarmEvents() []alarm.Published {
	return s.alarmListener.Events()
}

func (s *System) Alarms() []alarm.Alarm {
	return s.gasEvaluator.List()
}

func (s *System) StartExtraction(ctx context.Context, groupID, mode string) (extraction.Session, error) {
	session, err := s.extraction.Start(ctx, groupID, mode)
	if err == nil {
		s.recordActivity("extraction", "extraction session started", groupID, map[string]any{"session_id": session.ID, "mode": mode})
	}
	return session, err
}

func (s *System) ExtractionSessions() []extraction.Session {
	return s.extraction.Sessions()
}

func (s *System) TransferExtraction(currentID, successorID, mode string) (extraction.Session, error) {
	current, ok := s.extraction.Session(currentID)
	if !ok {
		return extraction.Session{}, fmt.Errorf("unknown extraction session %s", currentID)
	}
	return s.handover.Commit(current, successorID, mode)
}

func (s *System) CleanupExtraction(retired extraction.Session) (bool, error) {
	return s.cleanup.CloseGroup(retired)
}

func (s *System) BeginEvacuation(ctx context.Context, zoneID string, expected int) (evacuation.Round, error) {
	round, err := s.clearances.Begin(zoneID, expected)
	if err != nil {
		return evacuation.Round{}, err
	}
	if _, err := s.evacuationStore.Record(ctx, round); err != nil {
		return evacuation.Round{}, err
	}
	s.recordActivity("evacuation", "evacuation round started", zoneID, map[string]any{"round_id": round.ID, "expected": expected})
	return round, nil
}

func (s *System) AccountEvacuation(ctx context.Context, zoneID, roundID string, count int) (evacuation.Round, error) {
	round, err := s.evacuations.Account(zoneID, roundID, count)
	if err != nil {
		return evacuation.Round{}, err
	}
	if _, err := s.evacuationStore.Record(ctx, round); err != nil {
		return evacuation.Round{}, err
	}
	return round, nil
}

func (s *System) ApplyClearance(ctx context.Context, clearance model.AirClearance) (evacuation.Round, error) {
	if err := s.clearanceDispatch.Deliver(clearance); err != nil && !errors.Is(err, context.Canceled) {
		return evacuation.Round{}, err
	}
	round, err := s.clearances.Apply(clearance)
	if err != nil {
		return evacuation.Round{}, err
	}
	if _, err := s.evacuationStore.Record(ctx, round); err != nil {
		return evacuation.Round{}, err
	}
	return round, nil
}

func (s *System) Evacuations() []evacuation.Round {
	return s.evacuations.Rounds()
}

func (s *System) AccessAllowed(zoneID string) bool {
	return s.access.Allowed(zoneID)
}

func (s *System) Isolate(ctx context.Context, zoneID, fanID, doorGroup string, timeout time.Duration) (operation.IsolationResult, error) {
	return operation.Isolate(ctx, zoneID, fanID, doorGroup, timeout, s.fanStops, s.doors)
}

func (s *System) SetFanStopDelay(fanID string, delay time.Duration) {
	s.fanStops.SetDelay(fanID, delay)
}

func (s *System) DoorStates() []door.State {
	return s.doors.States()
}

func (s *System) Topology() model.TopologySnapshot {
	return s.topology.Current()
}

func (s *System) PublishTopology(next model.TopologySnapshot) (ventilation.Plan, error) {
	if err := s.topologyStore.Save(next); err != nil {
		return ventilation.Plan{}, err
	}
	if err := s.topology.Publish(next); err != nil {
		return ventilation.Plan{}, err
	}
	return s.planner.Build(s.topology.Current())
}

func (s *System) TopologyPath(from, to string) ([]string, error) {
	return topology.FindPath(s.topology.Current(), from, to)
}

func (s *System) DoorGraph() (uint64, []model.AirwayEdge) {
	return s.doorGraph.Edges()
}
