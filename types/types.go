package types

type RequestBody struct {
	Terms []string `json:"terms"`
	Types []string `json:"types"`
}

type LocalRateLimits struct {
	Snusbase string
	NoSINT   string
}
