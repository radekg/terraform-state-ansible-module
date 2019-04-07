//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package storage_test

import (
	"context"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/storage"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

const accountUrl = "testing"

var (
	putSnapLinkErrorType = errors.New("unable to put snaplink")
	linkPath             = "/stor/foobar.json"
	brokenLinkPath       = "/missingfolder/foo.json"
	sourcePath           = "/stor/foo.json"
)

func MockStorageClient() *storage.StorageClient {
	return &storage.StorageClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountUrl,
		}),
	}
}

func TestPutSnaplink(t *testing.T) {
	storageClient := MockStorageClient()

	do := func(ctx context.Context, sc *storage.StorageClient) error {
		defer testutils.DeactivateClient()

		return sc.SnapLinks().Put(ctx, &storage.PutSnapLinkInput{
			LinkPath:   linkPath,
			SourcePath: sourcePath,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, linkPath), putSnapLinkSuccess)

		err := do(context.Background(), storageClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, brokenLinkPath), putSnapLinkError)

		err := do(context.Background(), storageClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to put snaplink") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func putSnapLinkSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func putSnapLinkError(req *http.Request) (*http.Response, error) {
	return nil, putSnapLinkErrorType
}
