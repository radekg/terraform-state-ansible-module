//
//  Copyright (tcc) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute

import (
	"context"
	"net/http"
	"time"

	"fmt"

	"github.com/imdario/mergo"
	"github.com/joyent/triton-go/cmd/config"
	tcc "github.com/joyent/triton-go/compute"
	terrors "github.com/joyent/triton-go/errors"
	"github.com/pkg/errors"
)

type AgentComputeClient struct {
	client *tcc.ComputeClient
}

func NewComputeClient(cfg *config.TritonClientConfig) (*AgentComputeClient, error) {
	computeClient, err := tcc.NewClient(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Error Creating Triton Compute Client")
	}
	return &AgentComputeClient{
		client: computeClient,
	}, nil
}

func (c *AgentComputeClient) GetPackagesList() ([]*tcc.Package, error) {
	params := &tcc.ListPackagesInput{}

	name := config.GetPkgName()
	if name != "" {
		params.Name = name
	}

	memory := config.GetPkgMemory()
	if memory > 0 {
		params.Memory = int64(memory)
	}

	disk := config.GetPkgDisk()
	if disk > 0 {
		params.Disk = int64(disk)
	}

	swap := config.GetPkgSwap()
	if swap > 0 {
		params.Swap = int64(swap)
	}

	vpcus := config.GetPkgVPCUs()
	if vpcus > 0 {
		params.VCPUs = int64(vpcus)
	}

	packages, err := c.client.Packages().List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (c *AgentComputeClient) GetPackage() (*tcc.Package, error) {
	id := config.GetPkgID()
	if id != "" {
		pkg, err := c.getPackageByID(id)
		if err != nil {
			return nil, err
		}

		return pkg, nil
	}

	name := config.GetPkgName()
	if name != "" {
		pkg, err := c.getPackageByName(name)
		if err != nil {
			return nil, err
		}

		return pkg, nil
	}

	return nil, nil
}

func (c *AgentComputeClient) GetImagesList() ([]*tcc.Image, error) {
	images, err := c.client.Images().List(context.Background(), &tcc.ListImagesInput{})
	if err != nil {
		return nil, err
	}

	return sortImages(images), nil
}

func (c *AgentComputeClient) GetDataCenterList() ([]*tcc.DataCenter, error) {
	params := &tcc.ListDataCentersInput{}

	dcs, err := c.client.Datacenters().List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return dcs, nil
}

func (c *AgentComputeClient) GetServiceList() ([]*tcc.Service, error) {
	params := &tcc.ListServicesInput{}

	services, err := c.client.Services().List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c *AgentComputeClient) DeleteInstance() (*tcc.Instance, error) {
	var machine *tcc.Instance

	id := config.GetMachineID()
	if id != "" {
		instance, err := c.getInstanceByID(id)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	name := config.GetMachineName()
	if name != "" {
		instance, err := c.getInstanceByName(name)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	err := c.client.Instances().Delete(context.Background(), &tcc.DeleteInstanceInput{
		ID: machine.ID,
	})
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func (c *AgentComputeClient) RebootInstance() (*tcc.Instance, error) {
	var machine *tcc.Instance

	id := config.GetMachineID()
	if id != "" {
		instance, err := c.getInstanceByID(id)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	name := config.GetMachineName()
	if name != "" {
		instance, err := c.getInstanceByName(name)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	err := c.client.Instances().Reboot(context.Background(), &tcc.RebootInstanceInput{
		InstanceID: machine.ID,
	})
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func (c *AgentComputeClient) StartInstance() (*tcc.Instance, error) {
	var machine *tcc.Instance

	id := config.GetMachineID()
	if id != "" {
		instance, err := c.getInstanceByID(id)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	name := config.GetMachineName()
	if name != "" {
		instance, err := c.getInstanceByName(name)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	err := c.client.Instances().Start(context.Background(), &tcc.StartInstanceInput{
		InstanceID: machine.ID,
	})
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func (c *AgentComputeClient) StopInstance() (*tcc.Instance, error) {
	var machine *tcc.Instance

	id := config.GetMachineID()
	if id != "" {
		instance, err := c.getInstanceByID(id)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	name := config.GetMachineName()
	if name != "" {
		instance, err := c.getInstanceByName(name)
		if err != nil {
			return nil, err
		}

		machine = instance
	}

	err := c.client.Instances().Stop(context.Background(), &tcc.StopInstanceInput{
		InstanceID: machine.ID,
	})
	if err != nil {
		return nil, err
	}

	return machine, nil
}

func (c *AgentComputeClient) GetInstanceList() ([]*tcc.Instance, error) {
	params := &tcc.ListInstancesInput{}

	name := config.GetMachineName()
	if name != "" {
		params.Name = name
	}

	tags := config.GetSearchTags()
	if len(tags) > 0 {
		params.Tags = tags
	}

	state := config.GetMachineState()
	if state != "" {
		params.State = state
	}

	brand := config.GetMachineBrand()
	if brand != "" {
		params.Brand = brand
	}

	instances, err := c.client.Instances().List(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return sortInstances(instances), nil

}

func (c *AgentComputeClient) CountInstanceList() (int, error) {
	params := &tcc.ListInstancesInput{}

	name := config.GetMachineName()
	if name != "" {
		params.Name = name
	}

	tags := config.GetSearchTags()
	if len(tags) > 0 {
		params.Tags = tags
	}

	state := config.GetMachineState()
	if state != "" {
		params.State = state
	}

	brand := config.GetMachineBrand()
	if brand != "" {
		params.Brand = brand
	}

	instances, err := c.client.Instances().Count(context.Background(), params)
	if err != nil {
		return -1, err
	}

	return instances, nil

}

func (c *AgentComputeClient) GetInstance() (*tcc.Instance, error) {

	id := config.GetMachineID()
	if id != "" {
		instance, err := c.getInstanceByID(id)
		if err != nil {
			return nil, err
		}

		return instance, nil
	}

	name := config.GetMachineName()
	if name != "" {
		instance, err := c.getInstanceByName(name)
		if err != nil {
			return nil, err
		}

		return instance, nil
	}

	return nil, nil
}

func (c *AgentComputeClient) CreateInstance() (*tcc.Instance, error) {
	params := &tcc.CreateInstanceInput{
		Name:            config.GetMachineName(),
		FirewallEnabled: config.GetMachineFirewall(),
	}

	md := make(map[string]string, 0)

	userdata := config.GetMachineUserdata()
	if userdata != "" {
		md["user-data"] = userdata
	}

	networks := config.GetMachineNetworks()
	if len(networks) > 0 {
		params.Networks = networks
	}

	affinityRules := config.GetMachineAffinityRules()
	if len(affinityRules) > 0 {
		params.Affinity = affinityRules
	}

	tags := config.GetMachineTags()
	if tags != nil {
		params.Tags = tags
	}

	metadata := config.GetMachineMetadata()
	if metadata != nil {
		mergo.Merge(&md, metadata)
	}

	if len(md) > 0 {
		params.Metadata = md
	}

	pkgID := config.GetPkgID()
	if pkgID != "" {
		params.Package = pkgID
	} else {
		packages, err := c.GetPackagesList()
		if err != nil {
			return nil, err
		}

		for _, pkg := range packages {
			if pkg.Name == config.GetPkgName() {
				params.Package = pkg.ID
				break
			}
		}
	}

	imgID := config.GetImgID()
	if imgID != "" {
		params.Image = imgID
	} else {
		images, err := c.GetImagesList()
		if err != nil {
			return nil, err
		}

		for _, img := range images {
			if img.Name == config.GetImgName() {
				params.Image = img.ID
				break
			}
		}
	}

	machine, err := c.client.Instances().Create(context.Background(), params)
	if err != nil {
		return nil, err
	}

	if config.IsBlockingAction() {
		state := make(chan *tcc.Instance, 1)
		go func(machineID string) {
			for {
				time.Sleep(1 * time.Second)
				instance, err := c.getInstanceByID(machineID)
				if err != nil {
					panic(err)
				}
				if instance.State == "running" {
					state <- instance
				}
			}
		}(machine.ID)
	}

	return machine, nil
}

func (c *AgentComputeClient) getInstanceByName(instanceName string) (*tcc.Instance, error) {
	instances, err := c.client.Instances().List(context.Background(), &tcc.ListInstancesInput{
		Name: instanceName,
	})
	if err != nil {
		if terrors.IsSpecificStatusCode(err, http.StatusNotFound) || terrors.IsSpecificStatusCode(err, http.StatusGone) {
			return nil, errors.New("Instance not found")
		}
		return nil, err
	}

	if len(instances) == 0 {
		return nil, errors.New("No instance(s) found")
	}

	return instances[0], nil
}

func (c *AgentComputeClient) getInstanceByID(instanceID string) (*tcc.Instance, error) {
	instance, err := c.client.Instances().Get(context.Background(), &tcc.GetInstanceInput{
		ID: instanceID,
	})
	if err != nil {
		if terrors.IsSpecificStatusCode(err, http.StatusNotFound) || terrors.IsSpecificStatusCode(err, http.StatusGone) {
			return nil, errors.New("Instance not found")
		}
		return nil, err
	}

	return instance, nil
}

func (c *AgentComputeClient) getPackageByName(packageName string) (*tcc.Package, error) {
	pkgs, err := c.client.Packages().List(context.Background(), &tcc.ListPackagesInput{
		Name: packageName,
	})
	if err != nil {
		if terrors.IsSpecificStatusCode(err, http.StatusNotFound) || terrors.IsSpecificStatusCode(err, http.StatusGone) {
			return nil, errors.New("Package not found")
		}
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, errors.New("No package(s) found")
	}

	return pkgs[0], nil
}

func (c *AgentComputeClient) getPackageByID(packageID string) (*tcc.Package, error) {
	pkg, err := c.client.Packages().Get(context.Background(), &tcc.GetPackageInput{
		ID: packageID,
	})
	if err != nil {
		if terrors.IsSpecificStatusCode(err, http.StatusNotFound) || terrors.IsSpecificStatusCode(err, http.StatusGone) {
			return nil, errors.New("Package not found")
		}
		return nil, err
	}

	return pkg, nil
}

func (c *AgentComputeClient) FormatImageName(images []*tcc.Image, imgID string) string {
	for _, img := range images {
		if img.ID == imgID {
			return fmt.Sprintf("%s@%s", img.Name, img.Version)
		}
	}

	return string(imgID[:8])
}
