package kanthorsdk

import (
	"context"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type (
	EndpointCreateReq = openapi.EndpointCreateReq
	EndpointCreateRes = openapi.EndpointCreateRes
	EndpointGetRes    = openapi.EndpointGetRes
	EndpointUpdateReq = openapi.EndpointUpdateReq
	EndpointUpdateRes = openapi.EndpointUpdateRes
	EndpointDeleteRes = openapi.EndpointDeleteRes
	EndpointListRes   = openapi.EndpointListRes
)

type Endpoint struct {
	api *openapi.APIClient
}

func (instance *Endpoint) Create(ctx context.Context, req *EndpointCreateReq) (*EndpointCreateRes, error) {
	request := instance.api.EndpointAPI.EndpointPost(ctx)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Endpoint) Get(ctx context.Context, id string) (*EndpointGetRes, error) {
	request := instance.api.EndpointAPI.EndpointEpIdGet(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Endpoint) Update(ctx context.Context, id string, req *EndpointUpdateReq) (*EndpointUpdateRes, error) {
	request := instance.api.EndpointAPI.EndpointEpIdPatch(ctx, id)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Endpoint) List(ctx context.Context, queries ...Query) (*openapi.EndpointListRes, error) {
	q := &Queries{}
	for _, query := range queries {
		query(q)
	}

	request := instance.api.EndpointAPI.EndpointGet(ctx)
	request = request.Id(q.Id)
	request = request.Q(q.Search)
	request = request.Limit(q.Limit)
	request = request.Page(q.Page)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Endpoint) Delete(ctx context.Context, id string) (*EndpointDeleteRes, error) {
	request := instance.api.EndpointAPI.EndpointEpIdDelete(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}
