package metrics

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	sfxwriter "github.com/signalfx/signalfx-go/writer"
	"go.uber.org/zap"
)

type Writer struct {
	client          *sfxclient.HTTPSink
	datapointWriter *sfxwriter.DatapointWriter
	dpChan          chan []*datapoint.Datapoint
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

func New(conf *Config, logger *zap.Logger) *Writer {
	logger.Info("setting up the writer")
	dpChan := make(chan []*datapoint.Datapoint, 5000)
	ctx, cancel := context.WithCancel(context.Background())

	sw := &Writer{
		dpChan: dpChan,
		client: sfxclient.NewHTTPSink(),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
	sw.client.AuthToken = conf.SignalFxAccessToken
	sw.client.Client.Timeout = conf.HTTPTimeout
	sw.client.Client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	sw.client.DatapointEndpoint = fmt.Sprintf("https://ingest.%s.signalfx.com/v2/datapoint", conf.Realm)
	sw.datapointWriter = &sfxwriter.DatapointWriter{
		SendFunc: sw.sendDatapoints,
		OverwriteFunc: func() {
			sw.logger.Info(fmt.Sprintf("A datapoint was overwritten in the write buffer, please consider increasing the writer.maxDatapointsBuffered config option to something greater than %d", 90000))
		},
		MaxBatchSize: 8192,
		MaxRequests:  10,
		MaxBuffered:  90000,
		InputChan:    sw.dpChan,
	}
	return sw
}

func (sw *Writer) Start() {
	sw.datapointWriter.Start(sw.ctx)
}

func (sw *Writer) sendDatapoints(ctx context.Context, dps []*datapoint.Datapoint) error {
	// This sends synchronously and retries on transient connection errors
	sw.logger.Info("Trying to send datapoints", zap.Int("length", len(dps)))
	err := sw.client.AddDatapoints(ctx, dps)
	if err != nil {
		if isTransientError(err) {
			sw.logger.Info("retrying datapoint submission after receiving temporary network error. ", zap.Error(err))
			err = sw.client.AddDatapoints(ctx, dps)
		}
		if err != nil {
			sw.logger.Info("Error shipping datapoints to SignalFx.", zap.Error(err))
			return err
		}

	}
	sw.logger.Info("Number of datapoints sent", zap.Int("length", len(dps)))
	return nil
}

func isTransientError(candidate error) bool {
	var isTemporary, isReset, isEOF bool
	for candidate != nil {
		if temp, ok := candidate.(interface{ Temporary() bool }); ok {
			if temp.Temporary() {
				isTemporary = true
				break
			}
		}

		if se, ok := candidate.(syscall.Errno); ok {
			if se == syscall.ECONNRESET {
				isReset = true
				break
			}
		}

		if candidate == io.EOF {
			isEOF = true
			break
		}
		candidate = errors.Unwrap(candidate)
	}

	return isTemporary || isReset || isEOF
}
