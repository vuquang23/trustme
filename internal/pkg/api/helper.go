package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vuquang23/trustme/internal/pkg/util/requestid"
	"github.com/vuquang23/trustme/pkg/logger"
)

var ErrorResponseByError = map[error]ErrorResponse{}

type SuccessResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
}

type ErrorResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`

	HTTPStatus int    `json:"-"`
	RequestID  string `json:"requestId"`
}

var DefaultErrorResponse = ErrorResponse{
	HTTPStatus: http.StatusInternalServerError,
	Code:       500,
	Message:    "internal server error",
}

func RespondSuccess(c *gin.Context, data interface{}) {
	successResponse := SuccessResponse{
		Code:      0,
		Message:   "successfully",
		Data:      data,
		RequestID: requestid.ExtractRequestID(c),
	}

	c.JSON(
		http.StatusOK,
		successResponse,
	)
}

func RespondFailure(c *gin.Context, err error) {
	if errors.Is(err, context.Canceled) {
		respondContextCanceledError(c)
		return
	}

	requestID := requestid.ExtractRequestID(c)
	response := responseFromErr(err)
	response.RequestID = requestID

	logger.
		WithFields(c, logger.Fields{"request.id": requestID, "error": err}).
		Warn("respond failure")

	c.JSON(
		response.HTTPStatus,
		response,
	)
}

func responseFromErr(err error) ErrorResponse {
	for {
		if err == nil {
			return DefaultErrorResponse
		}

		if resp, ok := ErrorResponseByError[err]; ok {
			return resp
		}

		err = errors.Unwrap(err)
	}
}

const ClientClosedRequestStatusCode = 499

func respondContextCanceledError(c *gin.Context) {
	errorResponse := ErrorResponse{
		Code:      4990,
		Message:   "request was canceled",
		RequestID: requestid.ExtractRequestID(c),
	}

	c.JSON(
		ClientClosedRequestStatusCode,
		errorResponse,
	)
}
