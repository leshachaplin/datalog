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

func eventFromService(batch domain.EventBatch) eventBatch {
	events := make([]event, len(batch.Events))
	for i := 0; i < len(batch.Events); i++ {
		events[i] = event{
			IP:         batch.Events[i].IP,
			ServerTime: batch.Events[i].ServerTime.Format(time.DateTime),
			ClientTime: batch.Events[i].ClientTime,
			DeviceID:   batch.Events[i].DeviceID,
			DeviceOS:   batch.Events[i].DeviceOS,
			Session:    batch.Events[i].Session,
			Sequence:   int16(batch.Events[i].Sequence),
			EventType:  batch.Events[i].Event,
			ParamsInt:  int32(batch.Events[i].ParamInt),
			ParamStr:   batch.Events[i].ParamStr,
		}
	}
	return eventBatch{
		Events: events,
	}
}
