package metrics

import (
	"github.com/spf13/pflag"
	"time"
)

type Config struct {
	HTTPTimeout         time.Duration
	SignalFxAccessToken string
	WorkerCount         int
	Rate                int64
	TotalDuration       time.Duration
	BatchSizeToWriter   int64
	Realm               string
	NumMetrics          int
	HeaderDebugID       string
}

func (c *Config) Flags(fs *pflag.FlagSet) {
	fs.IntVar(&c.WorkerCount, "workers", 1, "Number of workers (goroutines) to run")
	fs.Int64Var(&c.Rate, "rate", 0, "Approximately how many metrics per second each worker should generate. Zero means no throttling.")
	fs.DurationVar(&c.TotalDuration, "duration", 0, "For how long to run the test")
	fs.Int64Var(&c.BatchSizeToWriter, "batch-writer-size", 100, "How many datapoints batch to submit to the writer")
	fs.DurationVar(&c.HTTPTimeout, "http-timeout", 5*time.Second, "Writer client http timeout")
	fs.StringVar(&c.SignalFxAccessToken, "access-token", "", "SignalFX Ingest Access Token")
	fs.StringVar(&c.Realm, "realm", "us0", "Realm your Org belongs to")
	fs.IntVar(&c.NumMetrics, "metrics", 1, "Number of metrics to generate in each worker (ignored if duration is provided)")
	fs.StringVar(&c.HeaderDebugID, "header-debug-id", "", "ID value to set the header X-Debug-Id value")
}
