package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vuquang23/trustme/internal/pkg/api/validator"
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

type DetailsBadRequest struct {
	FieldViolations []*DetailBadRequestFieldViolation `json:"fieldViolations"`
}

type DetailBadRequestFieldViolation struct {
	Field       string `json:"field"`
	Description string `json:"description"`
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
	var validationErr *validator.ValidationError
	if errors.As(err, &validationErr) {
		respondValidationError(c, validationErr)
		return
	}

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

func respondValidationError(c *gin.Context, err *validator.ValidationError) {
	errorResponse := ErrorResponse{
		Code:    4000,
		Message: "bad request",
		Details: []interface{}{
			&DetailsBadRequest{
				FieldViolations: []*DetailBadRequestFieldViolation{
					{
						Field:       err.Field,
						Description: err.Description,
					},
				},
			},
		},
		RequestID: requestid.ExtractRequestID(c),
	}

	c.JSON(
		http.StatusBadRequest,
		errorResponse,
	)
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
