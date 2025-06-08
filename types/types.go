package types

type RequestBody struct {
	Terms []string `json:"terms"`
	Types []string `json:"types"`
}

type LocalRateLimits struct {
	Snusbase string
	NoSINT   string
	Maigret  string
}

type Maigret struct {
	SiteName string   `json:"sitename"`
	UrlUser  string   `json:"urluser"`
	User     string   `json:"user"`
	Tags     []string `json:"tags"`
}

type MaigretList struct {
	MGList []Maigret `json:"data"`
}
