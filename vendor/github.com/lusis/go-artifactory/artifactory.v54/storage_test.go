package artifactory

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileList(t *testing.T) {
	responseFile, err := os.Open("assets/test/file_list.json")
	if err != nil {
		t.Fatalf("Unable to read test data: %s", err.Error())
	}
	defer func() { _ = responseFile.Close() }()
	responseBody, _ := ioutil.ReadAll(responseFile)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(responseBody))
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	fileList, err := client.GetFileList("libs-release-local", "org/acme")
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, "http://localhost:8081/artifactory/api/storage/libs-release-local/org/acme", fileList.URI, "should have expected uri")
	assert.Equal(t, "ISO8601", fileList.Created, "created should be ISO8601")
	assert.Len(t, fileList.Files, 3, "should have three files")
	assert.Equal(t, "/archived", fileList.Files[0].URI, "first file uri should be /archived")
	assert.Equal(t, -1, fileList.Files[0].Size, "first file size should be -1")
	assert.Equal(t, "ISO8601", fileList.Files[0].LastModified, "first file last modified should be ISO8601")
	assert.Equal(t, true, fileList.Files[0].Folder, "first file folder should be true")
	for _, file := range fileList.Files {
		assert.NotNil(t, file.URI, "Uri should not be empty")
		assert.NotNil(t, file.Size, "Size should not be empty")
		assert.NotNil(t, file.LastModified, "LastModified should not be empty")
		assert.NotNil(t, file.Folder, "Folder should not be empty")
	}
}

func TestGetItemProperties(t *testing.T) {
	responseFile, err := os.Open("assets/test/item_properties.json")
	if err != nil {
		t.Fatalf("Unable to read test data: %s", err.Error())
	}
	defer func() { _ = responseFile.Close() }()
	responseBody, _ := ioutil.ReadAll(responseFile)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(responseBody))
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	itemProps, err := client.GetItemProperties("libs-release-local", "org/acme")
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, "http://localhost:8081/artifactory/api/storage/libs-release-local/org/acme", itemProps.URI, "should have expected uri")
	assert.Equal(t, []string{"v1", "v2", "v3"}, itemProps.Properties["p1"], "p1 item property should have array of values equal to [v1, v2, v3]")
	for prop, values := range itemProps.Properties {
		assert.NotNil(t, prop, "Prop should not be empty")
		assert.NotNil(t, values, "Values should not be empty")
	}
}

func TestSetItemProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)

	properties := make(map[string][]string)
	properties["p1"] = []string{"v1"}
	properties["p2"] = []string{"v1", "v2", "v3"}

	err := client.SetItemProperties("libs-release-local", "org/acme", properties)
	assert.NoError(t, err, "should not return an error")
}

func TestDeleteItemProperties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)

	properties := []string{"v1"}

	err := client.DeleteItemProperties("libs-release-local", "org/acme", properties)
	assert.NoError(t, err, "should not return an error")
}
