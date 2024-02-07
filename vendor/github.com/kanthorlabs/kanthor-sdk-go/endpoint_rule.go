package kanthorsdk

import (
	"context"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type EndpointRule struct {
	api *openapi.APIClient
}

type (
	EndpointRuleCreateReq = openapi.EndpointRuleCreateReq
	EndpointRuleCreateRes = openapi.EndpointRuleCreateRes
	EndpointRuleGetRes    = openapi.EndpointRuleGetRes
	EndpointRuleUpdateReq = openapi.EndpointRuleUpdateReq
	EndpointRuleUpdateRes = openapi.EndpointRuleUpdateRes
	EndpointRuleDeleteRes = openapi.EndpointRuleDeleteRes
	EndpointRuleListRes   = openapi.EndpointRuleListRes
)

func (instance *EndpointRule) Create(ctx context.Context, req *EndpointRuleCreateReq) (*EndpointRuleCreateRes, error) {
	request := instance.api.EndpointRuleAPI.RulePost(ctx)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *EndpointRule) Get(ctx context.Context, id string) (*EndpointRuleGetRes, error) {
	request := instance.api.EndpointRuleAPI.RuleEprIdGet(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *EndpointRule) Update(ctx context.Context, id string, req *EndpointRuleUpdateReq) (*EndpointRuleUpdateRes, error) {
	request := instance.api.EndpointRuleAPI.RuleEprIdPatch(ctx, id)
	request = request.Payload(*req)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}

func (instance *EndpointRule) List(ctx context.Context, queries ...Query) (*openapi.EndpointRuleListRes, error) {
	q := &Queries{}
	for _, query := range queries {
		query(q)
	}

	request := instance.api.EndpointRuleAPI.RuleGet(ctx)
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

func (instance *EndpointRule) Delete(ctx context.Context, id string) (*EndpointRuleDeleteRes, error) {
	request := instance.api.EndpointRuleAPI.RuleEprIdDelete(ctx, id)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}
