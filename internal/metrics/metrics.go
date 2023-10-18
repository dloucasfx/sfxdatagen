package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func Start(cfg *Config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger.Info("starting the metrics generator with configuration", zap.Any("config", cfg))

	dpWriter := New(cfg, logger)

	if err = Run(cfg, logger, dpWriter); err != nil {
		logger.Error("failed running the load generator", zap.Error(err))
		return err
	}
	return nil
}

func Run(c *Config, logger *zap.Logger, dpWriter *Writer) error {
	if c.TotalDuration > 0 {
		c.NumMetrics = 0
	} else if c.NumMetrics <= 0 {
		return fmt.Errorf("either `metrics` or `duration` must be greater than 0")
	}

	limit := rate.Limit(c.Rate)
	if c.Rate == 0 {
		limit = rate.Inf
		logger.Info("generation of metrics isn't being throttled")
	} else {
		logger.Info("generation of metrics is limited", zap.Float64("per-second", float64(limit)))
	}

	wg := sync.WaitGroup{}

	running := &atomic.Bool{}
	running.Store(true)
	dpWriter.Start()

	for i := 0; i < c.WorkerCount; i++ {
		wg.Add(1)
		w := worker{
			batchSize:      c.BatchSizeToWriter,
			dpChan:         dpWriter.dpChan,
			numMetrics:     c.NumMetrics,
			limitPerSecond: limit,
			totalDuration:  c.TotalDuration,
			running:        running,
			wg:             &wg,
			logger:         logger.With(zap.Int("worker", i)),
			index:          i,
		}

		go w.simulateMetrics()
	}
	if c.TotalDuration > 0 {
		time.Sleep(c.TotalDuration)
		running.Store(false)
	}
	wg.Wait()
	for {
		if dpWriter.datapointWriter.TotalInFlight == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(5 * time.Second)
	dpWriter.cancel()
	return nil
}
