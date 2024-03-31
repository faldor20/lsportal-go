package lsportal

import (
	"io"
	"strings"
	"testing"

	"github.com/tliron/commonlog"
	. "github.com/tliron/commonlog"
)

type testBackend struct {
	verbosity int
	path      *string
	t         *testing.T
}

func (b *testBackend) Configure(verbosity int, path *string) {
	b.verbosity = verbosity
	b.path = path
}

func (b *testBackend) GetWriter() io.Writer {
	return nil
}

func (b *testBackend) NewMessage(level Level, depth int, name ...string) Message {
	return &testMessage{
		level: level,
		depth: depth,
		name:  strings.Join(name, "."),
		kv:    map[string]any{},
		t:     b.t,
	}
}

func (b *testBackend) AllowLevel(level Level, name ...string) bool {
	maxLevel := b.GetMaxLevel(name...)
	return level <= maxLevel
}

func (b *testBackend) SetMaxLevel(level Level, name ...string) {
	// Not implemented for testing.T logger
}

func (b *testBackend) GetMaxLevel(name ...string) Level {
	switch b.verbosity {
	case -4:
		return None
	case -3:
		return Critical
	case -2:
		return Error
	case -1:
		return Warning
	case 0:
		return Notice
	case 1:
		return Info
	default:
		return Debug
	}
}

type testMessage struct {
	level Level
	depth int
	name  string
	kv    map[string]any
	t     *testing.T
}

// Send implements commonlog.Message.
func (m *testMessage) Send() {
	m.t.Logf("%s %s", m.name, m.kv)
}

// Set implements commonlog.Message.
func (m *testMessage) Set(key string, value any) Message {
	m.kv[key] = value
	return m
}

func (m *testMessage) Log(args ...interface{}) {
	if m.level <= commonlog.Error {
		m.t.Error(args...)
	} else {
		m.t.Log(args...)
	}
}

func (m *testMessage) Logf(format string, args ...interface{}) {
	if m.level <= Error {
		m.t.Errorf(format, args...)
	} else {
		m.t.Logf(format, args...)
	}
}
