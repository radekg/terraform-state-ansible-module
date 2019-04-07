package services_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/services"
	"github.com/joyent/triton-go/testutils"
)

const (
	accountURL  = "testing"
	fakeGroupID = "8b81157f-28c2-4258-85b1-31b36df9c953"
	groupPath   = "/v1/tsg/groups"

	updatedName       = "new-name-1"
	updatedTemplateID = "8de73edb-3436-4d52-b0be-24b1916fabd4"
	updatedCapacity   = 3

	createdName       = "created-group-1"
	createdTemplateID = "288758c2-a4d7-4785-9dda-f4a7c41f0e3b"
	createdCapacity   = 2
)

func MockServicesClient() *services.ServiceGroupClient {
	return &services.ServiceGroupClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountURL,
		}),
	}
}

func TestGetGroup(t *testing.T) {
	servicesClient := MockServicesClient()

	do := func(ctx context.Context, sc *services.ServiceGroupClient) (*services.ServiceGroup, error) {
		defer testutils.DeactivateClient()

		group, err := sc.Groups().Get(ctx, &services.GetGroupInput{
			ID: fakeGroupID,
		})
		if err != nil {
			return nil, err
		}
		return group, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join(groupPath, fakeGroupID), getGroupSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join(groupPath, fakeGroupID), getGroupEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join(groupPath, fakeGroupID), getGroupBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to get service group"
		path := path.Join(groupPath, "not-a-real-group-id")
		testutils.RegisterResponder("GET", path, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		resp, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func TestListGroups(t *testing.T) {
	servicesClient := MockServicesClient()

	do := func(ctx context.Context, sc *services.ServiceGroupClient) ([]*services.ServiceGroup, error) {
		defer testutils.DeactivateClient()

		groups, err := sc.Groups().List(ctx, &services.ListGroupsInput{})
		if err != nil {
			return nil, err
		}
		return groups, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", groupPath, listGroupsSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", groupPath, getGroupEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", groupPath, listGroupsBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to list service groups"
		testutils.RegisterResponder("GET", groupPath, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		resp, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func TestListGroupInstances(t *testing.T) {
	servicesClient := MockServicesClient()
	instancesPath := path.Join(groupPath, fakeGroupID, "instances")

	do := func(ctx context.Context, sc *services.ServiceGroupClient) ([]*compute.Instance, error) {
		defer testutils.DeactivateClient()

		groups, err := sc.Groups().ListInstances(ctx, &services.ListGroupInstancesInput{
			ID: fakeGroupID,
		})
		if err != nil {
			return nil, err
		}
		return groups, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", instancesPath, listGroupInstancesSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", instancesPath, getGroupEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", instancesPath, listGroupInstancesBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to list group instances"
		testutils.RegisterResponder("GET", instancesPath, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		resp, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func TestCreateGroup(t *testing.T) {
	servicesClient := MockServicesClient()

	do := func(ctx context.Context, sc *services.ServiceGroupClient) (*services.ServiceGroup, error) {
		defer testutils.DeactivateClient()

		group, err := sc.Groups().Create(ctx, &services.CreateGroupInput{
			GroupName:  createdName,
			TemplateID: createdTemplateID,
			Capacity:   createdCapacity,
		})
		if err != nil {
			return nil, err
		}
		return group, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", groupPath, createGroupSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
		if resp.GroupName != createdName {
			t.Fatalf("expected group_name to be %q: got %q",
				createdName, resp.GroupName)
		}
		if resp.TemplateID != createdTemplateID {
			t.Fatalf("expected template_id to be %q: got %q",
				createdTemplateID, resp.TemplateID)
		}
		if resp.Capacity != createdCapacity {
			t.Fatalf("expected capacity to be %d: got %d",
				createdCapacity, resp.Capacity)
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("POST", groupPath, getGroupEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("POST", groupPath, getGroupBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to create service group"
		testutils.RegisterResponder("POST", groupPath, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		resp, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func TestUpdateGroup(t *testing.T) {
	servicesClient := MockServicesClient()
	updatePath := path.Join(groupPath, fakeGroupID)

	do := func(ctx context.Context, sc *services.ServiceGroupClient) (*services.ServiceGroup, error) {
		defer testutils.DeactivateClient()

		group, err := sc.Groups().Update(ctx, &services.UpdateGroupInput{
			ID:         fakeGroupID,
			GroupName:  updatedName,
			TemplateID: updatedTemplateID,
			Capacity:   updatedCapacity,
		})
		if err != nil {
			return nil, err
		}
		return group, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("PUT", updatePath, updateGroupSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
		if resp.GroupName != updatedName {
			t.Fatalf("expected group_name to be %q: got %q",
				updatedName, resp.GroupName)
		}
		if resp.TemplateID != updatedTemplateID {
			t.Fatalf("expected template_id to be %q: got %q",
				updatedTemplateID, resp.TemplateID)
		}
		if resp.Capacity != updatedCapacity {
			t.Fatalf("expected capacity to be %d: got %d",
				updatedCapacity, resp.Capacity)
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("PUT", updatePath, getGroupEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("PUT", updatePath, getGroupBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to update service group"
		testutils.RegisterResponder("PUT", updatePath, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		resp, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}
		if resp != nil {
			t.Error("expected resp to be nil")
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func TestDeleteGroup(t *testing.T) {
	servicesClient := MockServicesClient()
	deletePath := path.Join(groupPath, fakeGroupID)

	do := func(ctx context.Context, sc *services.ServiceGroupClient) error {
		defer testutils.DeactivateClient()

		err := sc.Groups().Delete(ctx, &services.DeleteGroupInput{
			ID: fakeGroupID,
		})
		if err != nil {
			return err
		}
		return nil
	}

	doInvalid := func(ctx context.Context, sc *services.ServiceGroupClient) error {
		defer testutils.DeactivateClient()

		err := sc.Groups().Delete(ctx, &services.DeleteGroupInput{})
		if err != nil {
			return err
		}
		return nil
	}

	t.Run("invalid", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", deletePath, deleteGroupSuccess)

		err := doInvalid(context.Background(), servicesClient)
		if err == nil {
			t.Fatal("expected error to not be nil")
		}
		if !strings.Contains(err.Error(), "unable to validate delete group input") {
			t.Errorf("expected error to include message: found %s", err)
		}
	})

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", deletePath, deleteGroupSuccess)

		err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to delete service group"
		testutils.RegisterResponder("DELETE", deletePath, func(req *http.Request) (*http.Response, error) {
			return nil, errors.New(errMesg)
		})

		err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), errMesg) {
			t.Errorf("expected error to include message: found %s", err)
		}
	})
}

func deleteGroupSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func getGroupError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("unable to get service group")
}

func getGroupBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"test-group-1",
  "template_id":"c012d81e-a6d2-4179-ae2d-be6fa8f13a60",
  "capacity":3,
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getGroupEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getGroupSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"test-group-1",
  "template_id":"c012d81e-a6d2-4179-ae2d-be6fa8f13a60",
  "capacity":3,
  "created_at": "2018-04-14T15:24:20.205784Z",
  "updated_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listGroupsBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"test-group-1",
  "template_id":"c012d81e-a6d2-4179-ae2d-be6fa8f13a60",
  "capacity":3
}, {
  "id":"a981d5d1-a105-4ff7-a201-e535eed1d295",
  "group_name":"test-group-1",
  "template_id":"bfdf52c2-67ac-4066-a7a1-73ed13966b22",
  "capacity":3,
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listGroupsSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"test-group-1",
  "template_id":"c012d81e-a6d2-4179-ae2d-be6fa8f13a60",
  "capacity":3,
  "created_at": "2018-04-14T15:24:20.205784Z",
  "updated_at": "2018-04-14T15:24:20.205784Z"
}, {
  "id":"a981d5d1-a105-4ff7-a201-e535eed1d295",
  "group_name":"test-group-1",
  "template_id":"bfdf52c2-67ac-4066-a7a1-73ed13966b22",
  "capacity":3,
  "created_at": "2018-04-14T15:24:20.205784Z",
  "updated_at": "2018-04-14T15:24:20.205784Z"
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listGroupInstancesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
  {
	"id": "af3f16f5-7bbe-45b6-92c4-734c14c36da5",
	"name": "test-1",
	"type": "smartmachine",
	"brand": "joyent",
	"state": "running",
	"image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
	"ips": [
	  "10.88.88.25",
	  "192.168.128.4"
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
	"primaryIp": "10.88.88.25",
	"firewall_enabled": false,
	"compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
	"package": "sdc_128",
  }, {
	"id": "176eadda-af72-4b8f-9356-9e7557576e48",
	"name": "test-2",
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

func listGroupInstancesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[
  {
	"id": "af3f16f5-7bbe-45b6-92c4-734c14c36da5",
	"name": "test-1",
	"type": "smartmachine",
	"brand": "joyent",
	"state": "running",
	"image": "2b683a82-a066-11e3-97ab-2faa44701c5a",
	"ips": [
	  "10.88.88.25",
	  "192.168.128.4"
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
	"primaryIp": "10.88.88.25",
	"firewall_enabled": false,
	"compute_node": "564d0b8e-6099-7648-351e-877faf6c56f6",
	"package": "sdc_128"
  }, {
	"id": "176eadda-af72-4b8f-9356-9e7557576e48",
	"name": "test-2",
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

func createGroupSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"created-group-1",
  "template_id":"288758c2-a4d7-4785-9dda-f4a7c41f0e3b",
  "capacity":2,
  "created_at": "2018-04-14T15:24:20.205784Z",
  "updated_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateGroupSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "group_name":"new-name-1",
  "template_id":"8de73edb-3436-4d52-b0be-24b1916fabd4",
  "capacity":3,
  "created_at": "2018-04-14T15:24:20.205784Z",
  "updated_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}
