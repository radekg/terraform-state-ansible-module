//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package network_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

var (
	fakeFabricVLANID      = 2
	deleteFabricErrorType = errors.New("unable to delete fabric")
	deleteVLanErrorType   = errors.New("unable to delete VLAN")
	createVLanErrorType   = errors.New("unable to create VLAN")
	createFabricErrorType = errors.New("unable to create fabric")
	getVlanErrorType      = errors.New("unable to get VLAN")
	getFabricErrorType    = errors.New("unable to get fabric")
	updateVlanErrorType   = errors.New("unable to update VLAN")
	listVlansErrorType    = errors.New("unable to list VLANs")
	listFabricsErrorType  = errors.New("unable to list fabrics")
)

func TestDeleteFabric(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) error {
		defer testutils.DeactivateClient()

		return nc.Fabrics().Delete(ctx, &network.DeleteFabricInput{
			FabricVLANID: fakeFabricVLANID,
			NetworkID:    fakeNetworkID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks", fakeNetworkID), deleteFabricSuccess)

		err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), deleteFabricError)

		err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete fabric") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestDeleteVLANs(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) error {
		defer testutils.DeactivateClient()

		return nc.Fabrics().DeleteVLAN(ctx, &network.DeleteVLANInput{
			ID: fakeFabricVLANID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), deleteVLANSuccess)

		err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fabrics", "default", "vlans"), deleteVLANError)

		err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete VLAN") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestUpdateVLAN(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FabricVLAN, error) {
		defer testutils.DeactivateClient()

		vlan, err := nc.Fabrics().UpdateVLAN(ctx, &network.UpdateVLANInput{
			Description: "my description updated",
			ID:          fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return vlan, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), updateVLANSuccess)

		_, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), updateVLANError)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update VLAN") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestCreateFabric(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.Network, error) {
		defer testutils.DeactivateClient()

		network, err := nc.Fabrics().Create(ctx, &network.CreateFabricInput{
			Name:             "new",
			ProvisionStartIP: "10.50.1.5",
			ProvisionEndIP:   "10.50.1.20",
			Gateway:          "10.50.1.1",
			Subnet:           "10.50.1.0/24",
			InternetNAT:      false,
			FabricVLANID:     fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return network, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), createFabricSuccess)

		_, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), createFabricError)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create fabric") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetVLAN(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FabricVLAN, error) {
		defer testutils.DeactivateClient()

		vlan, err := nc.Fabrics().GetVLAN(ctx, &network.GetVLANInput{
			ID: fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return vlan, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), getVLANSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), getVLANEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), getVLANBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans"), getVLANError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get VLAN") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetFabric(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.Network, error) {
		defer testutils.DeactivateClient()

		network, err := nc.Fabrics().Get(ctx, &network.GetFabricInput{
			NetworkID:    fakeNetworkID,
			FabricVLANID: fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return network, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks", fakeNetworkID), getFabricSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks", fakeNetworkID), getFabricEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks", fakeNetworkID), getFabricBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), getFabricError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get fabric") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestCreateVLAN(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FabricVLAN, error) {
		defer testutils.DeactivateClient()

		vlan, err := nc.Fabrics().CreateVLAN(ctx, &network.CreateVLANInput{
			Name:        "new",
			Description: "my description",
			ID:          fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return vlan, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fabrics", "default", "vlans"), createVLANSuccess)

		_, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fabrics", "default", "vlans"), createVLANError)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create VLAN") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestListVLANs(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.FabricVLAN, error) {
		defer testutils.DeactivateClient()

		vlans, err := nc.Fabrics().ListVLANs(ctx, &network.ListVLANsInput{})
		if err != nil {
			return nil, err
		}
		return vlans, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans"), listVLANsSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans"), listVLANsEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans"), listVLANsBadDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans"), listVLANsError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list VLANs") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestListFabrics(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.Network, error) {
		defer testutils.DeactivateClient()

		fabrics, err := nc.Fabrics().List(ctx, &network.ListFabricsInput{
			FabricVLANID: fakeFabricVLANID,
		})
		if err != nil {
			return nil, err
		}
		return fabrics, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), listFabricsSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), listFabricsEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID), "networks"), listFabricsBadDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fabrics", "default", "vlans", strconv.Itoa(fakeFabricVLANID)), listFabricsError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list fabrics") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func deleteFabricSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteFabricError(req *http.Request) (*http.Response, error) {
	return nil, deleteFabricErrorType
}

func deleteVLANSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteVLANError(req *http.Request) (*http.Response, error) {
	return nil, deleteVLanErrorType
}

func createVLANSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "new",
  "description": "my description",
  "vlan_id": 2
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createVLANError(req *http.Request) (*http.Response, error) {
	return nil, createVLanErrorType
}

func createFabricSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "7fa999c8-0d2c-453e-989c-e897716d0831",
  "name": "newnet",
  "public": false,
  "fabric": true,
  "gateway": "10.50.1.1",
  "internet_nat": true,
  "provision_end_ip": "10.50.1.20",
  "provision_start_ip": "10.50.1.5",
  "resolvers": [
    "8.8.8.8",
    "8.8.4.4"
  ],
  "routes": {
    "10.25.1.0/21": "10.50.1.2",
    "10.27.1.0/21": "10.50.1.3"
  },
  "subnet": "10.50.1.0/24",
  "vlan_id": 2
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createFabricError(req *http.Request) (*http.Response, error) {
	return nil, createFabricErrorType
}

func getVLANSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "default",
  "vlan_id": 2
}
`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getVLANError(req *http.Request) (*http.Response, error) {
	return nil, getVlanErrorType
}

func getVLANBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "default",
  "vlan_id": 2,
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getVLANEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getFabricSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "default",
  "vlan_id": 2
}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getFabricError(req *http.Request) (*http.Response, error) {
	return nil, getFabricErrorType
}

func getFabricBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "default",
  "vlan_id": 2,
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getFabricEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func updateVLANSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "name": "new",
  "description": "my description updated",
  "vlan_id": 2
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateVLANError(req *http.Request) (*http.Response, error) {
	return nil, updateVlanErrorType
}

func listVLANsEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listVLANsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "name": "default",
    "vlan_id": 2
  }
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listVLANsBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "name": "default",
    "vlan_id": 2,
  }
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listVLANsError(req *http.Request) (*http.Response, error) {
	return nil, listVlansErrorType
}

func listFabricsEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listFabricsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "7326787b-8039-436c-a533-5038f7280f04",
    "name": "default",
    "public": false,
    "fabric": true,
    "gateway": "192.168.128.1",
    "internet_nat": true,
    "provision_end_ip": "192.168.131.250",
    "provision_start_ip": "192.168.128.5",
    "resolvers": [
      "8.8.8.8",
      "8.8.4.4"
    ],
    "subnet": "192.168.128.0/22",
    "vlan_id": 2
  },
  {
    "id": "7fa999c8-0d2c-453e-989c-e897716d0831",
    "name": "newnet",
    "public": false,
    "fabric": true,
    "gateway": "10.50.1.1",
    "internet_nat": true,
    "provision_end_ip": "10.50.1.20",
    "provision_start_ip": "10.50.1.2",
    "resolvers": [
      "8.8.8.8",
      "8.8.4.4"
    ],
    "routes": {
      "10.25.1.0/21": "10.50.1.2",
      "10.27.1.0/21": "10.50.1.3"
    },
    "subnet": "10.50.1.0/24",
    "vlan_id": 2
  }
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listFabricsBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "7326787b-8039-436c-a533-5038f7280f04",
    "name": "default",
    "public": false,
    "fabric": true,
    "gateway": "192.168.128.1",
    "internet_nat": true,
    "provision_end_ip": "192.168.131.250",
    "provision_start_ip": "192.168.128.5",
    "resolvers": [
      "8.8.8.8",
      "8.8.4.4"
    ],
    "subnet": "192.168.128.0/22",
    "vlan_id": 2
  },
  {
    "id": "7fa999c8-0d2c-453e-989c-e897716d0831",
    "name": "newnet",
    "public": false,
    "fabric": true,
    "gateway": "10.50.1.1",
    "internet_nat": true,
    "provision_end_ip": "10.50.1.20",
    "provision_start_ip": "10.50.1.2",
    "resolvers": [
      "8.8.8.8",
      "8.8.4.4"
    ],
    "routes": {
      "10.25.1.0/21": "10.50.1.2",
      "10.27.1.0/21": "10.50.1.3"
    },
    "subnet": "10.50.1.0/24",
    "vlan_id": 2,
  }
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listFabricsError(req *http.Request) (*http.Response, error) {
	return nil, listFabricsErrorType
}
