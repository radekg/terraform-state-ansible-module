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
	"strings"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
)

type RolesClient struct {
	client *client.Client
}

type Role struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Policies       []string `json:"policies"`
	Members        []string `json:"members"`
	DefaultMembers []string `json:"default_members"`
}

type ListRolesInput struct{}

func (c *RolesClient) List(ctx context.Context, _ *ListRolesInput) ([]*Role, error) {
	fullPath := path.Join("/", c.client.AccountName, "roles")

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list roles")
	}

	var result []*Role
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list roles response")
	}

	return result, nil
}

type GetRoleInput struct {
	RoleID string
}

func (c *RolesClient) Get(ctx context.Context, input *GetRoleInput) (*Role, error) {
	fullPath := path.Join("/", c.client.AccountName, "roles", input.RoleID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get role")
	}

	var result *Role
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get roles response")
	}

	return result, nil
}

// CreateRoleInput represents the options that can be specified
// when creating a new role.
type CreateRoleInput struct {
	// Name of the role. Required.
	Name string `json:"name"`

	// This account's policies to be given to this role. Optional.
	Policies []string `json:"policies,omitempty"`

	// This account's user logins to be added to this role. Optional.
	Members []string `json:"members,omitempty"`

	// This account's user logins to be added to this role and have
	// it enabled by default. Optional.
	DefaultMembers []string `json:"default_members,omitempty"`
}

func (c *RolesClient) Create(ctx context.Context, input *CreateRoleInput) (*Role, error) {
	fullPath := path.Join("/", c.client.AccountName, "roles")
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to create role")

	}

	var result *Role
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode create role response")
	}

	return result, nil
}

// UpdateRoleInput represents the options that can be specified
// when updating a role. Anything but ID can be modified.
type UpdateRoleInput struct {
	// ID of the role to modify. Required.
	RoleID string `json:"id"`

	// Name of the role. Required.
	Name string `json:"name"`

	// This account's policies to be given to this role. Optional.
	Policies []string `json:"policies,omitempty"`

	// This account's user logins to be added to this role. Optional.
	Members []string `json:"members,omitempty"`

	// This account's user logins to be added to this role and have
	// it enabled by default. Optional.
	DefaultMembers []string `json:"default_members,omitempty"`
}

func (c *RolesClient) Update(ctx context.Context, input *UpdateRoleInput) (*Role, error) {
	fullPath := path.Join("/", c.client.AccountName, "roles", input.RoleID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to update role")
	}

	var result *Role
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode update role response")
	}

	return result, nil
}

type DeleteRoleInput struct {
	RoleID string
}

func (c *RolesClient) Delete(ctx context.Context, input *DeleteRoleInput) error {
	fullPath := path.Join("/", c.client.AccountName, "roles", input.RoleID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errors.Wrap(err, "unable to delete role")
	}

	return nil
}

// RoleTags represents a set of results after an operation
// of either querying or setting role tags.
type RoleTags struct {
	// Path to the resource.
	Name string `json:"name"`
	// The list of role tags assigned to this resource.
	RoleTags []string `json:"role-tag"`
}

// SetRoleTagsInput represents the options that can be
// specified when setting a given list of role tags for
// either a whole group of resources or an individual
// resource of the same type.
type SetRoleTagsInput struct {
	// A type of the resource. Required.
	ResourceType string `json:"-"`
	// An unique identifier of an individual resource.
	// Optional.
	ResourceID string `json:"-"`
	// The list of role tags assigned to this resource.
	// Required.
	RoleTags []string `json:"role-tag"`
}

// SetRoleTags sets a given list of role tags for either a whole group of
// resources or an individual resource of the same type. To set a given
// list of role tags for an individual resource an unique identifier (ID)
// of this resource has to be provided in an addition to the resource type.
// Returns an object with a set of results, or an error otherwise.
func (c *RolesClient) SetRoleTags(ctx context.Context, input *SetRoleTagsInput) (*RoleTags, error) {
	fullPath := path.Join("/", c.client.AccountName, input.ResourceType, input.ResourceID)
	reqInputs := client.RequestInput{
		Method: http.MethodPut,
		Path:   fullPath,
		Body:   input,
	}

	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set role tags")
	}
	defer respReader.Close()

	var result *RoleTags
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode set role tags response")
	}

	return result, nil
}

// GetRoleTagsInput represents the options that can be
// specified when querying for a list of currently set
// role tags for either a whole group of resources or
// an individual resource of the same type.
type GetRoleTagsInput struct {
	// A type of the resource. Required.
	ResourceType string
	// An unique identifier of an individual resource.
	// Optional.
	ResourceID string
}

// GetRoleTags retrieves a list of currently set role tags for either a whole
// group of resources or an individual resource of the same type. To retrieve
// a list of role tags currently set for an individual resource an unique
// identifier (ID) of this resource has to be provided in an additon to the
// resource type. Returns an object with a set of results, or an error
// otherwise.
func (c *RolesClient) GetRoleTags(ctx context.Context, input *GetRoleTagsInput) (*RoleTags, error) {
	fullPath := path.Join("/", c.client.AccountName, input.ResourceType, input.ResourceID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}

	response, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get role tags")
	}
	defer response.Body.Close()

	roleTags := response.Header.Get("Role-Tag")

	result := &RoleTags{
		Name:     fullPath,
		RoleTags: strings.Split(roleTags, ","),
	}

	return result, nil
}
