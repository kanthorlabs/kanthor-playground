package kanthorsdk

import (
	"net/http"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type Error struct {
	status int
	body   []byte
	error  string
}

func (e Error) Error() string {
	return e.error
}

func (e Error) Body() []byte {
	return e.body
}

func (e Error) Status() int {
	return e.status
}

func errorify(err error, res *http.Response) error {
	if openapiError, ok := err.(openapi.GenericOpenAPIError); ok {
		e := &Error{
			body:  openapiError.Body(),
			error: openapiError.Error(),
		}
		if res != nil {
			e.status = res.StatusCode
		}
		return e
	}
	return &Error{error: err.Error()}
}

var (
	ErrWebhookHeadersMissing = &Error{error: "WEBHOOK.HEADERS.MISSING_REQUIRED_PROPERTIES"}
	ErrWebhookHeadersTs      = &Error{error: "WEBHOOK.HEADERS.TIMESTAMP_INVALID"}
	ErrWebhookMessageTooOld  = &Error{error: "WEBHOOK.MESSAGE.TIMESTAMP_TOO_OLD"}
	ErrWebhookMessageTooNew  = &Error{error: "WEBHOOK.MESSAGE.TIMESTAMP_TOO_NEW"}
	ErrWebhookSignNotMatch   = &Error{error: "WEBHOOK.SIGNATURE.NOT_MATCH"}
)
