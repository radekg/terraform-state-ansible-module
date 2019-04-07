//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package account_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/account"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

var (
	getConfigErrorType    = errors.New("unable to get account config")
	updateConfigErrorType = errors.New("unable to update account config")
)

func TestAccConfig_Get(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "config",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return account.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "config",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.GetConfigInput{}
					return c.Config().Get(ctx, input)
				},
			},
			&testutils.StepAssertSet{
				StateBagKey: "config",
				Keys:        []string{"DefaultNetwork"},
			},
		},
	})
}

func TestGetConfiguration(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Config, error) {
		defer testutils.DeactivateClient()

		config, err := ac.Config().Get(ctx, &account.GetConfigInput{})
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "config"), getConfigSuccess)

		resp, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "config"), getConfigEmpty)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "config"), getConfigBadeDecode)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", "testAccount", "config"), getConfigError)

		resp, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get account config") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestUpdateConfig(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Config, error) {
		defer testutils.DeactivateClient()

		config, err := ac.Config().Update(ctx, &account.UpdateConfigInput{
			DefaultNetwork: "c00cbe98-7dea-44d3-b644-5bd078700bf8",
		})
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "config"), updateConfigSuccess)

		_, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "config"), updateConfigError)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update account") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func getConfigSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "default_network": "45607081-4cd2-45c8-baf7-79da760fffaa"
}
`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getConfigError(req *http.Request) (*http.Response, error) {
	return nil, getConfigErrorType
}

func getConfigBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "default_network": "45607081-4cd2-45c8-baf7-79da760fffaa",
}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getConfigEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func updateConfigSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "default_network": "c00cbe98-7dea-44d3-b644-5bd078700bf8"
}
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateConfigError(req *http.Request) (*http.Response, error) {
	return nil, updateConfigErrorType
}
