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

const accountUrl = "testing"

var (
	getAccountErrorType    = errors.New("unable to get account details")
	updateAccountErrorType = errors.New("unable to update account details")
)

func MockAccountClient() *account.AccountClient {
	return &account.AccountClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountUrl,
		}),
	}
}

func TestAccAccount_Get(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "account",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return account.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "account",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.GetInput{}
					return c.Get(ctx, input)
				},
			},

			&testutils.StepAssertSet{
				StateBagKey: "account",
				Keys:        []string{"ID", "Login", "Email"},
			},
		},
	})
}

func TestGetAccount(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Account, error) {
		defer testutils.DeactivateClient()

		user, err := ac.Get(ctx, &account.GetInput{})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl), getAccountSuccess)

		resp, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl), getAccountEmpty)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl), getAccountBadeDecode)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", "testAccount"), getAccountError)

		resp, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get account details") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestUpdateAccount(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Account, error) {
		defer testutils.DeactivateClient()

		user, err := ac.Update(ctx, &account.UpdateInput{
			Phone: "1 (234) 567 890",
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl), updateAccountSuccess)

		_, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl), updateAccountError)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update account") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func getAccountSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "b89d9dd3-62ce-4f6f-eb0d-f78e57d515d9",
  "login": "barbar",
  "email": "barbar@example.com",
  "companyName": "Example Inc",
  "firstName": "BarBar",
  "lastName": "Jinks",
  "phone": "123-456-7890",
  "updated": "2015-12-21T11:48:54.884Z",
  "created": "2015-12-21T11:48:54.884Z"
}
`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getAccountError(req *http.Request) (*http.Response, error) {
	return nil, getAccountErrorType
}

func getAccountBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "b89d9dd3-62ce-4f6f-eb0d-f78e57d515d9",
  "login": "barbar",
  "email": "barbar@example.com",
  "companyName": "Example Inc",
  "firstName": "BarBar",
  "lastName": "Jinks",
  "phone": "123-456-7890",
  "updated": "2015-12-21T11:48:54.884Z",
  "created": "2015-12-21T11:48:54.884Z",
}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getAccountEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func updateAccountSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "123-3456-2335",
    "login": "testuser",
    "email": "barbar@example.com",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z"
  }
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateAccountError(req *http.Request) (*http.Response, error) {
	return nil, updateAccountErrorType
}
