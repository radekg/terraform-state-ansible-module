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
	pkgerrors "github.com/pkg/errors"
)

const templatesPath = "/v1/tsg/templates"

type TemplatesClient struct {
	client *client.Client
}

type InstanceTemplate struct {
	ID              string            `json:"id"`
	TemplateName    string            `json:"template_name"`
	Package         string            `json:"package"`
	ImageID         string            `json:"image_id"`
	FirewallEnabled bool              `json:"firewall_enabled"`
	Networks        []string          `json:"networks"`
	Userdata        string            `json:"userdata"`
	Metadata        map[string]string `json:"metadata"`
	Tags            map[string]string `json:"tags"`
	CreatedAt       time.Time         `json:"created_at"`
}

type ListTemplatesInput struct{}

func (c *TemplatesClient) List(ctx context.Context, _ *ListTemplatesInput) ([]*InstanceTemplate, error) {
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   templatesPath,
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to list templates")
	}

	var results []*InstanceTemplate
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode list templates response")
	}

	return results, nil
}

type GetTemplateInput struct {
	ID string
}

func (i *GetTemplateInput) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("template id can not be empty")
	}

	return nil
}

func (c *TemplatesClient) Get(ctx context.Context, input *GetTemplateInput) (*InstanceTemplate, error) {
	if err := input.Validate(); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get instance template")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path.Join(templatesPath, input.ID),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to get template")
	}

	var results *InstanceTemplate
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&results); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode get template response")
	}

	return results, nil
}

type CreateTemplateInput struct {
	TemplateName    string            `json:"template_name"`
	Package         string            `json:"package"`
	ImageID         string            `json:"image_id"`
	FirewallEnabled bool              `json:"firewall_enabled"`
	Networks        []string          `json:"networks"`
	Userdata        string            `json:"userdata"`
	Metadata        map[string]string `json:"metadata"`
	Tags            map[string]string `json:"tags"`
}

func (input *CreateTemplateInput) toAPI() map[string]interface{} {
	result := make(map[string]interface{})

	if input.TemplateName != "" {
		result["template_name"] = input.TemplateName
	}

	if input.Package != "" {
		result["package"] = input.Package
	}

	if input.ImageID != "" {
		result["image_id"] = input.ImageID
	}

	result["firewall_enabled"] = input.FirewallEnabled

	if len(input.Networks) > 0 {
		result["networks"] = input.Networks
	}

	if input.Userdata != "" {
		result["userdata"] = input.Userdata
	}

	if len(input.Metadata) > 0 {
		result["metadata"] = input.Metadata
	}

	if len(input.Tags) > 0 {
		result["tags"] = input.Tags
	}

	return result
}

func (c *TemplatesClient) Create(ctx context.Context, input *CreateTemplateInput) (*InstanceTemplate, error) {
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   templatesPath,
		Body:   input.toAPI(),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, pkgerrors.Wrap(err, "unable to create template")
	}

	var result *InstanceTemplate
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, pkgerrors.Wrap(err, "unable to decode create template response")
	}

	return result, nil
}

type DeleteTemplateInput struct {
	ID string
}

func (i *DeleteTemplateInput) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("template name can not be empty")
	}

	return nil
}

func (c *TemplatesClient) Delete(ctx context.Context, input *DeleteTemplateInput) error {
	if err := input.Validate(); err != nil {
		return pkgerrors.Wrap(err, "unable to validate delete template input")
	}

	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path.Join(templatesPath, input.ID),
	}
	respReader, err := c.client.ExecuteRequestTSG(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return pkgerrors.Wrap(err, "unable to delete template")
	}

	return nil
}
