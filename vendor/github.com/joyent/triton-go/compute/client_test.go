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
)

// MockComputeClient is used to mock out compute.ComputeClient for all tests
// under the triton-go/compute package
func MockComputeClient() *compute.ComputeClient {
	return &compute.ComputeClient{
		Client: testutils.NewMockClient(testutils.MockClientInput{
			AccountName: accountURL,
		}),
	}
}

const (
	testHeaderName = "X-Test-Header"
	testHeaderVal1 = "number one"
	testHeaderVal2 = "number two"
)

func TestSetHeader(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) error {
		defer testutils.DeactivateClient()

		header := &http.Header{}
		header.Add(testHeaderName, testHeaderVal1)
		header.Add(testHeaderName, testHeaderVal2)
		cc.SetHeader(header)

		_, err := cc.Datacenters().List(ctx, &compute.ListDataCentersInput{})

		return err
	}

	t.Run("override header", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "datacenters"), overrideHeaderTest(t))

		err := do(context.Background(), computeClient)
		if err != nil {
			t.Error(err)
		}
	})
}

func overrideHeaderTest(t *testing.T) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		// test existence of custom headers at all
		if req.Header.Get(testHeaderName) == "" {
			t.Errorf("request header should contain '%s'", testHeaderName)
		}
		testHeader := strings.Join(req.Header[testHeaderName], ",")
		// test override of initial header
		if !strings.Contains(testHeader, testHeaderVal1) {
			t.Errorf("request header should not contain %q: got %q", testHeaderVal1, testHeader)
		}
		if strings.Contains(testHeader, testHeaderVal2) {
			t.Errorf("request header should contain '%s': got '%s'", testHeaderVal2, testHeader)
		}

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
}
