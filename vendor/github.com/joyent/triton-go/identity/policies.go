//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package identity

import (
	"context"
	"encoding/json"
	"net/http"

	"path"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
)

type PoliciesClient struct {
	client *client.Client
}

type Policy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Rules       []string `json:"rules"`
	Description string   `json:"description"`
}

type ListPoliciesInput struct{}

func (c *PoliciesClient) List(ctx context.Context, _ *ListPoliciesInput) ([]*Policy, error) {
	fullPath := path.Join("/", c.client.AccountName, "policies")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list policies")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	var result []*Policy
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list policies response")
	}

	return result, nil
}

type GetPolicyInput struct {
	PolicyID string
}

func (c *PoliciesClient) Get(ctx context.Context, input *GetPolicyInput) (*Policy, error) {
	fullPath := path.Join("/", c.client.AccountName, "policies", input.PolicyID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get policy")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	var result *Policy
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get policy response")
	}

	return result, nil
}

type DeletePolicyInput struct {
	PolicyID string
}

func (c *PoliciesClient) Delete(ctx context.Context, input *DeletePolicyInput) error {
	fullPath := path.Join("/", c.client.AccountName, "policies", input.PolicyID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return errors.Wrap(err, "unable to delete policy")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	return nil
}

// UpdatePolicyInput represents the options that can be specified
// when updating a policy. Anything but ID can be modified.
type UpdatePolicyInput struct {
	PolicyID    string   `json:"id"`
	Name        string   `json:"name,omitempty"`
	Rules       []string `json:"rules,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (c *PoliciesClient) Update(ctx context.Context, input *UpdatePolicyInput) (*Policy, error) {
	fullPath := path.Join("/", c.client.AccountName, "policies", input.PolicyID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update policy")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	var result *Policy
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode update policy response")
	}

	return result, nil
}

type CreatePolicyInput struct {
	Name        string   `json:"name"`
	Rules       []string `json:"rules"`
	Description string   `json:"description,omitempty"`
}

func (c *PoliciesClient) Create(ctx context.Context, input *CreatePolicyInput) (*Policy, error) {
	fullPath := path.Join("/", c.client.AccountName, "policies")
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create policy")
	}
	if respReader != nil {
		defer respReader.Close()
	}

	var result *Policy
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode create policy response")
	}

	return result, nil
}
