package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Sample struct {
	Operation  string    `json:"operation"`
	DurationMs int64     `json:"duration_ms"`
	Success    bool      `json:"success"`
	At         time.Time `json:"at"`
}

type OperationStats struct {
	Operation string    `json:"operation"`
	Count     int       `json:"count"`
	Success   int       `json:"success"`
	Failed    int       `json:"failed"`
	AvgMs     float64   `json:"avg_ms"`
	P50Ms     int64     `json:"p50_ms"`
	P95Ms     int64     `json:"p95_ms"`
	MaxMs     int64     `json:"max_ms"`
	LastAt    time.Time `json:"last_at"`
}

type MetricsSnapshot struct {
	GeneratedAt time.Time        `json:"generated_at"`
	SampleCount int              `json:"sample_count"`
	Operations  []OperationStats `json:"operations"`
	DBPing      *OperationStats  `json:"db_ping,omitempty"`
	Import      *OperationStats  `json:"import,omitempty"`
	Pipeline    *OperationStats  `json:"pipeline,omitempty"`
}

type Recorder struct {
	mu      sync.Mutex
	samples []Sample
	max     int
}

func NewRecorder(max int) *Recorder {
	if max <= 0 {
		max = 2000
	}
	return &Recorder{max: max}
}

var DefaultMetrics = NewRecorder(2000)

func Observe(operation string, duration time.Duration, success bool) {
	DefaultMetrics.Observe(operation, duration, success)
}

func (r *Recorder) Observe(operation string, duration time.Duration, success bool) {
	if r == nil || operation == "" {
		return
	}
	s := Sample{Operation: operation, DurationMs: duration.Milliseconds(), Success: success, At: time.Now().UTC()}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples = append(r.samples, s)
	if len(r.samples) > r.max {
		r.samples = r.samples[len(r.samples)-r.max:]
	}
}

func (r *Recorder) Snapshot() MetricsSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	grouped := map[string][]Sample{}
	for _, s := range r.samples {
		grouped[s.Operation] = append(grouped[s.Operation], s)
	}
	ops := make([]OperationStats, 0, len(grouped))
	for name, items := range grouped {
		ops = append(ops, statsFor(name, items))
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].Count > ops[j].Count })
	snap := MetricsSnapshot{GeneratedAt: time.Now().UTC(), SampleCount: len(r.samples), Operations: ops}
	for i := range ops {
		switch ops[i].Operation {
		case "db.ping":
			item := ops[i]
			snap.DBPing = &item
		case "import.process", "import.api":
			if snap.Import == nil || ops[i].Count > snap.Import.Count {
				item := ops[i]
				snap.Import = &item
			}
		case "processing.pipeline", "processing.pipeline.api":
			if snap.Pipeline == nil || ops[i].Count > snap.Pipeline.Count {
				item := ops[i]
				snap.Pipeline = &item
			}
		}
	}
	return snap
}

func statsFor(name string, items []Sample) OperationStats {
	durs := make([]int64, 0, len(items))
	var sum int64
	success, failed := 0, 0
	var last time.Time
	var max int64
	for _, s := range items {
		durs = append(durs, s.DurationMs)
		sum += s.DurationMs
		if s.Success {
			success++
		} else {
			failed++
		}
		if s.At.After(last) {
			last = s.At
		}
		if s.DurationMs > max {
			max = s.DurationMs
		}
	}
	sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })
	avg := 0.0
	if len(durs) > 0 {
		avg = float64(sum) / float64(len(durs))
	}
	return OperationStats{
		Operation: name, Count: len(items), Success: success, Failed: failed,
		AvgMs: avg, P50Ms: percentile(durs, 50), P95Ms: percentile(durs, 95), MaxMs: max, LastAt: last,
	}
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted) / 100)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(DefaultMetrics.Snapshot())
	}
}

func Track(ctx context.Context, operation string, fn func(context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	Observe(operation, time.Since(start), err == nil)
	return err
}
