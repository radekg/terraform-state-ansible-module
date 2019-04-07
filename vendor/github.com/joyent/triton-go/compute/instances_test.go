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
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"strings"

	"path"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

var (
	fakeMachineID         = "75cfe125-a5ce-49e8-82ac-09aa31ffdf26"
	fakeMachinePackageID  = "7041ccc7-3f9e-cf1e-8c85-a9ee41b7f968"
	fakeMachineTag        = "my-test-tag"
	fakeMachineMetaDataID = "foo"
	fakeMacID             = "90b8d02fb8f9"
	fakeNetworkID         = "7007b198-f6aa-48f0-9843-78a3149de3d7"
)

func getAnyInstanceID(t *testing.T, client *compute.ComputeClient) (string, error) {
	ctx := context.Background()
	input := &compute.ListInstancesInput{}
	instances, err := client.Instances().List(ctx, input)
	if err != nil {
		return "", err
	}

	for _, m := range instances {
		if len(m.ID) > 0 {
			return m.ID, nil
		}
	}

	t.Skip()
	return "", errors.New("no machines configured")
}

func RandInt() int {
	reseed()
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int()
}

func RandWithPrefix(name string) string {
	return fmt.Sprintf("%s-%d", name, RandInt())
}

// Seeds random with current timestamp
func reseed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestAccInstances_Create(t *testing.T) {
	testInstanceName := RandWithPrefix("acctest")

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					computeClient, err := compute.NewClient(config)
					if err != nil {
						return nil, err
					}

					networkClient, err := network.NewClient(config)
					if err != nil {
						return nil, err
					}

					return []interface{}{
						computeClient,
						networkClient,
					}, nil
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					clients := client.([]interface{})
					c := clients[0].(*compute.ComputeClient)
					n := clients[1].(*network.NetworkClient)

					images, err := c.Images().List(context.Background(), &compute.ListImagesInput{
						Name:    "ubuntu-16.04",
						Version: "20170403",
					})
					if err != nil {
						return nil, err
					}

					img := images[0]

					var net *network.Network
					networkName := "Joyent-SDC-Private"
					nets, err := n.List(context.Background(), &network.ListInput{})
					if err != nil {
						return nil, err
					}
					for _, found := range nets {
						if found.Name == networkName {
							net = found
						}
					}

					input := &compute.CreateInstanceInput{
						Name:     testInstanceName,
						Package:  "g4-highcpu-128M",
						Image:    img.ID,
						Networks: []string{net.Id},
						Metadata: map[string]string{
							"metadata1": "value1",
						},
						Tags: map[string]string{
							"tag1": "value1",
						},
						CNS: compute.InstanceCNS{
							Services: []string{"testapp", "testweb"},
						},
					}
					created, err := c.Instances().Create(context.Background(), input)
					if err != nil {
						return nil, err
					}

					state := make(chan *compute.Instance, 1)
					go func(createdID string, c *compute.ComputeClient) {
						for {
							time.Sleep(1 * time.Second)
							instance, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
								ID: createdID,
							})
							if err != nil {
								log.Fatalf("Get(): %v", err)
							}
							if instance.State == "running" {
								state <- instance
							}
						}
					}(created.ID, c)

					select {
					case instance := <-state:
						return instance, nil
					case <-time.After(5 * time.Minute):
						return nil, fmt.Errorf("Timed out waiting for instance to provision")
					}
				},
				CleanupFunc: func(client interface{}, stateBag interface{}) {
					instance, instOk := stateBag.(*compute.Instance)
					if !instOk {
						log.Println("Expected instance to be Instance")
						return
					}

					if instance.Name != testInstanceName {
						log.Printf("Expected instance to be named %s: found %s\n",
							testInstanceName, instance.Name)
						return
					}

					clients := client.([]interface{})
					c, clientOk := clients[0].(*compute.ComputeClient)
					if !clientOk {
						log.Println("Expected client to be ComputeClient")
						return
					}

					err := c.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
						ID: instance.ID,
					})
					if err != nil {
						log.Printf("Could not delete instance %s\n", instance.Name)
					}
					return
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					instanceRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}
					instance, ok := instanceRaw.(*compute.Instance)
					if !ok {
						return errors.New("Expected state to include instance")
					}

					if instance.State != "running" {
						return fmt.Errorf("Expected instance state to be \"running\": found %s",
							instance.State)
					}
					if instance.ID == "" {
						return fmt.Errorf("Expected instance ID: found \"\"")
					}
					if instance.Name == "" {
						return fmt.Errorf("Expected instance Name: found \"\"")
					}
					if instance.Memory != 128 {
						return fmt.Errorf("Expected instance Memory to be 128: found \"%d\"",
							instance.Memory)
					}

					metadataVal, metaOk := instance.Metadata["metadata1"]
					if !metaOk {
						return fmt.Errorf("Expected instance to have Metadata: found \"%v\"",
							instance.Metadata)
					}
					if metadataVal != "value1" {
						return fmt.Errorf("Expected instance Metadata \"metadata1\" to equal \"value1\": found \"%s\"",
							metadataVal)
					}

					tagVal, tagOk := instance.Tags["tag1"]
					if !tagOk {
						return fmt.Errorf("Expected instance to have Tags: found \"%v\"",
							instance.Tags)
					}
					if tagVal != "value1" {
						return fmt.Errorf("Expected instance Tag \"tag1\" to equal \"value1\": found \"%s\"",
							tagVal)
					}

					services := []string{"testapp", "testweb"}
					if !reflect.DeepEqual(instance.CNS.Services, services) {
						return fmt.Errorf("Expected instance CNS Services \"%s\", to equal \"%v\"",
							instance.CNS.Services, services)
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_Get(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &compute.GetInstanceInput{
						ID: instanceID,
					}
					return c.Instances().Get(ctx, input)
				},
			},

			&testutils.StepAssertSet{
				StateBagKey: "instances",
				Keys:        []string{"ID", "Name", "Type", "Tags"},
			},
		},
	})
}

