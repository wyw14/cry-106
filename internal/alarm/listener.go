package alarm

import "sync"

type ZoneSnapshot struct {
	ZoneID     string `json:"zone_id"`
	Phase      string `json:"phase"`
	Generation uint64 `json:"generation"`
}

type ZoneReader interface {
	ReadZone(string) (ZoneSnapshot, bool)
}

type Published struct {
	Alarm Alarm        `json:"alarm"`
	Zone  ZoneSnapshot `json:"zone"`
}

type Listener struct {
	mu        sync.Mutex
	reader    ZoneReader
	published []Published
}

func NewListener(reader ZoneReader) *Listener {
	return &Listener{reader: reader}
}

func (l *Listener) SetReader(reader ZoneReader) {
	l.mu.Lock()
	l.reader = reader
	l.mu.Unlock()
}

func (l *Listener) Publish(value Alarm) {
	l.mu.Lock()
	reader := l.reader
	l.mu.Unlock()
	var zone ZoneSnapshot
	if reader != nil {
		zone, _ = reader.ReadZone(value.ZoneID)
	}
	l.mu.Lock()
	l.published = append(l.published, Published{Alarm: value, Zone: zone})
	l.mu.Unlock()
}

func (l *Listener) Events() []Published {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]Published, len(l.published))
	copy(result, l.published)
	return result
}
