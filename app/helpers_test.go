package app

import (
	"io"
	"testing"

	log "github.com/sirupsen/logrus"
)

// newTestApp returns an App whose logger is silenced.
func newTestApp(t *testing.T) *App {
	t.Helper()
	a := New()
	a.Logger.Logger.SetOutput(io.Discard)
	a.Config.SetLogger()
	log.StandardLogger().SetOutput(io.Discard)
	return a
}

func ptr[T any](v T) *T { return &v }