// FIXME(seanc@): TestAccMachine_ListMachineTags assumes that any machine ID
// returned from getAnyInstanceID will have at least one tag.
func TestAccInstances_ListTags(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &compute.ListTagsInput{
						ID: instanceID,
					}
					return c.Instances().ListTags(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					tagsRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					tags := tagsRaw.(map[string]interface{})
					if len(tags) == 0 {
						return errors.New("Expected at least one tag on machine")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_UpdateMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &compute.UpdateMetadataInput{
						ID: instanceID,
						Metadata: map[string]string{
							"tester": os.Getenv("USER"),
						},
					}
					return c.Instances().UpdateMetadata(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					mdata := mdataRaw.(map[string]string)
					if len(mdata) == 0 {
						return errors.New("Expected metadata on machine")
					}

					if mdata["tester"] != os.Getenv("USER") {
						return errors.New("Expected test metadata to equal environ $USER")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_ListMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &compute.ListMetadataInput{
						ID: instanceID,
					}
					return c.Instances().ListMetadata(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataRaw, found := state.GetOk("instances")
					if !found {
						return fmt.Errorf("State key %q not found", "instances")
					}

					mdata := mdataRaw.(map[string]string)
					if len(mdata) == 0 {
						return errors.New("Expected metadata on machine")
					}

					if mdata["root_authorized_keys"] == "" {
						return errors.New("Expected test metadata to have key")
					}
					return nil
				},
			},
		},
	})
}

