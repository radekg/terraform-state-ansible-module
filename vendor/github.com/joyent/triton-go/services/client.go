//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package services

import (
	"net/http"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type ServiceGroupClient struct {
	Client *client.Client
}

func newServiceGroupClient(client *client.Client) *ServiceGroupClient {
	return &ServiceGroupClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Service Groups endpoints and
// resources within TSG
func NewClient(config *triton.ClientConfig) (*ServiceGroupClient, error) {
	// TODO: Utilize config interface within the function itself
	client, err := client.New(
		config.TritonURL,
		config.MantaURL,
		config.AccountName,
		config.Signers...,
	)
	if err != nil {
		return nil, err
	}
	return newServiceGroupClient(client), nil
}

// SetHeaders allows a consumer of the current client to set custom headers for
// the next backend HTTP request sent to CloudAPI
func (c *ServiceGroupClient) SetHeader(header *http.Header) {
	c.Client.RequestHeader = header
}

// Templates returns a TemplatesClient used for accessing functions pertaining
// to Instance Templates functionality in the TSG API.
func (c *ServiceGroupClient) Templates() *TemplatesClient {
	return &TemplatesClient{c.Client}
}

// Groups returns a GroupsClient used for accessing functions pertaining
// to Service Groups functionality in the TSG API.
func (c *ServiceGroupClient) Groups() *GroupsClient {
	return &GroupsClient{c.Client}
}
