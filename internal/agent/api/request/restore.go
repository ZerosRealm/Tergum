package request

import "zerosrealm.xyz/tergum/internal/entity"

type Restore struct {
	Job      string       `json:"job"`
	Repo     *entity.Repo `json:"repo"`
	Snapshot string       `json:"snapshot"`
	Target   string       `json:"target"`
	Include  []string     `json:"include"`
	Exclude  []string     `json:"exclude"`
}
