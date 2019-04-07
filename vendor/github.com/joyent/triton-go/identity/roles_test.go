//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package identity_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/identity"
	"github.com/joyent/triton-go/testutils"
)

const fakeRoleID = "e53b8fec-e661-4ded-a21e-959c9ba08cb2"

var (
	listRolesErrorType   = errors.New("unable to list roles")
	getRoleErrorType     = errors.New("unable to get role")
	deleteRoleErrorType  = errors.New("unable to delete role")
	createRoleErrorType  = errors.New("unable to create role")
	updateRoleErrorType  = errors.New("unable to update role")
	setRoleTagsErrorType = errors.New("unable to set role tags")
	getRoleTagsErrorType = errors.New("unable to get role tags")
)

func TestAccIdentity_SetRoleTags(t *testing.T) {
	testRoleName := testutils.RandPrefixString("TestAccSetRoleTags", 32)

	// Holds newly created role.
	var newRole *identity.Role

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{
			&testutils.StepClient{
				StateBagKey: "identity",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return identity.NewClient(config)
				},
			},
			&testutils.StepAPICall{
				StateBagKey: "identity",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.CreateRoleInput{
						Name: testRoleName,
					}
					role, err := c.Roles().Create(ctx, input)
					newRole = role

					// Allow a few seconds for the new role
					// to propagate within CloudAPI.
					time.Sleep(5 * time.Second)

					return role, err
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					role, roleOk := callState.(*identity.Role)
					if !roleOk {
						log.Println("Expected state to include role1")
						return
					}

					if role.Name != testRoleName {
						log.Printf("Expected role name to be %q, got %q\n",
							testRoleName, role.Name)
						return
					}

					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.DeleteRoleInput{
						RoleID: role.ID,
					}
					err := c.Roles().Delete(ctx, input)
					if err != nil {
						log.Printf("Could not delete role %q: %v\n", role.ID, err)
					}
				},
			},
			&testutils.StepAPICall{
				StateBagKey: "setRoleTags",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.SetRoleTagsInput{
						ResourceType: "roles",
						ResourceID:   newRole.ID,
						RoleTags: []string{
							newRole.Name,
						},
					}
					return c.Roles().SetRoleTags(ctx, input)
				},
			},
			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					roleTagsRaw, found := state.GetOk("setRoleTags")
					if !found {
						return fmt.Errorf("State key %q not found", "setRoleTags")
					}

					roleTags, roleTagsOk := roleTagsRaw.(*identity.RoleTags)
					if !roleTagsOk {
						return errors.New("Expected state to include role tags")
					}

					actual := roleTags.RoleTags[0]

					if actual != testRoleName {
						return fmt.Errorf("Expected role tags to contain %q, got %q",
							testRoleName, actual)
					}

					return nil

				},
			},
			&testutils.StepAPICall{
				ErrorKey: "setRoleTagsError",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.GetRoleTagsInput{
						ResourceType: "roles",
						ResourceID:   "Bad-Role-Resource-ID",
					}
					return c.Roles().GetRoleTags(ctx, input)
				},
			},
			&testutils.StepAssertTritonError{
				ErrorKey: "setRoleTagsError",
				Code:     "ResourceNotFound",
			},
		},
	})
}

