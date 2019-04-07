//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/joyent/triton-go/client"
	"github.com/joyent/triton-go/compute"
	pkgerrors "github.com/pkg/errors"
)

const groupsPath = "/v1/tsg/groups"

type GroupsClient struct {
	client *client.Client
}

type ServiceGroup struct {
	ID         string    `json:"id"`
	GroupName  string    `json:"group_name"`
	TemplateID string    `json:"template_id"`
	Capacity   int       `json:"capacity"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListGroupsInput struct{}

func (c *GroupsClient) List(ctx context.Context, _ *ListGroupsInput) ([]*ServiceGroup, error) {
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   groupsPath,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to list groups")
	}

	var results []*ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode list groups response")
	}

	return results, nil
}

// _Instance is a private facade over Instance that handles the necessary API
// overrides from VMAPI's machine endpoint(s).
type _Instance struct {
	compute.Instance
	Tags map[string]interface{} `json:"tags"`
}

// toNative() exports a given _Instance (API representation) to its native object
// format.
func (api *_Instance) toNative() (*compute.Instance, error) {
	m := compute.Instance(api.Instance)
	m.CNS, m.Tags = compute.TagsExtractMeta(api.Tags)
	return &m, nil
}

type ListGroupInstancesInput struct {
	ID string
}

func (i *ListGroupInstancesInput) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("group id can not be empty")
	}

	return nil
}

func (c *GroupsClient) ListInstances(ctx context.Context, input *ListGroupInstancesInput) ([]*compute.Instance, error) {
	if err := input.Validate(); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to validate list group instances input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path.Join(groupsPath, input.ID, "instances"),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to list group instances")
	}

	var results []*_Instance
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode group instance response")
	}

	machines := make([]*compute.Instance, 0, len(results))
	for _, machineAPI := range results {
		native, err := machineAPI.toNative()
		if err != nil {
			return nil, pkgerrors.Wrap(err, "unable to decode group instances response")
		}
		machines = append(machines, native)
	}

	return machines, nil
}

type GetGroupInput struct {
	ID string
}

func (i *GetGroupInput) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("group id can not be empty")
	}

	return nil
}

func (c *GroupsClient) Get(ctx context.Context, input *GetGroupInput) (*ServiceGroup, error) {
	if err := input.Validate(); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to validate get group input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path.Join(groupsPath, input.ID),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get service group")
	}

	var results *ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get group response")
	}

	return results, nil
}

type CreateGroupInput struct {
	GroupName  string `json:"group_name"`
	TemplateID string `json:"template_id"`
	Capacity   int    `json:"capacity"`
}

func (input *CreateGroupInput) toAPI() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if input.GroupName != "" {
		result["group_name"] = input.GroupName
	}

	if input.TemplateID == "" {
		return nil, fmt.Errorf("unable to create service group without template ID")
	}
	result["template_id"] = input.TemplateID

	result["capacity"] = input.Capacity

	return result, nil
}

func (c *GroupsClient) Create(ctx context.Context, input *CreateGroupInput) (*ServiceGroup, error) {
	body, err := input.toAPI()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to validate create group input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   groupsPath,
		Body:   body,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to create group")
	}

	var results *ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get group response")
	}

	return results, nil
}

type UpdateGroupInput struct {
	ID         string `json:"id"`
	GroupName  string `json:"group_name"`
	TemplateID string `json:"template_id"`
	Capacity   int    `json:"capacity"`
}

func (input *UpdateGroupInput) updateToAPI() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if input.ID != "" {
		result["id"] = input.ID
	}

	if input.GroupName != "" {
		result["group_name"] = input.GroupName
	}

	if input.TemplateID != "" {
		result["template_id"] = input.TemplateID
	}

	if input.Capacity != 0 {
		result["capacity"] = input.Capacity
	}

	return result, nil
}

func (c *GroupsClient) Update(ctx context.Context, input *UpdateGroupInput) (*ServiceGroup, error) {
	body, err := input.updateToAPI()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to validate create group input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodPut,
		Path:   path.Join(groupsPath, input.ID),
		Body:   body,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to create group")
	}

	var results *ServiceGroup
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get group response")
	}

	return results, nil
}

type DeleteGroupInput struct {
	ID string
}

func (i *DeleteGroupInput) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("group id can not be empty")
	}

	return nil
}

func (c *GroupsClient) Delete(ctx context.Context, input *DeleteGroupInput) error {
	if err := input.Validate(); err != nil {
		return pkgerrors.Wrap(err, "unable to validate delete group input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path.Join(groupsPath, input.ID),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return pkgerrors.Wrap(err, "unable to delete group")
	}

	return nil
}
