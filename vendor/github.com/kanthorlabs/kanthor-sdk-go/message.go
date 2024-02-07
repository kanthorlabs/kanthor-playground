package kanthorsdk

import (
	"context"
	"encoding/json"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type (
	MessageCreateReq struct {
		AppId   string
		Body    map[string]any
		Headers map[string]string
		Type    string
	}
	MessageCreateRes = openapi.MessageCreateRes
)

type Message struct {
	api *openapi.APIClient
}

func (instance *Message) Create(ctx context.Context, req *MessageCreateReq) (*MessageCreateRes, error) {
	body, err := json.Marshal(req.Body)
	if err != nil {
		return nil, err
	}

	request := instance.api.MessageAPI.MessagePost(ctx)
	request = request.Payload(openapi.MessageCreateReq{
		AppId:   req.AppId,
		Type:    req.Type,
		Headers: req.Headers,
		Body:    string(body),
	})

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}
