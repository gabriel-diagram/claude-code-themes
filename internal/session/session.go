// Package session holds the per-session scratch state, in $TMPDIR.
//
// What only becomes knowable by comparing one statusline refresh with the
// previous one - token rate, context peak, bypass turns, session length - lives
// here. The SessionEnd hook turns it into behaviour counters and deletes it.
//
// All three files share the Prefix so a single sweep reaches them all.
package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Prefix marks every file this package owns.
const Prefix = "claude-statusline-"

// MaxAge is how long a leftover survives before the sweep takes it.
const MaxAge = 24 * time.Hour

// Facts is what the statusline carries from one refresh to the next.
type Facts struct {
	Label string  `json:"label"`
	Peak  float64 `json:"peak"`

	// CtxPeak is the context's OWN peak, kept apart from Peak now that Peak is
	// the tightest of the three necks. Half the counters are named for context
	// and mean it - ctx_low, sessions_under_40, ctx100_sessions - and reading
	// them off the neck would credit "3 sessions touching 100% of context" to
	// somebody who never filled the window and only ran out of 5h quota.
	CtxPeak float64 `json:"ctx_peak"`

	T0       int64    `json:"t0"`
	Repo     string   `json:"repo"`
	PromptID string   `json:"prompt_id"`
	APIMs    *float64 `json:"api_ms"`
	TPS      *float64 `json:"tps"`
	TPSAt    float64  `json:"tps_at"`

	// Structured says the file actually carried session facts. A file with
	// only a bare label is one this version knows nothing about, and closing a
	// session on it would hand out counters for free.
	Structured bool `json:"-"`
}

// SafeID whitelists a session id, because it becomes part of a path.
func SafeID(id string) string {
	if id == "" || len(id) > 64 {
		return ""
	}
	for _, r := range id {
		ok := r == '_' || r == '-' ||
			(r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !ok {
			return ""
		}
	}
	return id
}

// TmpDir is where the scratch files live.
func TmpDir() string {
	if d := os.Getenv("TMPDIR"); d != "" {
		return d
	}
	return "/tmp"
}

// PathFor is the scratch file for a session. kind "" is the statusline's own.
func PathFor(id, kind string) string {
	safe := SafeID(id)
	if safe == "" {
		return ""
	}
	if kind != "" {
		kind += "-"
	}
	return filepath.Join(TmpDir(), Prefix+kind+safe)
}

// Load never fails. A file that is not json is read as a bare label, which is
// what the very first version of this wrote.
func Load(path string) Facts {
	var f Facts
	if path == "" {
		return f
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return f
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return f
	}
	if !strings.HasPrefix(text, "{") {
		f.Label = text
		return f
	}
	var doc map[string]any
	if json.Unmarshal([]byte(text), &doc) != nil {
		return Facts{}
	}
	// v1 keys, kept readable so a session that spans an upgrade is not lost.
	if v, ok := doc["etq"]; ok {
		doc["label"] = v
	}
	if v, ok := doc["pico"]; ok {
		doc["peak"] = v
	}
	if v, ok := doc["api"]; ok {
		doc["api_ms"] = v
	}
	if v, ok := doc["pid"]; ok {
		doc["prompt_id"] = v
	}
	if v, ok := doc["tps_t"]; ok {
		doc["tps_at"] = v
	}
	repacked, err := json.Marshal(doc)
	if err != nil {
		return Facts{}
	}
	if json.Unmarshal(repacked, &f) != nil {
		return Facts{}
	}
	_, hasPeak := doc["peak"]
	_, hasT0 := doc["t0"]
	f.Structured = hasPeak || hasT0
	// A file written before the two peaks were told apart has only the one.
	// Falling back keeps a session that spans an upgrade from reading as a
	// context that never rose above zero, which would hand out ctx_low and
	// sessions_under_40 for free.
	if _, has := doc["ctx_peak"]; !has {
		f.CtxPeak = f.Peak
	}
	return f
}

// Save writes the facts. A failure is not worth reporting: the next refresh
// will try again.
func Save(path string, f Facts) bool {
	if path == "" {
		return false
	}
	raw, err := json.Marshal(f)
	if err != nil {
		return false
	}
	return os.WriteFile(path, raw, 0o600) == nil
}

// Sweep drops the leftovers of sessions that died without a SessionEnd.
func Sweep(now time.Time) {
	dir := TmpDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := now.Add(-MaxAge)
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), Prefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		os.Remove(filepath.Join(dir, e.Name()))
	}
}
