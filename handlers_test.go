package main

import (
	"github.com/gin-gonic/gin"
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
			name: "ShouldParse_ResetAfter",
			args: args{
				context: &gin.Context{
					Params: []gin.Param{
						{Key: "job", Value: "/tank"},
					},
				},
				query: "/tank?ResetAfter=10s",
			},
			want: Parameters{
				JobName:    "tank",
				ResetAfter: 10 * time.Second,
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
