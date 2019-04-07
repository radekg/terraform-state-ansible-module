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
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/abdullin/seq"
	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
)

var (
	fakePackageId   = "7b17343c-94af-6266-e0e8-893a3b9993d0"
	fakePackageName = "g4-test"
)

func TestAccPackagesList(t *testing.T) {
	const stateKey = "packages"
	const package1ID = "14ad9d54-d0f8-11e5-a759-93bdb33c9583"
	const package2ID = "14af2214-d0f8-11e5-9399-77e0d621f66d"

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: stateKey,
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: stateKey,
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)
					ctx := context.Background()
					input := &compute.ListPackagesInput{}
					return c.Packages().List(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					packages, ok := state.GetOk(stateKey)
					if !ok {
						return fmt.Errorf("state key %q not found", stateKey)
					}

					toFind := []string{package1ID, package2ID}
					for _, pkgID := range toFind {
						found := false
						for _, pkg := range packages.([]*compute.Package) {
							if pkg.ID == pkgID {
								found = true
								state.Put(pkg.ID, pkg)
							}
						}
						if !found {
							return fmt.Errorf("couldn't find package %q", pkgID)
						}
					}

					return nil
				},
			},

			&testutils.StepAssert{
				StateBagKey: package1ID,
				Assertions: seq.Map{
					"name":        "g4-highcpu-128M",
					"memory":      128,
					"disk":        3072,
					"swap":        512,
					"vcpus":       0,
					"lwps":        4000,
					"default":     false,
					"id":          "14ad9d54-d0f8-11e5-a759-93bdb33c9583",
					"version":     "1.0.3",
					"description": "Compute Optimized 128M RAM - 0.0625 vCPU - 3 GB Disk",
					"group":       "Compute Optimized",
				},
			},

			&testutils.StepAssert{
				StateBagKey: package2ID,
				Assertions: seq.Map{
					"name":        "g4-highcpu-1G",
					"memory":      1024,
					"disk":        25600,
					"swap":        4096,
					"vcpus":       0,
					"lwps":        4000,
					"default":     true,
					"id":          "14af2214-d0f8-11e5-9399-77e0d621f66d",
					"version":     "1.0.3",
					"description": "Compute Optimized 1G RAM - 0.5 vCPU - 25 GB Disk",
					"group":       "Compute Optimized",
				},
			},
		},
	})
}

func TestListPackages(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Package, error) {
		defer testutils.DeactivateClient()

		packages, err := cc.Packages().List(ctx, &compute.ListPackagesInput{})
		if err != nil {
			return nil, err
		}
		return packages, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages"), listPackagesSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("filtered", func(t *testing.T) {
		v := url.Values{}
		v.Set("name", fakePackageName)

		filterURL := path.Join("/", accountURL, "packages") + "?" + v.Encode()
		testutils.RegisterResponder("GET", filterURL, listPackagesFiltered)
		defer testutils.DeactivateClient()

		ctx := context.Background()
		cc := computeClient
		packages, err := cc.Packages().List(ctx, &compute.ListPackagesInput{
			Name: fakePackageName,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(packages) != 1 {
			t.Fatalf("expected output but received empty body")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages"), listPackagesEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages"), listPackagesBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages"), listPackagesError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list packages") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetPackage(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Package, error) {
		defer testutils.DeactivateClient()

		pkg, err := cc.Packages().Get(ctx, &compute.GetPackageInput{
			ID: fakePackageId,
		})
		if err != nil {
			return nil, err
		}
		return pkg, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages", fakePackageId), getPackageSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages", fakePackageId), getPackageEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages", fakePackageId), getPackageBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "packages", "not-a-real-package-id"), getPackageError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get package") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func listPackagesFiltered(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
	"id": "7b17343c-94af-6266-e0e8-893a3b9993d0",
	"name": "g4-test",
	"memory": 1024,
	"disk": 25600,
	"swap": 4096,
	"vcpus": 0,
	"lwps": 4000,
	"default": false,
	"version": "1.0.0"
  }]
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listPackagesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
	"id": "7b17343c-94af-6266-e0e8-893a3b9993d0",
	"name": "sdc_128",
	"memory": 128,
	"disk": 12288,
	"swap": 256,
	"vcpus": 1,
	"lwps": 1000,
	"default": false,
	"version": "1.0.0"
  }]
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listPackagesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listPackagesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
	"id": "7b17343c-94af-6266-e0e8-893a3b9993d0",
	"name": "sdc_128",
	"memory": 128,
	"disk": 12288,
	"swap": 256,
	"vcpus": 1,
	"lwps": 1000,
	"default": false,
	"version": "1.0.0",
  }]`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listPackagesError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list packages")
}

func getPackageSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "7b17343c-94af-6266-e0e8-893a3b9993d0",
  "name": "sdc_128",
  "memory": 128,
  "disk": 12288,
  "swap": 256,
  "vcpus": 1,
  "lwps": 1000,
  "default": false,
  "version": "1.0.0"
}
`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getPackageBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "7b17343c-94af-6266-e0e8-893a3b9993d0",
  "name": "sdc_128",
  "memory": 128,
  "disk": 12288,
  "swap": 256,
  "vcpus": 1,
  "lwps": 1000,
  "default": false,
  "version": "1.0.0",
}`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getPackageEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getPackageError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get package")
}
