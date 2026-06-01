package configx_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bytechainx/configx/pkg/configx"
	"github.com/bytechainx/configx/testkit"
)

func TestHealthStatusJSONGolden(t *testing.T) {
	payload, err := json.Marshal(configx.HealthStatus{
		Name:      "configx",
		Status:    configx.HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Unix(0, 0).UTC(),
		LatencyMs: 7,
		Metadata: map[string]string{
			"kind": "template",
		},
	})
	if err != nil {
		t.Fatalf("marshal health status: %v", err)
	}

	payload = append(payload, '\n')
	testkit.RequireGolden(t, "testdata/golden/health_status.json", payload)
}
