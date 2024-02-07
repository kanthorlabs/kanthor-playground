package kanthorsdk

import (
	"context"

	"github.com/kanthorlabs/kanthor-sdk-go/internal/openapi"
)

type (
	AccountGetRes = openapi.AccountGetRes
)

type Account struct {
	api *openapi.APIClient
}

func (instance *Account) Get(ctx context.Context) (*AccountGetRes, error) {
	request := instance.api.AccountAPI.AccountMeGet(ctx)

	response, res, err := request.Execute()
	if err != nil {
		return nil, errorify(err, res)
	}

	return response, nil
}
