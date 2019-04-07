package services_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/joyent/triton-go/services"
	"github.com/joyent/triton-go/testutils"
)

const (
	fakeTemplateID = "8b81157f-28c2-4258-85b1-31b36df9c953"
	templatePath   = "/v1/tsg/templates"
)

func TestGetTemplate(t *testing.T) {
	servicesClient := MockServicesClient()
	getPath := path.Join(templatePath, fakeGroupID)

	do := func(ctx context.Context, sc *services.ServiceGroupClient) (*services.InstanceTemplate, error) {
		defer testutils.DeactivateClient()

		template, err := sc.Templates().Get(ctx, &services.GetTemplateInput{
			ID: fakeTemplateID,
		})
		if err != nil {
			return nil, err
		}
		return template, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", getPath, getTemplateSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", getPath, getEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", getPath, getTemplateBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to get instance template"
		testutils.RegisterResponder("GET", getPath, func(req *http.Request) (*http.Response, error) {
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

func TestListTemplates(t *testing.T) {
	servicesClient := MockServicesClient()

	do := func(ctx context.Context, sc *services.ServiceGroupClient) ([]*services.InstanceTemplate, error) {
		defer testutils.DeactivateClient()

		templates, err := sc.Templates().List(ctx, &services.ListTemplatesInput{})
		if err != nil {
			return nil, err
		}
		return templates, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", templatePath, listTemplatesSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", templatePath, getEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", templatePath, listTemplatesBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to list instance templates"
		testutils.RegisterResponder("GET", templatePath, func(req *http.Request) (*http.Response, error) {
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

func TestCreateTemplate(t *testing.T) {
	servicesClient := MockServicesClient()

	input := &services.CreateTemplateInput{
		TemplateName:    "test-template-1",
		Package:         "g4-highcpu-1G",
		ImageID:         "c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
		Networks:        []string{"346f321b-df77-44df-9d44-5b2bda4d5405"},
		Userdata:        "bash script here",
		Metadata:        map[string]string{"metadata": "test"},
		Tags:            map[string]string{"tag": "test"},
		FirewallEnabled: true,
	}

	do := func(ctx context.Context, sc *services.ServiceGroupClient) (*services.InstanceTemplate, error) {
		defer testutils.DeactivateClient()

		template, err := sc.Templates().Create(ctx, input)
		if err != nil {
			return nil, err
		}
		return template, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("POST", templatePath, createTemplateSuccess)

		resp, err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatalf("expected output but got nil")
		}
		if resp.TemplateName != input.TemplateName {
			t.Fatalf("expected template_name to be %q: got %q",
				input.TemplateName, resp.TemplateName)
		}
		if resp.Package != input.Package {
			t.Fatalf("expected package to be %q: got %q",
				input.Package, resp.Package)
		}
		if resp.ImageID != input.ImageID {
			t.Fatalf("expected image_id to be %q: got %q",
				input.ImageID, resp.ImageID)
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("POST", templatePath, getEmpty)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("POST", templatePath, getTemplateBadDecode)

		_, err := do(context.Background(), servicesClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to create instance template"
		testutils.RegisterResponder("POST", templatePath, func(req *http.Request) (*http.Response, error) {
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

func TestDeleteTemplate(t *testing.T) {
	servicesClient := MockServicesClient()
	deletePath := path.Join(templatePath, fakeTemplateID)

	do := func(ctx context.Context, sc *services.ServiceGroupClient) error {
		defer testutils.DeactivateClient()

		err := sc.Templates().Delete(ctx, &services.DeleteTemplateInput{
			ID: fakeTemplateID,
		})
		if err != nil {
			return err
		}
		return nil
	}

	doInvalid := func(ctx context.Context, sc *services.ServiceGroupClient) error {
		defer testutils.DeactivateClient()

		err := sc.Templates().Delete(ctx, &services.DeleteTemplateInput{})
		if err != nil {
			return err
		}
		return nil
	}

	t.Run("invalid", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", deletePath, deleteTemplateSuccess)

		err := doInvalid(context.Background(), servicesClient)
		if err == nil {
			t.Fatal("expected error to not be nil")
		}
		if !strings.Contains(err.Error(), "unable to validate delete template input") {
			t.Errorf("expected error to include message: found %s", err)
		}
	})

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("DELETE", deletePath, deleteTemplateSuccess)

		err := do(context.Background(), servicesClient)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error", func(t *testing.T) {
		errMesg := "unable to delete instance template"
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

func deleteTemplateSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     header,
	}, nil
}

func getTemplateBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func getEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func getTemplateSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
  "created_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listTemplatesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"}
}, {
  "id":"cb2d0434-e4c4-4ba0-966e-ff5c17d3378a",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listTemplatesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
  "created_at": "2018-04-14T15:24:20.205784Z"
}, {
  "id":"cb2d0434-e4c4-4ba0-966e-ff5c17d3378a",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
  "created_at": "2018-04-14T15:24:20.205784Z"
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func createTemplateSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
  "created_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func updateTemplateSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "id":"8b81157f-28c2-4258-85b1-31b36df9c953",
  "template_name":"test-template-1",
  "package":"g4-highcpu-1G",
  "image_id":"c10a0bc4-ffe0-4ca9-a469-62b7e37990bb",
  "firewall_enabled":true,
  "networks":["346f321b-df77-44df-9d44-5b2bda4d5405"],
  "userdata":"bash script here",
  "metadata":{"hello":"again"},
  "tags":{"admin":"cheap","foo":"bar"},
  "created_at": "2018-04-14T15:24:20.205784Z"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}
