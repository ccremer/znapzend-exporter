package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestParseAndValidateInput(t *testing.T) {
	type args struct {
		context *gin.Context
		query   string
	}
	tests := []struct {
		name    string
		args    args
		want    Parameters
		wantErr bool
	}{
		{
			name: "ShouldParse_JobName",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank/backup"},
					},
				},
				query: "/tank/backup",
			},
			want: Parameters{JobName: "tank/backup"},
		},
		{
			name: "ShouldFail_AtParsingJobName",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{},
				},
				query: "/",
			},
			want:    Parameters{},
			wantErr: true,
		},
		{
			name: "ShouldParse_SelfResetAfter",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?SelfResetAfter=10s",
			},
			want: Parameters{
				JobName:        "tank",
				SelfResetAfter: 10 * time.Second,
			},
		},
		{
			name: "ShouldParse_BooleanFlags",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?ResetPreSnap=true&ResetPostSnap=true&ResetPreSend=true&ResetPostSend=true",
			},
			want: Parameters{
				JobName:       "tank",
				ResetPreSnap:  true,
				ResetPostSnap: true,
				ResetPreSend:  true,
				ResetPostSend: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.args.query, nil)
			tt.args.context.Request = req
			got, err := ParseAndValidateInput(tt.args.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
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
			name: "PreSnap_ShouldSetMetric",
			args: args{
				query: "/presnap/pool",
			},
			expectations: []expectation{
				{gauge: preSnapMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "PostSnap_ShouldSetMetric",
			args: args{
				query: "/postsnap/pool",
			},
			expectations: []expectation{
				{gauge: postSnapMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "PostSnap_ShouldReset_PreSnap",
			args: args{
				query: "/postsnap/pool?ResetPreSnap=true",
			},
			expectations: []expectation{
				{gauge: preSnapMetric.WithLabelValues("pool"), initial: 1, expected: 0},
				{gauge: postSnapMetric.WithLabelValues("pool"), initial: 0, expected: 1},
			},
		},
		{
			name: "PreSend_ShouldSetMetric",
			args: args{
				query: "/presend/pool",
			},
			expectations: []expectation{
				{gauge: preSendMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "PreSend_ShouldReset_PostSnap",
			args: args{
				query: "/presend/pool?ResetPostSnap=true",
			},
			expectations: []expectation{
				{gauge: postSnapMetric.WithLabelValues("pool"), initial: 1, expected: 0},
				{gauge: preSendMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "PostSend_ShouldSetMetric",
			args: args{
				query: "/postsend/pool?ResetPostSnap=true",
			},
			expectations: []expectation{
				{gauge: postSendMetric.WithLabelValues("pool"), expected: 1},
			},
		},
		{
			name: "PostSend_ShouldReset_PreSend",
			args: args{
				query: "/postsend/pool?ResetPreSend=true",
			},
			expectations: []expectation{
				{gauge: preSendMetric.WithLabelValues("pool"), initial: 1, expected: 0},
				{gauge: postSendMetric.WithLabelValues("pool"), expected: 1},
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
