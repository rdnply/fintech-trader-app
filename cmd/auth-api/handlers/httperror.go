package handlers

import (
	"encoding/json"

	"github.com/pkg/errors"
)

var _ ClientError = &HTTPError{}

type HTTPError struct {
	Context string `json:"-"`
	Cause   error  `json:"-"`
	Detail  string `json:"error,omitempty"`
	Status  int    `json:"-"`
}

func NewHTTPError(ctx string, err error, detail string, status int) error {
	return &HTTPError{
		Context: ctx,
		Cause:   err,
		Detail:  detail,
		Status:  status,
	}
}

func (e *HTTPError) Error() string {
	if e.Cause == nil {
		return e.Context
	}

	return e.Context + ": " + e.Cause.Error()
}

func (e *HTTPError) ResponseBody() ([]byte, error) {
	if e.Detail == "" {
		return nil, nil
	}

	body, err := json.Marshal(e)
	if err != nil {
		return nil, errors.Wrap(err, "error while parsing response body")
	}

	return body, nil
}

func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json",
	}
}

