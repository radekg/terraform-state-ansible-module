//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package identity

import (
	"net/http"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type IdentityClient struct {
	Client *client.Client
}

func newIdentityClient(client *client.Client) *IdentityClient {
	return &IdentityClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Identity endpoints and
// resources within CloudAPI
func NewClient(config *triton.ClientConfig) (*IdentityClient, error) {
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
	return newIdentityClient(client), nil
}

// SetHeaders allows a consumer of the current client to set custom headers for
// the next backend HTTP request sent to CloudAPI
func (c *IdentityClient) SetHeader(header *http.Header) {
	c.Client.RequestHeader = header
}

// Roles returns a Roles client used for accessing functions pertaining to
// Role functionality in the Triton API.
func (c *IdentityClient) Roles() *RolesClient {
	return &RolesClient{c.Client}
}

// Users returns a Users client used for accessing functions pertaining to
// User functionality in the Triton API.
func (c *IdentityClient) Users() *UsersClient {
	return &UsersClient{c.Client}
}

// Policies returns a Policies client used for accessing functions pertaining to
// Policy functionality in the Triton API.
func (c *IdentityClient) Policies() *PoliciesClient {
	return &PoliciesClient{c.Client}
}
