package clickhouse

import (
	"context"

	"github.com/leshachaplin/datalog/internal/domain"
)

func (c *Clickhouse) StoreEvents(ctx context.Context, events domain.EventBatch) error {
	eBatch := eventFromService(events)

	batch, err := c.conn.PrepareBatch(ctx, `INSERT INTO events`)
	if err != nil {
		return err
	}
	for i := 0; i < len(eBatch.Events); i++ {
		if errAppend := batch.AppendStruct(&eBatch.Events[i]); errAppend != nil {
			return errAppend
		}
	}
	return batch.Send()
}
