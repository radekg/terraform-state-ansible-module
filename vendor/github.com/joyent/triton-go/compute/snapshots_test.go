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
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
)

const accountURL = "testing"

var (
	listSnapshotErrorType             = errors.New("unable to list snapshots")
	getSnapshotErrorType              = errors.New("unable to get snapshot")
	deleteSnapshotErrorType           = errors.New("unable to delete snapshot")
	createSnapshotErrorType           = errors.New("unable to create snapshot")
	startMachineFromSnapshotErrorType = errors.New("unable to start machine")
	testMachineId                     = "123-3456-2335"
)

func TestListSnapshots(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		ping, err := cc.Snapshots().List(ctx, &compute.ListSnapshotsInput{
			MachineID: "123-3456-2335",
		})
		if err != nil {
			return nil, err
		}
		return ping, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), listSnapshotsSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), listSnapshotEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), listSnapshotBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), listSnapshotError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list snapshots") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetSnapshot(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		snapshot, err := cc.Snapshots().Get(ctx, &compute.GetSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), getSnapshotSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), getSnapshotEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), getSnapshotBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), getSnapshotError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get snapshot") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteSnapshot(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Snapshots().Delete(ctx, &compute.DeleteSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), deleteSnapshotSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), deleteSnapshotError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete snapshot") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestStartMachineFromSnapshot(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Snapshots().StartMachine(ctx, &compute.StartMachineFromSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), startMachineFromSnapshotSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", testMachineId, "snapshots", "sample-snapshot"), startMachineFromSnapshotError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to start machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCreateSnapshot(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Snapshot, error) {
		defer testutils.DeactivateClient()

		snapshot, err := cc.Snapshots().Create(ctx, &compute.CreateSnapshotInput{
			MachineID: "123-3456-2335",
			Name:      "sample-snapshot",
		})
		if err != nil {
			return nil, err
		}
		return snapshot, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), createSnapshotSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", testMachineId, "snapshots"), createSnapshotError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create snapshot") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func listSnapshotsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
	"name": "sample-snapshot",
	"state": "queued",
	"updated": "2015-12-23T06:41:11.032Z",
	"created": "2015-12-23T06:41:11.032Z"
  },
  {
	"name": "sample-snapshot-2",
	"state": "queued",
	"updated": "2015-12-23T06:41:11.032Z",
	"created": "2015-12-23T06:41:11.032Z"
  }
]`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listSnapshotEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listSnapshotBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
	"name": "sample-snapshot",
	"state": "queued",
	"updated": "2015-12-23T06:41:11.032Z",
	"created": "2015-12-23T06:41:11.032Z",}]`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, listSnapshotErrorType
}

func getSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "sample-snapshot",
	"state": "queued",
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

func getSnapshotBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "sample-snapshot",
	"state": "queued",
	"updated": "2015-12-23T06:41:11.032Z",
	"created": "2015-12-23T06:41:11.032Z",}`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getSnapshotEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, getSnapshotErrorType
}

func deleteSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 204,
		Header:     header,
	}, nil
}

func deleteSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, deleteSnapshotErrorType
}

func startMachineFromSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 202,
		Header:     header,
	}, nil
}

func startMachineFromSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, startMachineFromSnapshotErrorType
}

func createSnapshotSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"name": "sample-snapshot",
	"state": "queued",
	"updated": "2015-12-23T06:41:11.032Z",
	"created": "2015-12-23T06:41:11.032Z"
  }
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createSnapshotError(req *http.Request) (*http.Response, error) {
	return nil, createSnapshotErrorType
}
