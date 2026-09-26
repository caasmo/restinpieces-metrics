package metrics

import (
	"net/http"
	"testing"
	"time"

	"github.com/caasmo/restinpieces/core"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDefaultRecorder_Record(t *testing.T) {
	recorder := NewDefaultRecorder()

	recorder.Record(&core.ResponseRecorder{
		Status:    http.StatusOK,
		StartTime: time.Now(),
	})

	if got := testutil.ToFloat64(recorder.RequestsTotal.WithLabelValues("200")); got != 1 {
		t.Errorf("counter for code 200 = %v, want 1", got)
	}

	if got := testutil.CollectAndCount(recorder.RequestDuration, requestDuration); got != 1 {
		t.Errorf("histogram series = %d, want 1", got)
	}
}
