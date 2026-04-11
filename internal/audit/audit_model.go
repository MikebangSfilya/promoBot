package audit

import "time"

type Log struct {
	Code   string    `json:"code"`
	Action string    `json:"action"`
	By     string    `json:"created_by"`
	At     time.Time `json:"created_at"`
}
