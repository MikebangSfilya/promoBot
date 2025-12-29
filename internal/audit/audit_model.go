package audit

type Log struct {
	Code   string `json:"code"`
	Action string `json:"action"`
	By     string `json:"by"`
}
