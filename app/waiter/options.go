package waiter

import (
	"os"
)

type Option func(*waiterCfg)

func WithSignals(signals ...os.Signal) Option {
	return func(cfg *waiterCfg) {
		cfg.signals = signals
	}
}
