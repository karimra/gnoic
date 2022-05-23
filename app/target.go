package app

import (
	"github.com/karimra/gnoic/api"
)

func (a *App) GetTargets() (map[string]*api.Target, error) {
	targetsConfigs, err := a.Config.GetTargets()
	if err != nil {
		return nil, err
	}
	targets := make(map[string]*api.Target)
	for n, tc := range targetsConfigs {
		targets[n] = api.NewTargetFromConfig(tc)
	}
	return targets, nil
}
