package http_client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestSimpleGetRequest(t *testing.T) {
	resp := <-SendHttpRequestAsync(context.Background(), "http://ip-api.com/json/24.48.0.1",
		"GET", "application/json", http.MethodPost, nil, false, 10*time.Second)

	assert.Nil(t, resp.GetError())
	assert.Equal(t, "{\"status\":\"success\",\"country\":\"Canada\",\"countryCode\":\"CA\",\"region\":\"QC\",\"regionName\":\"Quebec\",\"city\":\"Montreal\",\"zip\":\"H1A\",\"lat\":45.6752,\"lon\":-73.5022,\"timezone\":\"America/Toronto\",\"isp\":\"Le Groupe Videotron Ltee\",\"org\":\"Videotron Ltee\",\"as\":\"AS5769 Videotron Telecom Ltee\",\"query\":\"24.48.0.1\"}",
		string(resp.GetRawResponse()))
}
