package app

import "time"

const (
	defaultGrpcPort   = "57400"
	msgSize           = 512 * 1024 * 1024
	defaultRetryTimer = 10 * time.Second
)

var tlsVersions = []string{"1.3", "1.2", "1.1", "1.0", "1"}

type TargetError struct {
	TargetName string
	Err        error
}
