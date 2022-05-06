package application

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testingConfig struct {
	BoolValue    bool
	StringValue  string
	DecimalValue decimal.Decimal
	IntValue     int
	Int64Value   int64
}

func TestConfigurator(t *testing.T) {
	configurator := NewConfigurator[testingConfig]().
		WithRetriever(NewFileRetriever("./test_data/configurator.json")).
		MustInit()

	assert.Equal(t, 5566, configurator.Values.IntValue)
	assert.Equal(t, int64(32563246435322), configurator.Values.Int64Value)
	assert.Equal(t, "Totally Random value", configurator.Values.StringValue)
	assert.Equal(t, true, configurator.Values.BoolValue)
	assert.Equal(t, "225.6852", configurator.Values.DecimalValue.StringFixed(4))
}

type mockRetriever struct {
	counter int
}

type miniConfig struct {
	IntVal int
}

func (m *mockRetriever) Retrieve(_ []string, _ context.Context) (map[string]string, error) {
	m.counter += 1

	return map[string]string{"IntVal": fmt.Sprint(m.counter)}, nil
}

func TestTimeBasedChanges(t *testing.T) {
	mm := &mockRetriever{}
	configurator := NewConfigurator[miniConfig]().
		WithRetriever(mm).
		WithInterval(50 * time.Millisecond).
		MustInit()

	assert.Equal(t, 1, configurator.Values.IntVal)

	time.Sleep(120 * time.Millisecond)

	fmt.Println(configurator.Values.IntVal)

	assert.NotEqual(t, 1, configurator.Values.IntVal)
	assert.Equal(t, mm.counter, configurator.Values.IntVal)
}