func TestAccInstances_GetMetadata(t *testing.T) {
	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: "instances",
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: "instances",
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)

					instanceID, err := getAnyInstanceID(t, c)
					if err != nil {
						return nil, err
					}

					ctx := context.Background()
					input := &compute.UpdateMetadataInput{
						ID: instanceID,
						Metadata: map[string]string{
							"testkey": os.Getenv("USER"),
						},
					}
					_, err = c.Instances().UpdateMetadata(ctx, input)
					if err != nil {
						return nil, err
					}

					ctx2 := context.Background()
					input2 := &compute.GetMetadataInput{
						ID:  instanceID,
						Key: "testkey",
					}
					return c.Instances().GetMetadata(ctx2, input2)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					mdataValue := state.Get("instances")
					retValue := fmt.Sprintf("\"%s\"", os.Getenv("USER"))
					if mdataValue != retValue {
						return errors.New("Expected test metadata to equal environ \"$USER\"")
					}
					return nil
				},
			},
		},
	})
}

func TestValidateInstanceInput(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		input := &compute.GetInstanceInput{
			ID: fakeImageID,
		}
		err := input.Validate()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		input := &compute.GetInstanceInput{}
		err := input.Validate()
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "machine ID can not be empty") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestGetInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Instance, error) {
		defer testutils.DeactivateClient()

		instance, err := cc.Instances().Get(ctx, &compute.GetInstanceInput{
			ID: fakeMachineID,
		})
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID), getMachineSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID), getMachineEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID), getMachineBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", "not-a-real-instance-id"), getMachineError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListInstances(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Instance, error) {
		defer testutils.DeactivateClient()

		instances, err := cc.Instances().List(ctx, &compute.ListInstancesInput{
			Offset: 100,
		})
		if err != nil {
			return nil, err
		}
		return instances, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines?offset=100"), listMachinesSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines?offset=100"), listMachinesEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines?offset=100"), listMachinesBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines?offset=100"), listMachinesError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list machines") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestDeleteInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Delete(ctx, &compute.DeleteInstanceInput{
			ID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID), deleteMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID), deleteMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestRenameInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Rename(ctx, &compute.RenameInstanceInput{
			ID:   fakeMachineID,
			Name: "new-machine-name",
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=rename&name=new-machine-name", accountURL, fakeMachineID), renameMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=rename&name=new-machine-name", accountURL, fakeMachineID), renameMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to rename machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestRebootInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Reboot(ctx, &compute.RebootInstanceInput{
			InstanceID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=reboot", accountURL, fakeMachineID), rebootMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=reboot", accountURL, fakeMachineID), rebootMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to reboot machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestStartInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Start(ctx, &compute.StartInstanceInput{
			InstanceID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=start", accountURL, fakeMachineID), startMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=start", accountURL, fakeMachineID), startMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to start machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestStopInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Stop(ctx, &compute.StopInstanceInput{
			InstanceID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=stop", accountURL, fakeMachineID), stopMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=stop", accountURL, fakeMachineID), stopMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to stop machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestResizeInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().Resize(ctx, &compute.ResizeInstanceInput{
			ID:      fakeMachineID,
			Package: fakeMachinePackageID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=resize&package=%s", accountURL, fakeMachineID, fakeMachinePackageID), resizeMachineSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=resize&package=%s", accountURL, fakeMachineID, fakeMachinePackageID), resizeMachineError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to resize machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestEnableInstanceFirewall(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().EnableFirewall(ctx, &compute.EnableFirewallInput{
			ID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=enable_firewall", accountURL, fakeMachineID), enableMachineFirewallSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=enable_firewall", accountURL, fakeMachineID), enableMachineFirewallError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to enable machine firewall") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDisableInstanceFirewall(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DisableFirewall(ctx, &compute.DisableFirewallInput{
			ID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=disable_firewall", accountURL, fakeMachineID), disableMachineFirewallSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=disable_firewall", accountURL, fakeMachineID), disableMachineFirewallError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to disable machine firewall") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteInstanceTags(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DeleteTags(ctx, &compute.DeleteTagsInput{
			ID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), deleteMachineTagsSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), deleteMachineTagsError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete machine tags") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteInstanceTag(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DeleteTag(ctx, &compute.DeleteTagInput{
			ID:  fakeMachineID,
			Key: fakeMachineTag,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "tags", fakeMachineTag), deleteMachineTagSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "tags", fakeMachineTag), deleteMachineTagError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete machine tag") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestGetInstanceTag(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (string, error) {
		defer testutils.DeactivateClient()

		tag, err := cc.Instances().GetTag(ctx, &compute.GetTagInput{
			ID:  fakeMachineID,
			Key: fakeMachineTag,
		})
		if err != nil {
			return "", err
		}

		return tag, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "tags", fakeMachineTag), getMachineTagSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == "" {
			t.Fatalf("Expected an output but got empty string")
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "tags", fakeMachineTag), getMachineTagError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != "" {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get machine tag") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListInstanceTag(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (map[string]interface{}, error) {
		defer testutils.DeactivateClient()

		tag, err := cc.Instances().ListTags(ctx, &compute.ListTagsInput{
			ID: fakeMachineID,
		})
		if err != nil {
			return nil, err
		}

		return tag, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), listMachineTagSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got empty string")
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), listMachineTagError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list machine tag") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestReplaceInstanceTags(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().ReplaceTags(ctx, &compute.ReplaceTagsInput{
			ID: fakeMachineID,
			Tags: map[string]string{
				"foo":   "bar",
				"group": "test",
			},
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), replaceMachineTagsSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("PUT", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), replaceMachineTagsError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to replace machine tags") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestAddInstanceTags(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().AddTags(ctx, &compute.AddTagsInput{
			ID: fakeMachineID,
			Tags: map[string]string{
				"foo":   "bar",
				"group": "test",
			},
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), addMachineTagsSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "tags"), addMachineTagsError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to add tags to machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestGetInstanceMetaData(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (string, error) {
		defer testutils.DeactivateClient()

		tag, err := cc.Instances().GetMetadata(ctx, &compute.GetMetadataInput{
			ID:  fakeMachineID,
			Key: fakeMachineMetaDataID,
		})
		if err != nil {
			return "", err
		}

		return tag, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "metadata", fakeMachineMetaDataID), getMachineMetaDataSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == "" {
			t.Fatalf("Expected an output but got empty string")
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "metadata", fakeMachineMetaDataID), getMachineMetaDataError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != "" {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get machine metadata") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListInstanceMetaData(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (map[string]string, error) {
		defer testutils.DeactivateClient()

		metadata, err := cc.Instances().ListMetadata(ctx, &compute.ListMetadataInput{
			ID: fakeMachineID,
		})
		if err != nil {
			return nil, err
		}

		return metadata, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), listMachineMetaDataSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got empty string")
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), listMachineMetaDataError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list machine metadata") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteInstanceMetaData(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DeleteMetadata(ctx, &compute.DeleteMetadataInput{
			ID:  fakeMachineID,
			Key: fakeMachineMetaDataID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "metadata", fakeMachineMetaDataID), deleteMachineMetaDataSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "metadata", fakeMachineMetaDataID), deleteMachineMetaDataError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete machine metadata") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDeleteAllInstanceMetaData(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DeleteAllMetadata(ctx, &compute.DeleteAllMetadataInput{
			ID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), deleteAllMachineMetaDataSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), deleteAllMachineMetaDataError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to delete all machine metadata") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestUpdateInstanceMetaData(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (map[string]string, error) {
		defer testutils.DeactivateClient()

		metadata, err := cc.Instances().UpdateMetadata(ctx, &compute.UpdateMetadataInput{
			ID: fakeMachineID,
			Metadata: map[string]string{
				"foo":   "bar",
				"group": "test",
			},
		})
		if err != nil {
			return nil, err
		}

		return metadata, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), updateMachineMetaDataSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "metadata"), updateMachineMetaDataError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to update machine metadata") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestListInstanceNICs(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.NIC, error) {
		defer testutils.DeactivateClient()

		nics, err := cc.Instances().ListNICs(ctx, &compute.ListNICsInput{
			InstanceID: fakeMachineID,
		})
		if err != nil {
			return nil, err
		}
		return nics, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), listMachineNICsSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), listMachineNICsEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), listMachineNICsBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), listMachineNICsError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to list machine NICs") {
			t.Errorf("expected error to equal testError: found %v", err)
		}
	})
}

