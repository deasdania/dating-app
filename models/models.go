package models

import (
	"fmt"
	"strings"

	"github.com/deasdania/dating-app/status"
)

type ResponseBase struct {
	Status  int64                 `json:"status"`
	Details status.StatusResponse `json:"details"`
	Data    interface{}           `json:"data"`
}

// NewResponseError creates a new error response with the given status code and message
func NewResponseError(httpStatus int64, statusCode status.DatingStatusCode, msg string) *ResponseBase {
	statusRes := status.ResponseFromCode(statusCode)
	if strings.Contains(statusRes.StatusDesc, ":") {
		statusRes.StatusDesc = fmt.Sprintf(statusRes.StatusDesc, msg)
	}
	return &ResponseBase{
		Status:  httpStatus,
		Details: statusRes,
	}
}
