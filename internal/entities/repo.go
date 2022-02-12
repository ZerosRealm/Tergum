package entities

// Repo for storing repository information.
type Repo struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Repo     string   `json:"repo"`
	Password string   `json:"password"`
	Settings []string `json:"settings"`
}