func TestGetInstanceNIC(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.NIC, error) {
		defer testutils.DeactivateClient()

		nic, err := cc.Instances().GetNIC(ctx, &compute.GetNICInput{
			InstanceID: fakeMachineID,
			MAC:        fakeMacID,
		})
		if err != nil {
			return nil, err
		}
		return nic, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics", fakeMacID), getMachineNICSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatalf("Expected an output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics", fakeMacID), getMachineNICEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", fakeMachineID, "nics", fakeMacID), getMachineNICBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "machines", "not-a-real-instance-id", "nics", fakeMacID), getMachineNICError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), "unable to get machine NIC") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestRemoveInstanceNIC(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().RemoveNIC(ctx, &compute.RemoveNICInput{
			InstanceID: fakeMachineID,
			MAC:        fakeMacID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "nics", fakeMacID), deleteMachineNICSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", path.Join("/", accountURL, "machines", fakeMachineID, "nics", fakeMacID), deleteMachineNICError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to remove NIC from machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestAddNICToInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.NIC, error) {
		defer testutils.DeactivateClient()

		nic, err := cc.Instances().AddNIC(ctx, &compute.AddNICInput{
			InstanceID: fakeMachineID,
			Network:    fakeNetworkID,
		})
		if err != nil {
			return nil, err
		}
		return nic, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), createMachineNICSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines", fakeMachineID, "nics"), createMachineNICError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to add NIC to machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})

}

