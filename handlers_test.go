package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
)

func (p Job) Initialize() Job {
	p.ResetPreSnap = true
	p.ResetPostSnap = true
	p.ResetPreSend = true
	p.ResetPostSend = true
	return p
}

func TestParseAndValidateInput(t *testing.T) {
	type args struct {
		context *gin.Context
		query   string
	}
	tests := []struct {
		name    string
		args    args
		want    Job
		wantErr bool
	}{
		{
			name: "GivenQuery_WhenValidJob_ThenParseName",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank/backup"},
					},
				},
				query: "/tank/backup",
			},
			want: Job{JobName: "tank/backup"}.Initialize(),
		},
		{
			name: "GivenQuery_WhenInvalidQuery_ThenThrowError",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{},
				},
				query: "/",
			},
			want:    Job{}.Initialize(),
			wantErr: true,
		},
		{
			name: "GivenQueryWithParameter_WhenInvalidBooleanParameter_ThenThrowError",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?ResetPreSnap=asdf",
			},
			want:    Job{JobName: "tank"}.Initialize(),
			wantErr: true,
		},
		{
			name: "GivenQueryWithParameter_WhenQueryContainsSelfResetAfter_ThenParseDuration",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?SelfResetAfter=10s",
			},
			want: Job{
				JobName:        "tank",
				SelfResetAfter: 10 * time.Second,
			}.Initialize(),
		},
		{
			name: "GivenQueryWithParameters_WhenBooleanParameters_ThenParseAll",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?ResetPreSnap=false&ResetPostSnap=false&ResetPreSend=false&ResetPostSend=false",
			},
			want: Job{
				JobName:       "tank",
				ResetPreSnap:  false,
				ResetPostSnap: false,
				ResetPreSend:  false,
				ResetPostSend: false,
			},
		},
		{
			name: "GivenPreSendQuery_WhenNoHostGiven_ThenThrowError",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/presend/tank",
			},
			wantErr: true,
		},
		{
			name: "GivenPostsendQuery_WhenNoHostGiven_ThenThrowError",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/postsend/tank",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.args.query, nil)
			tt.args.context.Request = req
			got, err := ParseAndValidateInput(tt.args.context)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_handleCommands(t *testing.T) {
	type args struct {
		query string
	}
	type expectation struct {
		gauge    prometheus.Gauge
		initial  float64
		expected float64
	}
	tests := []struct {
		name         string
		args         args
		expectations []expectation
	}{
		{
			name: "GivenPreSnapQuery_WhenJobValid_ThenSetMetricValue",
			args: args{
				query: "/presnap/pool",
			},
			expectations: []expectation{
				{gauge: preSnapMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "GivenPostSnapQuery_WhenJobValid_ThenSetMetric",
			args: args{
				query: "/postsnap/pool",
			},
			expectations: []expectation{
				{gauge: postSnapMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "GivenPostSnapQuery_WhenResetPreSnapTrue_ThenResetPreSnap",
			args: args{
				query: "/postsnap/pool?ResetPreSnap=true&TargetHost=host",
			},
			expectations: []expectation{
				{gauge: preSnapMetric.WithLabelValues("pool"), initial: 1, expected: 0},
				{gauge: postSnapMetric.WithLabelValues("pool"), initial: 0, expected: 1},
			},
		},
		{
			name: "GivenPreSendQuery_WhenHostGiven_ThenSetMetric",
			args: args{
				query: "/presend/pool?TargetHost=host",
			},
			expectations: []expectation{
				{gauge: preSendMetric.WithLabelValues("pool", "host"), expected: 1},
			},
		},
		{
			name: "GivenPreSendQuery_WhenResetPostSnapTrue_ThenResetPostSnap",
			args: args{
				query: "/presend/pool?ResetPostSnap=true&TargetHost=host",
			},
			expectations: []expectation{
				{gauge: postSnapMetric.WithLabelValues("pool"), initial: 1, expected: 0},
				{gauge: preSendMetric.WithLabelValues("pool", "host"), expected: 1},
			},
		},
		{
			name: "GivenPostSendQuery_WhenResetPostSnapTrue_ThenSetMetric",
			args: args{
				query: "/postsend/pool?ResetPostSnap=true&TargetHost=host",
			},
			expectations: []expectation{
				{gauge: postSendMetric.WithLabelValues("pool", "host"), expected: 1},
			},
		},
		{
			name: "GivenPostSendQuery_WhenResetPreSendTrue_ThenResetPreSend",
			args: args{
				query: "/postsend/pool?ResetPreSend=true&TargetHost=host",
			},
			expectations: []expectation{
				{gauge: preSendMetric.WithLabelValues("pool", "host"), initial: 1, expected: 0},
				{gauge: postSendMetric.WithLabelValues("pool", "host"), expected: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectation := range tt.expectations {
				expectation.gauge.Set(expectation.initial)
			}
			//log.SetLevel(log.DebugLevel)
			req := httptest.NewRequest("GET", tt.args.query, nil)
			w := httptest.NewRecorder()
			r := SetupRouter()
			r.ServeHTTP(w, req)

			for _, expectation := range tt.expectations {
				assert.EqualValues(t, expectation.expected, testutil.ToFloat64(expectation.gauge), "Gauge: %s", expectation.gauge.Desc())

			}
		})
	}
}
