package clickhouse

import (
	"time"

	"github.com/leshachaplin/datalog/internal/domain"
)

type eventBatch struct {
	Events []event `json:"events"`
}

type event struct {
	IP         string `ch:"ip"`
	ServerTime string `ch:"server_time"`
	ClientTime string `ch:"client_time"`
	DeviceID   string `ch:"device_id"`
	DeviceOS   string `ch:"device_os"`
	Session    string `ch:"session"`
	Sequence   int16  `ch:"sequence"`
	EventType  string `ch:"event_type"`
	ParamsInt  int32  `ch:"param_int"`
	ParamStr   string `ch:"param_str"`
}

func eventFromService(evnts []domain.Event) eventBatch {
	events := make([]event, len(evnts))
	for i := 0; i < len(events); i++ {
		events[i] = event{
			IP:         evnts[i].IP,
			ServerTime: evnts[i].ServerTime.Format(time.DateTime),
			ClientTime: evnts[i].ClientTime,
			DeviceID:   evnts[i].DeviceID,
			DeviceOS:   evnts[i].DeviceOS,
			Session:    evnts[i].Session,
			Sequence:   int16(evnts[i].Sequence),
			EventType:  evnts[i].Event,
			ParamsInt:  int32(evnts[i].ParamInt),
			ParamStr:   evnts[i].ParamStr,
		}
	}
	return eventBatch{
		Events: events,
	}
}
