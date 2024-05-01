package helper

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type WebResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type WebErrorInputResponse struct {
	Code       int             `json:"code"`
	Status     string          `json:"status"`
	ErrorField json.RawMessage `json:"errorField"`
}

type ServiceResponse struct {
	Message string `json:"message"`
}

func NewHTTPErrorHandler(ctx *fiber.Ctx, err error) error {
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	switch e := err.(type) {
	case *HTTPError:
		response := WebResponse{
			Code:    e.Code,
			Status:  http.StatusText(e.Code),
			Message: e.Error(),
		}
		return ctx.Status(e.Code).JSON(response)
	case *HTTPInputValidationError:
		response := WebErrorInputResponse{
			Code:       fiber.StatusBadRequest,
			Status:     http.StatusText(fiber.StatusBadRequest),
			ErrorField: json.RawMessage(e.Error()),
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(response)
	case *fiber.Error:
		response := WebResponse{
			Code:    e.Code,
			Status:  http.StatusText(e.Code),
			Message: e.Error(),
		}
		return ctx.Status(e.Code).JSON(response)
	default:
		response := WebResponse{
			Code:    fiber.StatusInternalServerError,
			Status:  http.StatusText(fiber.StatusInternalServerError),
			Message: e.Error(),
		}
		return ctx.Status(response.Code).JSON(response)
	}
}

type HTTPError struct {
	Code int
	Err  error
}

func NewHTTPError(code int, err error) error {
	return &HTTPError{
		Code: code,
		Err:  err,
	}
}

func (e *HTTPError) Error() string {
	return e.Err.Error()
}

type HTTPInputValidationError struct {
	Err error
}

func NewHTTPInputValidationError(err error) error {
	return &HTTPInputValidationError{Err: err}
}

func (e *HTTPInputValidationError) Error() string {
	return e.Err.Error()
}
