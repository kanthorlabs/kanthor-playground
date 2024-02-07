package kanthorsdk

import (
	"context"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type (
	ApplicationCreateReq = openapi.ApplicationCreateReq
	ApplicationCreateRes = openapi.ApplicationCreateRes
	ApplicationGetRes    = openapi.ApplicationGetRes
	ApplicationUpdateReq = openapi.ApplicationUpdateReq
	ApplicationUpdateRes = openapi.ApplicationUpdateRes
	ApplicationDeleteRes = openapi.ApplicationDeleteRes
	ApplicationListRes   = openapi.ApplicationListRes
)

type Application struct {
	api *openapi.APIClient
}

func (instance *Application) Create(ctx context.Context, req *ApplicationCreateReq) (*ApplicationCreateRes, error) {
	request := instance.api.ApplicationAPI.ApplicationPost(ctx)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Application) Get(ctx context.Context, id string) (*ApplicationGetRes, error) {
	request := instance.api.ApplicationAPI.ApplicationAppIdGet(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Application) Update(ctx context.Context, id string, req *ApplicationUpdateReq) (*ApplicationUpdateRes, error) {
	request := instance.api.ApplicationAPI.ApplicationAppIdPatch(ctx, id)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *Application) List(ctx context.Context, queries ...Query) (*openapi.ApplicationListRes, error) {
	q := &Queries{}
	for _, query := range queries {
		query(q)
	}

	request := instance.api.ApplicationAPI.ApplicationGet(ctx)
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

func (instance *Application) Delete(ctx context.Context, id string) (*ApplicationDeleteRes, error) {
	request := instance.api.ApplicationAPI.ApplicationAppIdDelete(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}
