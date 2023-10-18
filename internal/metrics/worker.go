// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type worker struct {
	batchSize      int64
	dpChan         chan []*datapoint.Datapoint
	running        *atomic.Bool    // pointer to shared flag that indicates it's time to stop the test
	numMetrics     int             // how many metrics the worker has to generate (only when duration==0)
	totalDuration  time.Duration   // how long to run the test for (overrides `numMetrics`)
	limitPerSecond rate.Limit      // how many metrics per second to generate
	wg             *sync.WaitGroup // notify when done
	logger         *zap.Logger     // logger
	index          int             // worker index
}

func (w worker) simulateMetrics() {
	limiter := rate.NewLimiter(w.limitPerSecond, 1)

	var i int64
	var metrics []*datapoint.Datapoint
	metricName := fmt.Sprintf("gensfxdp%d", w.index)
	for w.running.Load() {
		metrics = append(metrics, &datapoint.Datapoint{
			Metric: metricName,
			Dimensions: map[string]string{
				"source":   "dim1Val",
				"plugin":   "name",
				"testdim1": "dim1",
				"testdim2": "dim2",
				"testdim3": "dim3",
				"testdim4": "dim4",
				"testdim5": "dim5",
				"testdim6": "dim6",
				"testdim7": "dim7",
				"testdim8": "dim8",
				"testdim9": "dim9",
			},
			Value:      datapoint.NewIntValue(i),
			MetricType: datapoint.Counter,
			Timestamp:  time.Now(),
		})

		if err := limiter.Wait(context.Background()); err != nil {
			w.logger.Fatal("limiter wait failed, retry", zap.Error(err))
		}

		i++
		if i%w.batchSize == 0 {
			w.logger.Info("adding to buffer", zap.Int("slice len", len(metrics)))
			w.dpChan <- metrics
			metrics = nil
		}
		if w.numMetrics != 0 && i >= int64(w.numMetrics) {
			w.dpChan <- metrics
			metrics = nil
			break
		}
	}
	if len(metrics) > 0 {
		w.dpChan <- metrics
	}
	w.logger.Info("metrics generated", zap.Int64("metrics", i))
	w.wg.Done()
}
