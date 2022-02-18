package entity

type Forget struct {
	ID      int  `json:"id"`
	Enabled bool `json:"enabled"`
	LastX   int  `json:"lastX"`
	Hourly  int  `json:"hourly"`
	Daily   int  `json:"daily"`
	Weekly  int  `json:"weekly"`
	Monthly int  `json:"monthly"`
	Yearly  int  `json:"yearly"`
}
