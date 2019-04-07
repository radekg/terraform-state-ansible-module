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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

const accountURL = "testing"

var (
	fakeFirewallRuleID        = "38de17c4-39e8-48c7-a168-0f58083de860"
	fakeMachineID             = "3d51f2d5-46f2-4da5-bb04-3238f2f64768"
	listRuleMachinesErrorType = errors.New("unable to list firewall rule machines")
	listMachineRulesErrorType = errors.New("unable to list machine firewall rules")
	deleteRuleErrorType       = errors.New("unable to delete firewall rule")
	disableRuleErrorType      = errors.New("unable to disable firewall rule")
	enableRuleErrorType       = errors.New("unable to enable firewall rule")
	createRuleErrorType       = errors.New("unable to create firewall rule")
	updateRuleErrorType       = errors.New("unable to update firewall rule")
	getRuleErrorType          = errors.New("unable to get firewall rule")
	listRulesErrorType        = errors.New("unable to list firewall rules")
)

func MockNetworkClient() *network.NetworkClient {
	return &network.NetworkClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountURL,
		}),
	}
}

func TestListRuleMachines(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.Machine, error) {
		defer testutils.DeactivateClient()

		machines, err := nc.Firewall().ListRuleMachines(ctx, &network.ListRuleMachinesInput{
			ID: fakeFirewallRuleID,
		})
		if err != nil {
			return nil, err
		}

		return machines, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "machines"), listRuleMachinesSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "machines"), listRuleMachinesEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "machines"), listRuleMachinesBadDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "machines"), listRuleMachinesError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list firewall rule machines") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestListMachineRules(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rules, err := nc.Firewall().ListMachineRules(ctx, &network.ListMachineRulesInput{
			MachineID: fakeMachineID,
		})
		if err != nil {
			return nil, err
		}

		return rules, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "fwrules"), listMachineRulesSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "fwrules"), listMachineRulesEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "fwrules"), listMachineRulesBadDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", "fakeMachineName", "fwrules"), listMachineRulesError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list machine firewall rules") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestDeleteRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) error {
		defer testutils.DeactivateClient()

		return nc.Firewall().DeleteRule(ctx, &network.DeleteRuleInput{
			ID: fakeFirewallRuleID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), deleteRuleSuccess)

		err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "fwrules"), deleteRuleError)

		err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestDisableRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rule, err := nc.Firewall().DisableRule(ctx, &network.DisableRuleInput{
			ID: fakeFirewallRuleID,
		})
		if err != nil {
			return nil, err
		}

		return rule, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "disable"), disableRuleSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "disable"), disableRuleEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "disable"), disableRuleBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), disableRuleError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to disable firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestEnableRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rule, err := nc.Firewall().EnableRule(ctx, &network.EnableRuleInput{
			ID: fakeFirewallRuleID,
		})
		if err != nil {
			return nil, err
		}

		return rule, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "enable"), enableRuleSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "enable"), enableRuleEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID, "enable"), enableRuleBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), enableRuleError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to enable firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestUpdateRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rule, err := nc.Firewall().UpdateRule(ctx, &network.UpdateRuleInput{
			ID:   fakeFirewallRuleID,
			Rule: "FROM any TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
		})
		if err != nil {
			return nil, err
		}

		return rule, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), updateRuleSuccess)

		_, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), updateRuleError)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestCreateRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rule, err := nc.Firewall().CreateRule(ctx, &network.CreateRuleInput{
			Rule:    fmt.Sprintf("FROM vm %s TO subnet 10.99.99.0/24 BLOCK tcp PORT 25", fakeMachineID),
			Enabled: true,
		})
		if err != nil {
			return nil, err
		}

		return rule, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules"), createRuleSuccess)

		_, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "fwrules"), createRuleError)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetRule(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) (*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rule, err := nc.Firewall().GetRule(ctx, &network.GetRuleInput{
			ID: fakeFirewallRuleID,
		})
		if err != nil {
			return nil, err
		}

		return rule, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), getRuleSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), getRuleEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules", fakeFirewallRuleID), getRuleBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules"), getRuleError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get firewall rule") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestListUsers(t *testing.T) {
	networkClient := MockNetworkClient()

	do := func(ctx context.Context, nc *network.NetworkClient) ([]*network.FirewallRule, error) {
		defer testutils.DeactivateClient()

		rules, err := nc.Firewall().ListRules(ctx, &network.ListRulesInput{})
		if err != nil {
			return nil, err
		}

		return rules, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules"), listRulesSuccess)

		resp, err := do(context.Background(), networkClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules"), listRulesEmpty)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules"), listRulesBadeDecode)

		_, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "fwrules"), listRulesError)

		resp, err := do(context.Background(), networkClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list firewall rules") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func listRuleMachinesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listRuleMachinesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "b6979942-7d5d-4fe6-a2ec-b812e950625a",
    "name": "test",
    "type": "smartmachine",
    "brand": "joyent",
    "state": "running",
    "image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
    "ips": [
      "10.88.88.26",
      "192.168.128.5"
    ],
    "memory": 128,
    "disk": 12288,
    "metadata": {
      "root_authorized_keys": "test-key"
    },
    "tags": {},
    "created": "2016-01-04T12:55:50.539Z",
    "updated": "2016-01-21T08:56:59.000Z",
    "networks": [
      "a9c130da-e3ba-40e9-8b18-112aba2d3ba7",
      "45607081-4cd2-45c8-baf7-79da760fffaa"
    ],
    "primaryIp": "10.88.88.26",
    "firewall_enabled": false,
    "compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
    "package": "sdc_128"
  }
]`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRuleMachinesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "b6979942-7d5d-4fe6-a2ec-b812e950625a",
    "name": "test",
    "type": "smartmachine",
    "brand": "joyent",
    "state": "running",
    "image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
    "ips": [
      "10.88.88.26",
      "192.168.128.5"
    ],
    "memory": 128,
    "disk": 12288,
    "metadata": {
      "root_authorized_keys": "test-key"
    },
    "tags": {},
    "created": "2016-01-04T12:55:50.539Z",
    "updated": "2016-01-21T08:56:59.000Z",
    "networks": [
      "a9c130da-e3ba-40e9-8b18-112aba2d3ba7",
      "45607081-4cd2-45c8-baf7-79da760fffaa"
    ],
    "primaryIp": "10.88.88.26",
    "firewall_enabled": false,
    "compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
    "package": "sdc_128",
  }
]`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRuleMachinesError(req *http.Request) (*http.Response, error) {
	return nil, listRuleMachinesErrorType
}

func listMachineRulesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listMachineRulesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "38de17c4-39e8-48c7-a168-0f58083de860",
    "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
    "enabled": true
  }
]`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineRulesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "38de17c4-39e8-48c7-a168-0f58083de860",
    "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
    "enabled": true,
  }
]`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineRulesError(req *http.Request) (*http.Response, error) {
	return nil, listMachineRulesErrorType
}

func deleteRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteRuleError(req *http.Request) (*http.Response, error) {
	return nil, deleteRuleErrorType
}

func disableRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp (PORT 25 AND PORT 80)",
  "enabled": false
}
`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func disableRuleError(req *http.Request) (*http.Response, error) {
	return nil, disableRuleErrorType
}

func disableRuleBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp (PORT 25 AND PORT 80)",
  "enabled": false,
}`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func disableRuleEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func enableRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp (PORT 25 AND PORT 80)",
  "enabled": true
}
`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func enableRuleError(req *http.Request) (*http.Response, error) {
	return nil, enableRuleErrorType
}

func enableRuleBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp (PORT 25 AND PORT 80)",
  "enabled": true,
}`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func enableRuleEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func createRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
  "enabled": true
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createRuleError(req *http.Request) (*http.Response, error) {
	return nil, createRuleErrorType
}

func updateRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM any TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
  "enabled": true
}
`)

	return &http.Response{
		StatusCode: http.StatusCreated,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateRuleError(req *http.Request) (*http.Response, error) {
	return nil, updateRuleErrorType
}

func getRuleSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
  "enabled": true
}
`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getRuleError(req *http.Request) (*http.Response, error) {
	return nil, getRuleErrorType
}

func getRuleBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "38de17c4-39e8-48c7-a168-0f58083de860",
  "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
  "enabled": true,
}`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getRuleEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listRulesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listRulesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
    "id": "38de17c4-39e8-48c7-a168-0f58083de860",
    "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
    "enabled": true
  }
]`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRulesBadeDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{[
	{
    "id": "38de17c4-39e8-48c7-a168-0f58083de860",
    "rule": "FROM vm 3d51f2d5-46f2-4da5-bb04-3238f2f64768 TO subnet 10.99.99.0/24 BLOCK tcp PORT 25",
    "enabled": true
  }
]}`)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listRulesError(req *http.Request) (*http.Response, error) {
	return nil, listRulesErrorType
}
