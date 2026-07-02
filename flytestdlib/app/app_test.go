package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// TestServe_ScopeAndMetricsEndpoint verifies that serve() initializes a
// non-nil SetupContext.Scope before Setup runs, and that metrics registered
// through that scope are exposed on the /metrics endpoint.
func TestServe_ScopeAndMetricsEndpoint(t *testing.T) {
	const port = 18099

	var gotScope promutils.Scope
	a := &App{
		Name: "test-app-metrics",
		Setup: func(_ context.Context, sc *SetupContext) error {
			sc.Port = port
			gotScope = sc.Scope
			require.NotNil(t, sc.Scope)
			counter := sc.Scope.MustNewCounter("test_counter", "test counter for /metrics endpoint")
			counter.Inc()
			return nil
		},
	}

	go func() {
		_ = a.serve(context.Background())
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	var resp *http.Response
	require.Eventually(t, func() bool {
		r, err := http.Get(url)
		if err != nil {
			return false
		}
		resp = r
		return true
	}, 5*time.Second, 50*time.Millisecond, "server never started listening")
	defer resp.Body.Close()

	require.NotNil(t, gotScope)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "test_app_metrics:test_counter 1")
}