func TestAccIdentity_GetRoleTags(t *testing.T) {
	testRoleName := testutils.RandPrefixString("TestAccGetRoleTags", 32)

	// Holds newly created role.
	var newRole *identity.Role

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{
			&testutils.StepClient{
				StateBagKey: "identity",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return identity.NewClient(config)
				},
			},
			&testutils.StepAPICall{
				StateBagKey: "identity",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.CreateRoleInput{
						Name: testRoleName,
					}
					role, err := c.Roles().Create(ctx, input)
					newRole = role

					// Allow a few seconds for the new role
					// to propagate within CloudAPI.
					time.Sleep(5 * time.Second)

					return role, err
				},
				CleanupFunc: func(client interface{}, callState interface{}) {
					role, roleOk := callState.(*identity.Role)
					if !roleOk {
						log.Println("Expected state to include role1")
						return
					}

					if role.Name != testRoleName {
						log.Printf("Expected role name to be %q, got %q\n",
							testRoleName, role.Name)
						return
					}

					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.DeleteRoleInput{
						RoleID: role.ID,
					}
					err := c.Roles().Delete(ctx, input)
					if err != nil {
						log.Printf("Could not delete role %q: %v\n", role.ID, err)
					}
				},
			},
			&testutils.StepAPICall{
				StateBagKey: "getRoleTags",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					setInput := &identity.SetRoleTagsInput{
						ResourceType: "roles",
						ResourceID:   newRole.ID,
						RoleTags: []string{
							newRole.Name,
						},
					}
					_, err := c.Roles().SetRoleTags(ctx, setInput)
					if err != nil {
						return nil, err
					}

					// Allow a few seconds for the new role tag
					// to propagate within CloudAPI.
					time.Sleep(5 * time.Second)

					getInput := &identity.GetRoleTagsInput{
						ResourceType: "roles",
						ResourceID:   newRole.ID,
					}
					return c.Roles().GetRoleTags(ctx, getInput)
				},
			},
			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					roleTagsRaw, found := state.GetOk("getRoleTags")
					if !found {
						return fmt.Errorf("State key %q not found", "getRoleTags")
					}

					roleTags, roleTagsOk := roleTagsRaw.(*identity.RoleTags)
					if !roleTagsOk {
						return errors.New("Expected state to include role tags")
					}

					actual := roleTags.RoleTags[0]

					if actual != testRoleName {
						return fmt.Errorf("Expected role tags to contain %q, got %q",
							testRoleName, actual)
					}

					return nil

				},
			},
			&testutils.StepAPICall{
				ErrorKey: "getRoleTagsError",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*identity.IdentityClient)
					ctx := context.Background()

					input := &identity.GetRoleTagsInput{
						ResourceType: "roles",
						ResourceID:   "Bad-Role-Resource-ID",
					}
					return c.Roles().GetRoleTags(ctx, input)
				},
			},
			&testutils.StepAssertTritonError{
				ErrorKey: "getRoleTagsError",
				Code:     "ResourceNotFound",
			},
		},
	})
}
func TestDeleteRole(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) error {
		defer testutils.DeactivateClient()

		return ic.Roles().Delete(ctx, &identity.DeleteRoleInput{
			RoleID: fakeRoleID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountUrl, "roles", fakeRoleID), deleteRoleSuccess)

		err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountUrl, "roles"), deleteRoleError)

		err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete role") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCreateRole(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.Role, error) {
		defer testutils.DeactivateClient()

		role, err := ic.Roles().Create(ctx, &identity.CreateRoleInput{
			Name: "readable",
		})

		if err != nil {
			return nil, err
		}
		return role, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "roles"), createRoleSuccess)

		_, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "roles"), createRoleError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create role") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestUpdateRole(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.Role, error) {
		defer testutils.DeactivateClient()

		user, err := ic.Roles().Update(ctx, &identity.UpdateRoleInput{
			RoleID: "e53b8fec-e661-4ded-a21e-959c9ba08cb2",
			Name:   "updated-role-name",
		})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "roles", "e53b8fec-e661-4ded-a21e-959c9ba08cb2"), updateRoleSuccess)

		_, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountUrl, "roles", "e53b8fec-e661-4ded-a21e-959c9ba08cb2"), updateRoleError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update role") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListRoles(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) ([]*identity.Role, error) {
		defer testutils.DeactivateClient()

		roles, err := ic.Roles().List(ctx, &identity.ListRolesInput{})
		if err != nil {
			return nil, err
		}
		return roles, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles"), listRolesSuccess)

		resp, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles"), listRolesEmpty)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles"), listRolesBadeDecode)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles"), listRolesError)

		resp, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list roles") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestGetRole(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.Role, error) {
		defer testutils.DeactivateClient()

		role, err := ic.Roles().Get(ctx, &identity.GetRoleInput{
			RoleID: fakeRoleID,
		})
		if err != nil {
			return nil, err
		}
		return role, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles", fakeRoleID), getRoleSuccess)

		resp, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles", fakeRoleID), getRoleEmpty)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles", fakeRoleID), getRoleBadeDecode)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles"), getRoleError)

		resp, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get role") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestSetRoleTags(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.RoleTags, error) {
		defer testutils.DeactivateClient()

		roleTags, err := ic.Roles().SetRoleTags(ctx, &identity.SetRoleTagsInput{
			ResourceType: "roles",
			ResourceID:   fakeRoleID,
		})
		if err != nil {
			return nil, err
		}
		return roleTags, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, "roles", fakeRoleID), setRoleTagsSuccess)

		response, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		if response == nil {
			t.Error("expected response to be set, got nil")
		}

		actual := response.RoleTags[0]

		if actual != "test-role" {
			t.Errorf("expected role tags to contain test role, got %s", actual)
		}

	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, "roles", fakeRoleID), setRoleTagsError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to set role tags") {
			t.Errorf("expected error to equal test error, got %s", err)
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, "roles", fakeRoleID), setRoleTagsEmpty)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF, got %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountUrl, "roles", fakeRoleID), setRoleTagsBadDecode)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail, got %s", err)
		}
	})
}

