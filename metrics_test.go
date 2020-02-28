package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJobContext_SetMetric(t *testing.T) {
	type fields struct {
		Parameters Parameters
		Context    *gin.Context
	}
	type args struct {
		vec *prometheus.GaugeVec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ShouldSetMetricTo1",
			fields: fields{
				Parameters: Parameters{JobName: "pool"},
			},
			args: args{vec: preSendMetric},
		},
		{
			name: "ShouldResetMetric",
			fields: fields{
				Parameters: Parameters{JobName: "pool", SelfResetAfter: time.Second},
			},
			args: args{vec: preSendMetric},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JobContext{
				Parameters: tt.fields.Parameters,
				Context:    tt.fields.Context,
			}
			j.setMetric(tt.args.vec)
			assert.EqualValues(t, float64(1), testutil.ToFloat64(preSendMetric))
			if tt.fields.Parameters.SelfResetAfter > 0 {
				time.Sleep(1200 * time.Millisecond)
				assert.EqualValues(t, float64(0), testutil.ToFloat64(preSendMetric))
			}
		})
	}
}

func TestJobContext_ResetMetrics(t *testing.T) {
	type fields struct {
		Parameters Parameters
		Context    *gin.Context
	}
	type args struct {
		tuples []ResetMetricTuple
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected float64
	}{
		{
			name: "ShouldResetMetric_IfTrue",
			fields: fields{
				Parameters: Parameters{JobName: "tank"},
			},
			args: args{tuples: []ResetMetricTuple{
				{resetEnabled: true, vec: preSendMetric},
			}},
			expected: 0,
		},
		{
			name: "ShouldNotResetMetric_IfFalse",
			fields: fields{
				Parameters: Parameters{JobName: "tank"},
			},
			args: args{tuples: []ResetMetricTuple{
				{resetEnabled: false, vec: preSendMetric},
			}},
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gauge := preSendMetric.WithLabelValues(tt.fields.Parameters.JobName)
			gauge.Set(1)
			j := &JobContext{
				Parameters: tt.fields.Parameters,
				Context:    tt.fields.Context,
			}
			j.ResetMetrics(tt.args.tuples[:]...)
			assert.EqualValues(t, tt.expected, testutil.ToFloat64(gauge))
		})
	}
}
