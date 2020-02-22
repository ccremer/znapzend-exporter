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
				Parameters: Parameters{JobName: "pool", AutoReset: true, ResetAfter: time.Second},
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
			if err := j.setMetric(tt.args.vec); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.EqualValues(t, float64(1), testutil.ToFloat64(preSendMetric))
			if tt.fields.Parameters.AutoReset {
				time.Sleep(1200 * time.Millisecond)
				assert.EqualValues(t, float64(0), testutil.ToFloat64(preSendMetric))
			}
		})
	}
}

func TestJobContext_ResetMetricIf(t *testing.T) {
	type fields struct {
		Parameters Parameters
		Context    *gin.Context
	}
	type args struct {
		condition bool
		vec       *prometheus.GaugeVec
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected float64
	}{
		{
			name: "ShouldResetMetric_IfTrue",
			fields: fields{
				Parameters: Parameters{JobName: "tank"},
			},
			args:     args{vec: preSendMetric, condition: true},
			expected: 0,
		},
		{
			name: "ShouldNotResetMetric_IfFalse",
			fields: fields{
				Parameters: Parameters{JobName: "tank"},
			},
			args:     args{vec: preSendMetric},
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
			if err := j.resetMetricIf(tt.args.condition, tt.args.vec); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.EqualValues(t, tt.expected, testutil.ToFloat64(gauge))
		})
	}
}
