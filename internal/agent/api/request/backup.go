package request

import "zerosrealm.xyz/tergum/internal/entity"

type Backup struct {
	Job
	Repo   *entity.Repo   `json:"repo"`
	Backup *entity.Backup `json:"backup"`
}
