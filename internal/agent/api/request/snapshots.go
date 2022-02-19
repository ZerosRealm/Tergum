package request

import "zerosrealm.xyz/tergum/internal/entity"

type GetSnapshots struct {
	Repo *entity.Repo `json:"repo"`
}

type Forget struct {
	Repo   *entity.Repo   `json:"repo"`
	Policy *entity.Forget `json:"policy"`
}

type DeleteSnapshot struct {
	Repo      *entity.Repo `json:"repo"`
	Snapshots []string     `json:"snapshots"`
}

type List struct {
	Repo     *entity.Repo `json:"repo"`
	Snapshot string       `json:"snapshot"`
}