func TestCreateInstance(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (*compute.Instance, error) {
		defer testutils.DeactivateClient()

		instance, err := cc.Instances().Create(ctx, &compute.CreateInstanceInput{
			Image:   "2b683a82-a066-11e3-97ab-2faa44701c5a",
			Package: "7b17343c-94af-6266-e0e8-893a3b9993d0",
		})
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines"), createMachineSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines"), createMachineError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to create machine") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestCountInstances(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) (int, error) {
		defer testutils.DeactivateClient()

		instances, err := cc.Instances().Count(ctx, &compute.ListInstancesInput{})
		if err != nil {
			return -1, err
		}
		return instances, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("HEAD", path.Join("/", accountURL, "machines?offset=0"), countMachinesSuccess)

		_, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", path.Join("/", accountURL, "machines?offset=0"), countMachinesError)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to get machines count") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestEnableDeletionProtection(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().EnableDeletionProtection(ctx, &compute.EnableDeletionProtectionInput{
			InstanceID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=enable_deletion_protection", accountURL, fakeMachineID), enableDeletionProtectionSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=enable_deletion_protection", accountURL, fakeMachineID), enableDeletionProtectionError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to enable deletion protection") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func TestDisableDeletionProtection(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		return cc.Instances().DisableDeletionProtection(ctx, &compute.DisableDeletionProtectionInput{
			InstanceID: fakeMachineID,
		})
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=disable_deletion_protection", accountURL, fakeMachineID), disableDeletionProtectionSuccess)

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("POST", fmt.Sprintf("/%s/machines/%s?action=disable_deletion_protection", accountURL, fakeMachineID), disableDeletionProtectionError)

		err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "unable to disable deletion protection") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})
}

func getMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
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
	"root_authorized_keys": "..."
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
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
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
	"root_authorized_keys": "...",
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
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get instance")
}

func listMachinesSuccess(req *http.Request) (*http.Response, error) {
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
	  "root_authorized_keys": "..."
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

func listMachinesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listMachinesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
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
	  "root_authorized_keys": "..."
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
  }]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachinesError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list machines")
}

func deleteMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete machine")
}

func renameMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func renameMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to rename machine")
}

func rebootMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func rebootMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to reboot machine")
}

func startMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func startMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to start machine")
}

func stopMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func stopMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to stop machine")
}

func resizeMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func resizeMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to resize machine")
}

func enableMachineFirewallSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func enableMachineFirewallError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to enable machine firewall")
}

func disableMachineFirewallSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func disableMachineFirewallError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to disable machine firewall")
}

func deleteMachineTagsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteMachineTagsError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete machine tags")
}

func deleteMachineTagSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteMachineTagError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete machine tag")
}

func getMachineTagSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`"bar"
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineTagError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get machine tag")
}

func listMachineTagSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{"foo":"bar","group":"test"}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineTagError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list machine tag")
}

func getMachineMetaDataSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`"bar"
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineMetaDataError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get machine metadata")
}

func listMachineMetaDataSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{"foo":"bar","group":"test"}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineMetaDataError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list machine metadata")
}

func deleteMachineMetaDataSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteMachineMetaDataError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete machine metadata")
}

func deleteAllMachineMetaDataSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteAllMachineMetaDataError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to delete machine metadata")
}

func listMachineNICsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
	{
		"mac": "90:b8:d0:2f:b8:f9",
		"primary": true,
		"ip": "10.88.88.137",
		"netmask": "255.255.255.0",
		"gateway": "10.88.88.2",
		"state": "running",
		"network": "6b3229b6-c535-11e5-8cf9-c3a24fa96e35"
	}
]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineNICsEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func listMachineNICsBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
		"mac": "90:b8:d0:2f:b8:f9",
		"primary": true,
		"ip": "10.88.88.137",
		"netmask": "255.255.255.0",
		"gateway": "10.88.88.2",
		"state": "running",
		"network": "6b3229b6-c535-11e5-8cf9-c3a24fa96e35",
	}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listMachineNICsError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to list machine NICs")
}

func getMachineNICSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
		"mac": "90:b8:d0:2f:b8:f9",
		"primary": true,
		"ip": "10.88.88.137",
		"netmask": "255.255.255.0",
		"gateway": "10.88.88.2",
		"state": "running",
		"network": "6b3229b6-c535-11e5-8cf9-c3a24fa96e35"
	}
`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineNICBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
		"mac": "90:b8:d0:2f:b8:f9",
		"primary": true,
		"ip": "10.88.88.137",
		"netmask": "255.255.255.0",
		"gateway": "10.88.88.2",
		"state": "running",
		"network": "6b3229b6-c535-11e5-8cf9-c3a24fa96e35",
	}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getMachineNICEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getMachineNICError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get instance NIC")
}

func deleteMachineNICSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func deleteMachineNICError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to remove NIC from machine")
}

func createMachineNICSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"network": "7007b198-f6aa-48f0-9843-78a3149de3d7"
}
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createMachineNICError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to add NIC to machine")
}

func createMachineSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id": "e8622950-af78-486c-b682-dd147c938dc6",
  "name": "e8622950",
  "type": "smartmachine",
  "brand": "joyent",
  "state": "provisioning",
  "image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
  "ips": [],
  "memory": 128,
  "disk": 12288,
  "metadata": {
	"root_authorized_keys": "..."
  },
  "tags": {},
  "created": "2016-01-21T12:57:52.759Z",
  "updated": "2016-01-21T12:57:52.979Z",
  "networks": [],
  "firewall_enabled": false,
  "compute_node": null,
  "package": "sdc_128"
}
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createMachineError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to create machine")
}

func countMachinesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("x-resource-count", "3")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func countMachinesError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get machines count")
}

func replaceMachineTagsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
	}, nil
}

func replaceMachineTagsError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to replace machine tags")
}

func addMachineTagsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
	}, nil
}

func addMachineTagsError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to replace machine tags")
}

func updateMachineMetaDataSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "foo": "bar",
  "group": "test"
}
`)

	return &http.Response{
		StatusCode: 201,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateMachineMetaDataError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to update machine metadata")
}

func enableDeletionProtectionSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func enableDeletionProtectionError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to enable deletion protection")
}

func disableDeletionProtectionSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func disableDeletionProtectionError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to disable deletion protection")
}
