package worker

type Config struct {
	NumWorkers       int `mapstructure:"num_workers"`
	BatchSize        int `mapstructure:"batch_size"`
	MaxBatchCapacity int `mapstructure:"max_batch_capacity"`
}
