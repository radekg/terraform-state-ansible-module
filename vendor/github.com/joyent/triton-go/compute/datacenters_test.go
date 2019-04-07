//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/abdullin/seq"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
)

const dataCenterName = "us-east-1"

// Note that this is specific to Joyent Public Cloud and will not pass on
// private installations of Triton.
func TestAccDataCenters_Get(t *testing.T) {
	const dataCenterName = "us-east-1"
	const dataCenterURL = "https://us-east-1.api.joyentcloud.com"

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "datacenter",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "datacenter",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)
					ctx := context.Background()
					input := &compute.GetDataCenterInput{
						Name: dataCenterName,
					}
					return c.Datacenters().Get(ctx, input)
				},
			},

			&testutils.StepAssert{
				StateBagKey: "datacenter",
				Assertions: seq.Map{
					"name": dataCenterName,
					"url":  dataCenterURL,
				},
			},
		},
	})
}

// Note that this is specific to Joyent Public Cloud and will not pass on
// private installations of Triton.
func TestAccDataCenters_List(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "datacenter",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "datacenters",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)
					ctx := context.Background()
					input := &compute.ListDataCentersInput{}
					return c.Datacenters().List(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					dcs, ok := state.GetOk("datacenters")
					if !ok {
						return fmt.Errorf("State key %q not found", "datacenters")
					}

					toFind := []string{"us-east-1", "eu-ams-1"}
					for _, dcName := range toFind {
						found := false
						for _, dc := range dcs.([]*compute.DataCenter) {
							if dc.Name == dcName {
								found = true
								if dc.URL == "" {
									return fmt.Errorf("%q has no URL", dc.Name)
								}
							}
						}
						if !found {
							return fmt.Errorf("Did not find DC %q", dcName)
						}
					}

					return nil
				},
			},
		},
	})
}

func TestListDataCenters(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.DataCenter, error) {
		defer testutils.DeactivateClient()

		dcs, err := cc.Datacenters().List(ctx, &compute.ListDataCentersInput{})
		if err != nil {
			return nil, err
		}

		return dcs, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), listDataCentersSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), listDataCentersEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), listDataCentersBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), listDataCentersError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list datacenters") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetDataCenter(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.DataCenter, error) {
		defer testutils.DeactivateClient()

		dc, err := cc.Datacenters().Get(ctx, &compute.GetDataCenterInput{
			Name: dataCenterName,
		})
		if err != nil {
			return nil, err
		}

		return dc, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), listDataCentersSuccess)
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), getDataCenterSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}

		if resp.URL != "https://us-east-1.api.joyentcloud.com" {
			t.Fatal("Expected URL to be `https://us-east-1.api.joyentcloud.com` but got `https://us-east-1.api.joyentcloud.com`", resp.URL)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), getDataCenterError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), fmt.Sprintf("datacenter 'not-a-real-dc-name' not found")) {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func listDataCentersSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"us-east-1": "https://us-east-1.api.joyentcloud.com",
	"us-west-1": "https://us-west-1.api.joyentcloud.com",
	"us-sw-1": "https://us-sw-1.api.joyentcloud.com",
	"eu-ams-1": "https://eu-ams-1.api.joyentcloud.com",
	"us-east-2": "https://us-east-2.api.joyentcloud.com",
	"us-east-3": "https://us-east-3.api.joyentcloud.com"
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listDataCentersEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listDataCentersBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"us-east-1": "https://us-east-1.api.joyentcloud.com",
	"us-west-1": "https://us-west-1.api.joyentcloud.com",
	"us-sw-1": "https://us-sw-1.api.joyentcloud.com",
	"eu-ams-1": "https://eu-ams-1.api.joyentcloud.com",
	"us-east-2": "https://us-east-2.api.joyentcloud.com",
	"us-east-3": "https://us-east-3.api.joyentcloud.com",
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listDataCentersError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list datacenters")
}

func getDataCenterSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"us-east-1": "https://us-east-1.api.joyentcloud.com"
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getDataCenterError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("datacenter 'not-a-real-dc-name' not found")
}
