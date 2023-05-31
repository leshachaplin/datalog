package domain

import "time"

type Event struct {
	ServerTime time.Time `json:"server_time"`
	IP         string    `json:"ip"`
	ClientTime string    `json:"client_time"`
	DeviceID   string    `json:"device_id"`
	DeviceOS   string    `json:"device_os"`
	Session    string    `json:"session"`
	Event      string    `json:"event"`
	ParamStr   string    `json:"param_str"`
	Sequence   int       `json:"sequence"`
	ParamInt   int       `json:"param_int"`
}

func (e *Event) EnrichWith(clientIP string, serverTime time.Time) {
	e.IP = clientIP
	e.ServerTime = serverTime
}

type EventBatch struct {
	ID     string  `json:"id"`
	Events []Event `json:"events"`
}
