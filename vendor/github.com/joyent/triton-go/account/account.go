//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package account

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"time"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
)

type Account struct {
	ID               string    `json:"id"`
	Login            string    `json:"login"`
	Email            string    `json:"email"`
	CompanyName      string    `json:"companyName"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Address          string    `json:"address"`
	PostalCode       string    `json:"postalCode"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Phone            string    `json:"phone"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
	TritonCNSEnabled bool      `json:"triton_cns_enabled"`
}

type GetInput struct{}

func (c AccountClient) Get(ctx context.Context, input *GetInput) (*Account, error) {
	fullPath := path.Join("/", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get account details")
	}

	var result *Account
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get account details")
	}

	return result, nil
}

type UpdateInput struct {
	Email            string `json:"email,omitempty"`
	CompanyName      string `json:"companyName,omitempty"`
	FirstName        string `json:"firstName,omitempty"`
	LastName         string `json:"lastName,omitempty"`
	Address          string `json:"address,omitempty"`
	PostalCode       string `json:"postalCode,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	Country          string `json:"country,omitempty"`
	Phone            string `json:"phone,omitempty"`
	TritonCNSEnabled bool   `json:"triton_cns_enabled,omitempty"`
}

// UpdateAccount updates your account details with the given parameters.
// TODO(jen20) Work out a safe way to test this
func (c AccountClient) Update(ctx context.Context, input *UpdateInput) (*Account, error) {
	fullPath := path.Join("/", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update account")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	var result *Account
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode update account response")
	}

	return result, nil
}
