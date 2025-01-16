package models

import (
	"net/http"
	"testing"

	"github.com/deasdania/dating-app/status"
	"github.com/stretchr/testify/assert"
)

func TestNewResponseError(t *testing.T) {
	type args struct {
		httpStatus int64
		statusCode status.DatingStatusCode
		msg        string
	}
	tests := []struct {
		name string
		args args
		want *ResponseBase
	}{
		// TODO: Add test cases.
		{
			name: "error",
			args: args{
				httpStatus: http.StatusBadRequest,
				statusCode: status.UserErrCode_InvalidRequest,
				msg:        "invalid",
			},
			want: &ResponseBase{
				Status: http.StatusBadRequest,
				Details: status.StatusResponse{
					StatusCode: string(status.UserErrCode_InvalidRequest),
					StatusDesc: "Invalid request: invalid",
				},
			},
		},
		{
			name: "error",
			args: args{
				httpStatus: http.StatusBadRequest,
				statusCode: status.UserErrCode_InvalidRequestPremiumPackage,
				msg:        "",
			},
			want: &ResponseBase{
				Status: http.StatusBadRequest,
				Details: status.StatusResponse{
					StatusCode: string(status.UserErrCode_InvalidRequestPremiumPackage),
					StatusDesc: "type is required ('remove_quota' or 'verified_label')",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewResponseError(tt.args.httpStatus, tt.args.statusCode, tt.args.msg)
			assert.Equal(t, got.Details.StatusCode, string(tt.want.Details.StatusCode))
			assert.Equal(t, got.Details.StatusDesc, string(tt.want.Details.StatusDesc))
		})
	}
}
