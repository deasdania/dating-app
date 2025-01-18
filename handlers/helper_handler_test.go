package handlers

import (
	"net/http"
	"testing"

	"github.com/deasdania/dating-app/status"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func TestHandlers_RespondWithError(t *testing.T) {

	type args struct {
		c          echo.Context
		statusCode int64
		errCode    status.DatingStatusCode
		errMsg     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "generate error",
			args: args{
				statusCode: http.StatusUnauthorized,
				errCode:    status.UserErrCode_Unauthorized,
				errMsg:     "message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handlers{
				Log: &logrus.Entry{},
			}
			if err := h.RespondWithError(tt.args.c, tt.args.statusCode, tt.args.errCode, tt.args.errMsg); (err != nil) != tt.wantErr {
				t.Errorf("Handlers.RespondWithError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
