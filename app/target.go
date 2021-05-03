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

func (a *App) GetTargets() (map[string]*Target, error) {
	targetsConfigs, err := a.Config.GetTargets()
	if err != nil {
		return nil, err
	}
	targets := make(map[string]*Target)
	for n, tc := range targetsConfigs {
		targets[n] = NewTarget(tc)
	}
	return targets, nil
}
