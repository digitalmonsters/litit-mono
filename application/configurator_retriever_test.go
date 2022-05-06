package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHttpRetriever(t *testing.T) {
	port := "8090"

	go func() {
		http.HandleFunc("/json", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			d, err := ioutil.ReadAll(req.Body)

			if err != nil {
				t.Error(err)
			}

			var r getConfigRequest

			if err := json.Unmarshal(d, &r); err != nil {
				t.Error(err)
			}

			assert.Equal(t, []string{"1234", "4567"}, r.Items)

			_, _ = w.Write([]byte("{\n  \"BoolValue\": \"true\",\n  \"StringValue\": \"Totally Random value\",\n  \"IntValue\": \"5566\",\n  \"Int64Value\": \"32563246435322\",\n  \"DecimalValue\": \"225.6852\"\n}"))
		})

		_ = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", port), nil)
	}()

	time.Sleep(100 * time.Millisecond)

	retr := NewHttpRetriever(fmt.Sprintf("http://127.0.0.1:%v/json", port))

	val, err := retr.Retrieve([]string{"1234", "4567"}, context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, 5, len(val))
}
