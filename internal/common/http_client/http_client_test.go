package http_client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSimpleGetRequest(t *testing.T) {
	resp := <-DefaultHttpClient.NewRequestWithTimeout(context.Background(), 10*time.Millisecond).GetAsync("http://ip-api.com/json/24.48.0.1")

	if resp.Err != nil {
		t.Fatal(resp.Err)
	}

	assert.Equal(t, "{\"status\":\"success\",\"country\":\"Canada\",\"countryCode\":\"CA\",\"region\":\"QC\",\"regionName\":\"Quebec\",\"city\":\"Montreal\",\"zip\":\"H1A\",\"lat\":45.6752,\"lon\":-73.5022,\"timezone\":\"America/Toronto\",\"isp\":\"Le Groupe Videotron Ltee\",\"org\":\"Videotron Ltee\",\"as\":\"AS5769 Videotron Telecom Ltee\",\"query\":\"24.48.0.1\"}",
		string(resp.Resp.String()))
}
