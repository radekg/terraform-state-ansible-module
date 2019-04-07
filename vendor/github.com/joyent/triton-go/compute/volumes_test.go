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
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

const fakeVolumeID = "1fd4ecb9-7f66-cf31-9a8a-b661d3adebcf"

func TestGetVolume(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Volume, error) {
		defer testutils.DeactivateClient()

		volume, err := cc.Volumes().Get(ctx, &compute.GetVolumeInput{
			ID: fakeVolumeID,
		})
		if err != nil {
			return nil, err
		}
		return volume, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes", fakeVolumeID), getVolumeSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes", fakeVolumeID), getVolumeEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes", fakeVolumeID), getVolumeBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes", "not-a-real-volume-id"), getVolumeError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get volume") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestDeleteVolume(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Volumes().Delete(ctx, &compute.DeleteVolumeInput{
			ID: fakeVolumeID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "volumes", fakeVolumeID), deleteVolumeSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "volumes", fakeVolumeID), deleteVolumeError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete volume") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestCreateVolume(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Volume, error) {
		defer testutils.DeactivateClient()

		volume, err := cc.Volumes().Create(ctx, &compute.CreateVolumeInput{
			Name:     "my-test-volume-1",
			Type:     "tritonnfs",
			Networks: []string{"d251d640-a02e-47b5-8ae6-8b45d859528e"},
		})
		if err != nil {
			return nil, err
		}
		return volume, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "volumes"), createVolumeSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "volumes"), createVolumeError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create volume") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestUpdateVolume(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Volumes().Update(ctx, &compute.UpdateVolumeInput{
			Name: "my-test-volume-1",
			ID:   fakeVolumeID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "volumes", fakeVolumeID), updateVolumeSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "volumes"), updateVolumeError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update volume") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestListVolumes(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Volume, error) {
		defer testutils.DeactivateClient()

		volumes, err := cc.Volumes().List(ctx, &compute.ListVolumesInput{})
		if err != nil {
			return nil, err
		}
		return volumes, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes"), listVolumesSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes"), listVolumesEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes"), listVolumesBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "volumes"), listVolumesError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list volumes") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func getVolumeSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "my-test-volume-1-updated",
	"owner_uuid": "7530fb4f-aa7a-4f3c-ff9d-be6f2c510d14",
	"size": 10240,
	"type": "tritonnfs",
	"create_timestamp": "2018-01-12T21:17:35.909Z",
	"state": "ready",
	"networks": ["d251d640-a02e-47b5-8ae6-8b45d859528e"],
	"filesystem_path": "192.168.128.9:/exports/data",
	"id": "1fd4ecb9-7f66-cf31-9a8a-b661d3adebcf"
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getVolumeBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "my-test-volume-1-updated",
	"owner_uuid": "7530fb4f-aa7a-4f3c-ff9d-be6f2c510d14",
	"size": 10240,
	"type": "tritonnfs",
	"create_timestamp": "2018-01-12T21:17:35.909Z",
	"state": "ready",
	"networks": ["d251d640-a02e-47b5-8ae6-8b45d859528e"],
	"filesystem_path": "192.168.128.9:/exports/data",
	"id": "1fd4ecb9-7f66-cf31-9a8a-b661d3adebcf",
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getVolumeEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getVolumeError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get volume")
}

func deleteVolumeSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteVolumeError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete volume")
}

func createVolumeSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "my-test-volume-1",
	"owner_uuid": "7530fb4f-aa7a-4f3c-ff9d-be6f2c510d14",
	"size": 10240,
	"type": "tritonnfs",
	"create_timestamp": "2018-01-12T21:09:09.788Z",
	"state": "creating",
	"networks": ["d251d640-a02e-47b5-8ae6-8b45d859528e"],
	"id": "1edcc6ad-7987-4372-b13a-d21e678ba1e9"
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createVolumeError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to create volume")
}

func updateVolumeSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func updateVolumeError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to update volume")
}

func listVolumesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
	"name": "my-test-volume-1",
	"owner_uuid": "7530fb4f-aa7a-4f3c-ff9d-be6f2c510d14",
	"size": 10240,
	"type": "tritonnfs",
	"create_timestamp": "2018-01-12T21:09:09.788Z",
	"state": "creating",
	"networks": ["d251d640-a02e-47b5-8ae6-8b45d859528e"],
	"id": "1edcc6ad-7987-4372-b13a-d21e678ba1e9"
}
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listVolumesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listVolumesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
	"name": "my-test-volume-1",
	"owner_uuid": "7530fb4f-aa7a-4f3c-ff9d-be6f2c510d14",
	"size": 10240,
	"type": "tritonnfs",
	"create_timestamp": "2018-01-12T21:09:09.788Z",
	"state": "creating",
	"networks": ["d251d640-a02e-47b5-8ae6-8b45d859528e"],
	"id": "1edcc6ad-7987-4372-b13a-d21e678ba1e9",
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listVolumesError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list volumes")
}
