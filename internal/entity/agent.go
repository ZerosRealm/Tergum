package entity

// Agent to send jobs to.
type Agent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	PSK  string `json:"psk"`
}
