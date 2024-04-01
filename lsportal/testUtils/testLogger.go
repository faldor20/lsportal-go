package testUtils

import (
	"io"
	"strings"
	"testing"

	"github.com/tliron/commonlog"
	. "github.com/tliron/commonlog"
)

type TestBackend struct {
	Verbosity int
	Path      *string
	T         *testing.T
}

func (b *TestBackend) Configure(verbosity int, path *string) {
	b.Verbosity = verbosity
	b.Path = path
}

func (b *TestBackend) GetWriter() io.Writer {
	return nil
}

func (b *TestBackend) NewMessage(level Level, depth int, name ...string) Message {
	return &testMessage{
		level: level,
		depth: depth,
		name:  strings.Join(name, "."),
		kv:    map[string]any{},
		t:     b.T,
	}
}

func (b *TestBackend) AllowLevel(level Level, name ...string) bool {
	maxLevel := b.GetMaxLevel(name...)
	return level <= maxLevel
}

func (b *TestBackend) SetMaxLevel(level Level, name ...string) {
	// Not implemented for testing.T logger
}

func (b *TestBackend) GetMaxLevel(name ...string) Level {
	switch b.Verbosity {
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
