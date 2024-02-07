package kanthorsdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	WebhookVersion    = "v1"
	HeaderWebhookId   = "webhook-id"
	HeaderWebhookTs   = "webhook-timestamp"
	HeaderWebhookSign = "webhook-signature"
)

func NewWebhook(secret string, options ...WebhookOption) (*Webhook, error) {
	opts := &WebhookOptions{
		tolerance: time.Minute * 5,
	}
	for _, option := range options {
		option(opts)
	}

	return &Webhook{secret: []byte(secret), options: opts}, nil
}

type Webhook struct {
	secret  []byte
	options *WebhookOptions
}

func (instance *Webhook) Sign(id string, ts int64, payload []byte) (string, error) {
	sign := fmt.Sprintf("%s.%d.%s", id, ts, payload)

	mac := hmac.New(sha256.New, instance.secret)
	mac.Write([]byte(sign))

	signed := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s,%s", WebhookVersion, signed), nil
}

func (instance *Webhook) Verify(payload []byte, headers http.Header) error {
	id := headers.Get(HeaderWebhookId)
	timestamp := headers.Get(HeaderWebhookTs)
	signature := headers.Get(HeaderWebhookSign)
	if id == "" || timestamp == "" || signature == "" {
		return ErrWebhookHeadersMissing
	}

	ts, err := instance.timestamp(timestamp)
	if err != nil {
		return err
	}

	signed, err := instance.Sign(id, ts, payload)
	if err != nil {
		return err
	}

	return instance.compare(signed, signature)
}

func (instance *Webhook) timestamp(ts string) (int64, error) {
	init, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return 0, ErrWebhookHeadersTs
	}

	now := time.Now().UnixMilli()

	old := (now - init) > instance.options.tolerance.Milliseconds()
	if old {
		return 0, ErrWebhookMessageTooOld
	}

	new := (init - now) > instance.options.tolerance.Milliseconds()
	if new {
		return 0, ErrWebhookMessageTooNew
	}

	return init, nil
}

func (instance *Webhook) compare(expected, actual string) error {
	expectedParts := strings.Split(expected, ",")

	actualSigns := strings.Split(actual, " ")
	for _, actualSign := range actualSigns {
		actualParts := strings.Split(actualSign, ",")

		if expectedParts[0] == actualParts[0] {
			if hmac.Equal([]byte(expectedParts[1]), []byte(actualParts[1])) {
				return nil
			}
		}
	}

	return ErrWebhookSignNotMatch
}

type WebhookOptions struct {
	tolerance time.Duration
}

type WebhookOption func(*WebhookOptions)

func WithWebhookTolerance(duration time.Duration) WebhookOption {
	return func(wo *WebhookOptions) {
		wo.tolerance = duration
	}
}
