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

	"github.com/abdullin/seq"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/account"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

const testAccCreateKeyMaterial = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBOJ5z6jTdY3SYK2Nc+MQLSQstAOzxFqDN00MJ9SMhJea8ZQbZFlhCAZBFE4TUBDI3zXBxFjygh84lb1QlNu1dmZeoQ10MThuowZllBAfg9Eb5RkXqLvDdYh9+rLdEdUL4+aiYZ8JYtQ+K5ZnogZoxdzNQ3WnVhMGJIrj1zcRveUSvQ6tMhaEQDxDWrAMDLxnLI/6SNmkhdF1ZKE8iQ+BnazYp0vg5jAzkHzEYJY9kFUOubupOxio93B9OTkpQ0jZD+J9iR1t8Me3JdhHy85inaAFc0fkjznDYluV8aqfIprD/WE9grQ/GfEYfsvQdQr1ljLBJZdad7DvnKqU0M4vJ James@jn-mpb15`
const testAccCreateKeyFingerprint = `ab:f4:8f:bc:26:e1:cf:1d:06:a3:9d:40:39:7c:5a:78`

var (
	listKeysErrorType  = errors.New("unable to list keys")
	getKeyErrorType    = errors.New("unable to get key")
	deleteKeyErrorType = errors.New("unable to delete key")
	createKeyErrorType = errors.New("unable to create key")
)

func TestAccKey_Create(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccCreateKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return account.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},
			&testutils.StepAssert{
				StateBagKey: "key",
				Assertions: seq.Map{
					"name":        keyName,
					"key":         testAccCreateKeyMaterial,
					"fingerprint": testAccCreateKeyFingerprint,
				},
			},
		},
	})
}

func TestAccKey_Get(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccGetKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return account.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "getKey",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.GetKeyInput{
						KeyName: keyName,
					}
					return c.Keys().Get(ctx, input)
				},
			},

			&testutils.StepAssert{
				StateBagKey: "getKey",
				Assertions: seq.Map{
					"name":        keyName,
					"key":         testAccCreateKeyMaterial,
					"fingerprint": testAccCreateKeyFingerprint,
				},
			},
		},
	})
}

func TestAccKey_Delete(t *testing.T) {
	keyName := testutils.RandPrefixString("TestAccGetKey", 32)

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "key",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return account.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "key",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.CreateKeyInput{
						Name: keyName,
						Key:  testAccCreateKeyMaterial,
					}
					return c.Keys().Create(ctx, input)
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.DeleteKeyInput{
						KeyName: keyName,
					}
					c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "noop",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.DeleteKeyInput{
						KeyName: keyName,
					}
					return nil, c.Keys().Delete(ctx, input)
				},
			},

			&testutils.StepAPICall{
				ErrorKey: "getKeyError",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*account.AccountClient)
					ctx := context.Background()
					input := &account.GetKeyInput{
						KeyName: keyName,
					}
					return c.Keys().Get(ctx, input)
				},
			},

			&testutils.StepAssertTritonError{
				ErrorKey: "getKeyError",
				Code:     "ResourceNotFound",
			},
		},
	})
}

func TestGetKey(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Key, error) {
		defer testutils.DeactivateClient()

		user, err := ac.Keys().Get(ctx, &account.GetKeyInput{
			KeyName: "my-test-key",
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys", "my-test-key"), getKeySuccess)

		resp, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys", "my-test-key"), getKeyEmpty)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys", "my-test-key"), getKeyBadeDecode)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys"), getKeyError)

		resp, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get key") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteKey(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) error {
		defer testutils.DeactivateClient()

		return ac.Keys().Delete(ctx, &account.DeleteKeyInput{
			KeyName: "my-test-key",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountUrl, "keys", "my-test-key"), deleteKeySuccess)

		err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountUrl, "keys", "my-test-key"), deleteKeyError)

		err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete key") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCreateUser(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) (*account.Key, error) {
		defer testutils.DeactivateClient()

		key, err := ac.Keys().Create(ctx, &account.CreateKeyInput{
			Key:  testAccCreateKeyMaterial,
			Name: "my-test-key",
		})

		if err != nil {
			return nil, err
		}
		return key, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "keys"), createKeySuccess)

		_, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "keys"), createKeyError)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create key") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListKeys(t *testing.T) {
	accountClient := MockAccountClient()

	do := func(ctx context.Context, ac *account.AccountClient) ([]*account.Key, error) {
		defer testutils.DeactivateClient()

		keys, err := ac.Keys().List(ctx, &account.ListKeysInput{})
		if err != nil {
			return nil, err
		}
		return keys, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys"), listKeysSuccess)

		resp, err := do(context.Background(), accountClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys"), listKeysEmpty)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys"), listKeysBadeDecode)

		_, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "keys"), listKeysError)

		resp, err := do(context.Background(), accountClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list keys") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func getKeySuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
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

func getKeyError(req *http.Request) (*http.Response, error) {
	return nil, getKeyErrorType
}

func getKeyBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "id": "4fc13ac6-1e7d-cd79-f3d2-96276af0d638",
    "login": "barbar",
    "email": "barbar@example.com",
    "companyName": "Example",
    "firstName": "BarBar",
    "lastName": "Jinks",
    "phone": "(123)457-6890",
    "updated": "2015-12-23T06:41:11.032Z",
    "created": "2015-12-23T06:41:11.032Z",}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getKeyEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func deleteKeySuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 204,
		Header:     header,
	}, nil
}

func deleteKeyError(req *http.Request) (*http.Response, error) {
	return nil, deleteKeyErrorType
}

func createKeySuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "my-test-key",
  "fingerprint": "03:7f:8e:ef:da:3d:3b:9e:a4:82:67:71:8c:35:2c:aa",
  "key": "<...>"
}
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createKeyError(req *http.Request) (*http.Response, error) {
	return nil, createKeyErrorType
}

func listKeysEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listKeysSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "name": "barbar",
    "fingerprint": "03:7f:8e:ef:da:3d:3b:9e:a4:82:67:71:8c:35:2c:aa",
    "key": "<...>"
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listKeysBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{[
	{
    "name": "barbar",
    "fingerprint": "03:7f:8e:ef:da:3d:3b:9e:a4:82:67:71:8c:35:2c:aa",
    "key": "<...>"
  }
]}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listKeysError(req *http.Request) (*http.Response, error) {
	return nil, listKeysErrorType
}
