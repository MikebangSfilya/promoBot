package audit

import "time"

type Log struct {
	Code   string    `json:"code"`
	Action string    `json:"action"`
	By     string    `json:"by"`
	At     time.Time `json:"at"`
}
