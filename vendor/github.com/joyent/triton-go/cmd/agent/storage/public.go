//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package storage

import (
	"context"

	"github.com/joyent/triton-go/cmd/config"
	tsc "github.com/joyent/triton-go/storage"
	"github.com/pkg/errors"
)

type AgentStorageClient struct {
	client *tsc.StorageClient
}

func NewStorageClient(cfg *config.TritonClientConfig) (*AgentStorageClient, error) {
	storageClient, err := tsc.NewClient(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Error Creating Triton Storage Client")
	}
	return &AgentStorageClient{
		client: storageClient,
	}, nil
}

func (c *AgentStorageClient) GetDirectoryListing(args []string) (*tsc.ListDirectoryOutput, error) {

	input := &tsc.ListDirectoryInput{}
	if len(args) > 0 {
		input.DirectoryName = args[0]
	}

	directoryOutput, err := c.client.Dir().List(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return directoryOutput, nil
}
