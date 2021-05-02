package app

import (
	"github.com/karimra/gnoic/config"
	"google.golang.org/grpc"
)

type Target struct {
	Config *config.TargetConfig
	client *grpc.ClientConn
}

func NewTarget(tc *config.TargetConfig) *Target {
	return &Target{Config: tc}
}
