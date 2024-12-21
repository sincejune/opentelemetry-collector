// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package scraperhelper

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/testdata"
	"go.opentelemetry.io/collector/scraper"
)

func TestScrapeLogsDataOp(t *testing.T) {
	tt, err := componenttest.SetupTelemetry(receiverID)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, tt.Shutdown(context.Background())) })

	parentCtx, parentSpan := tt.TelemetrySettings().TracerProvider.Tracer("test").Start(context.Background(), t.Name())
	defer parentSpan.End()

	params := []testParams{
		{items: 23, err: partialErrFake},
		{items: 29, err: errFake},
		{items: 15, err: nil},
	}
	for i := range params {
		var sf scraper.ScrapeLogsFunc
		sf, err = newObsLogs(func(context.Context) (plog.Logs, error) {
			return testdata.GenerateLogs(params[i].items), params[i].err
		}, receiverID, scraperID, tt.TelemetrySettings())
		require.NoError(t, err)
		_, err = sf.ScrapeLogs(parentCtx)
		require.ErrorIs(t, err, params[i].err)
	}

	spans := tt.SpanRecorder.Ended()
	require.Equal(t, len(params), len(spans))

	var scrapedLogRecords, erroredLogRecords int
	for i, span := range spans {
		assert.Equal(t, "scraper/"+scraperID.String()+"/ScrapeLogs", span.Name())
		switch {
		case params[i].err == nil:
			scrapedLogRecords += params[i].items
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: scrapedLogRecordsKey, Value: attribute.Int64Value(int64(params[i].items))})
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: erroredLogRecordsKey, Value: attribute.Int64Value(0)})
			assert.Equal(t, codes.Unset, span.Status().Code)
		case errors.Is(params[i].err, errFake):
			// Since we get an error, we cannot record any metrics because we don't know if the returned plog.Logs is valid instance.
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: scrapedLogRecordsKey, Value: attribute.Int64Value(0)})
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: erroredLogRecordsKey, Value: attribute.Int64Value(0)})
			assert.Equal(t, codes.Error, span.Status().Code)
			assert.Equal(t, params[i].err.Error(), span.Status().Description)
		case errors.Is(params[i].err, partialErrFake):
			scrapedLogRecords += params[i].items
			erroredLogRecords += 2
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: scrapedLogRecordsKey, Value: attribute.Int64Value(int64(params[i].items))})
			require.Contains(t, span.Attributes(), attribute.KeyValue{Key: erroredLogRecordsKey, Value: attribute.Int64Value(2)})
			assert.Equal(t, codes.Error, span.Status().Code)
			assert.Equal(t, params[i].err.Error(), span.Status().Description)
		default:
			t.Fatalf("unexpected err param: %v", params[i].err)
		}
	}

	require.NoError(t, tt.CheckScraperLogs(receiverID, scraperID, int64(scrapedLogRecords), int64(erroredLogRecords)))
}

func TestCheckScraperLogs(t *testing.T) {
	tt, err := componenttest.SetupTelemetry(receiverID)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, tt.Shutdown(context.Background())) })

	var sf scraper.ScrapeLogsFunc
	sf, err = newObsLogs(func(context.Context) (plog.Logs, error) {
		return testdata.GenerateLogs(7), nil
	}, receiverID, scraperID, tt.TelemetrySettings())
	require.NoError(t, err)
	_, err = sf.ScrapeLogs(context.Background())
	assert.NoError(t, err)

	require.NoError(t, tt.CheckScraperLogs(receiverID, scraperID, 7, 0))
	require.Error(t, tt.CheckScraperLogs(receiverID, scraperID, 7, 7))
	require.Error(t, tt.CheckScraperLogs(receiverID, scraperID, 0, 0))
	require.Error(t, tt.CheckScraperLogs(receiverID, scraperID, 0, 7))
}