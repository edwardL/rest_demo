package v1

type ApiItem struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Value  string `json:"value"`
	Label  string `json:"label"`
}