func TestGetRoleTags(t *testing.T) {
	identityClient := MockIdentityClient()

	do := func(ctx context.Context, ic *identity.IdentityClient) (*identity.RoleTags, error) {
		defer testutils.DeactivateClient()

		roleTags, err := ic.Roles().GetRoleTags(ctx, &identity.GetRoleTagsInput{
			ResourceType: "roles",
			ResourceID:   fakeRoleID,
		})
		if err != nil {
			return nil, err
		}
		return roleTags, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles", fakeRoleID), getRoleTagsSuccess)

		response, err := do(context.Background(), identityClient)
		if err != nil {
			t.Fatal(err)
		}

		actual := response.RoleTags[0]

		if actual != "test-role" {
			t.Errorf("expected role tags to contain test role, got %s", actual)
		}

	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountUrl, "roles", fakeRoleID), getRoleTagsError)

		_, err := do(context.Background(), identityClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to get role tags") {
			t.Errorf("expected error to equal test error, got %s", err)
		}
	})
}

func deleteRoleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 204,
		Header:     header,
	}, nil
}

func deleteRoleError(req *http.Request) (*http.Response, error) {
	return nil, deleteRoleErrorType
}

func createRoleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "readable",
  "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2"
}
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createRoleError(req *http.Request) (*http.Response, error) {
	return nil, createRoleErrorType
}

func updateRoleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "updated-role-name",
  "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2"
}
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateRoleError(req *http.Request) (*http.Response, error) {
	return nil, updateRoleErrorType
}

func listRolesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listRolesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "name": "readable",
    "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2",
    "members": [
      "foo"
    ],
    "default_members": [
      "foo"
    ],
    "policies": [
      "readinstance"
    ]
  }
]`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRolesBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{[
	{
    "name": "readable",
    "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2",
    "members": [
      "foo"
    ],
    "default_members": [
      "foo"
    ],
    "policies": [
      "readinstance"
    ]
  }
]}`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRolesError(req *http.Request) (*http.Response, error) {
	return nil, listRolesErrorType
}

func getRoleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "name": "readable",
    "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2",
    "members": [
      "foo"
    ],
    "default_members": [
      "foo"
    ],
    "policies": [
      "readinstance"
    ]
  }
`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getRoleError(req *http.Request) (*http.Response, error) {
	return nil, getRoleErrorType
}

func getRoleBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
    "name": "readable",
    "id": "e53b8fec-e661-4ded-a21e-959c9ba08cb2",
    "members": [
      "foo"
    ],
    "default_members": [
      "foo"
    ],
    "policies": [
      "readinstance"
    ],
  }`)
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getRoleEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func setRoleTagsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`
{
  "name": "/testing/role/e53b8fec-e661-4ded-a21e-959c9ba08cb2",
  "role-tag": [
    "test-role"
  ]
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func setRoleTagsError(req *http.Request) (*http.Response, error) {
	return nil, setRoleTagsErrorType
}

func setRoleTagsBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`
{
  "name": "/testing/role/e53b8fec-e661-4ded-a21e-959c9ba08cb2",
  "role-tag": [
    "test-role",
  ]
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func setRoleTagsEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       http.NoBody,
	}, nil
}

func getRoleTagsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("Role-Tag", "test-role")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       http.NoBody,
	}, nil
}

func getRoleTagsError(req *http.Request) (*http.Response, error) {
	return nil, getRoleTagsErrorType
}
