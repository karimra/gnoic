package config

import log "github.com/sirupsen/logrus"

func logDebugLevel() log.Level { return log.DebugLevel }

func ptr[T any](v T) *T { return &v }
