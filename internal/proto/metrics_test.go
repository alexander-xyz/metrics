package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMetricGetters(t *testing.T) {
	metric := &Metric{
		Id:    "Alloc",
		Type:  Metric_GAUGE,
		Value: 12.5,
	}

	assert.Equal(t, "Alloc", metric.GetId())
	assert.Equal(t, Metric_GAUGE, metric.GetType())
	assert.InDelta(t, 12.5, metric.GetValue(), 0.001)
	assert.Zero(t, metric.GetDelta())

	counter := &Metric{Id: "PollCount", Type: Metric_COUNTER, Delta: 7}
	assert.Equal(t, int64(7), counter.GetDelta())
	assert.Equal(t, Metric_COUNTER, counter.GetType())
}

func TestMetricZeroValueGetters(t *testing.T) {
	var metric *Metric

	assert.Empty(t, metric.GetId())
	assert.Equal(t, Metric_GAUGE, metric.GetType())
	assert.Zero(t, metric.GetValue())
	assert.Zero(t, metric.GetDelta())
}

func TestMetricTypeNames(t *testing.T) {
	assert.Equal(t, "GAUGE", Metric_GAUGE.String())
	assert.Equal(t, "COUNTER", Metric_COUNTER.String())
	assert.Equal(t, Metric_MType(1), Metric_COUNTER)
	assert.NotNil(t, Metric_MType(0).Descriptor())
	assert.NotNil(t, Metric_GAUGE.Type())
	assert.EqualValues(t, 1, Metric_COUNTER.Number())
	assert.Equal(t, "GAUGE", Metric_MType_name[0])
	assert.EqualValues(t, 1, Metric_MType_value["COUNTER"])
}

func TestRequestRoundTrip(t *testing.T) {
	request := &UpdateMetricsRequest{
		Metrics: []*Metric{
			{Id: "Alloc", Type: Metric_GAUGE, Value: 1.5},
			{Id: "PollCount", Type: Metric_COUNTER, Delta: 3},
		},
	}

	data, err := proto.Marshal(request)
	require.NoError(t, err)

	var restored UpdateMetricsRequest
	require.NoError(t, proto.Unmarshal(data, &restored))

	require.Len(t, restored.GetMetrics(), 2)
	assert.Equal(t, "Alloc", restored.GetMetrics()[0].GetId())
	assert.Equal(t, int64(3), restored.GetMetrics()[1].GetDelta())
}

func TestMessagesSupportProtoInterface(t *testing.T) {
	request := &UpdateMetricsRequest{Metrics: []*Metric{{Id: "Alloc"}}}
	assert.NotEmpty(t, request.String())
	assert.NotNil(t, request.ProtoReflect())
	assert.NotNil(t, request.Descriptor)

	request.Reset()
	assert.Empty(t, request.GetMetrics())

	response := &UpdateMetricsResponse{}
	assert.NotNil(t, response.ProtoReflect())
	assert.Empty(t, response.String())

	response.Reset()

	metric := &Metric{Id: "Alloc"}
	assert.NotEmpty(t, metric.String())
	assert.NotNil(t, metric.ProtoReflect())

	metric.Reset()
	assert.Empty(t, metric.GetId())
}

func TestEmptyRequestGetters(t *testing.T) {
	var request *UpdateMetricsRequest

	assert.Empty(t, request.GetMetrics())
}
