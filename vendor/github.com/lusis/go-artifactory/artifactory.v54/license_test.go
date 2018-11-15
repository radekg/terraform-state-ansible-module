package artifactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLicenseInfo(t *testing.T) {
	responseFile, err := os.Open("assets/test/license_information.json")
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
	info, err := client.GetLicenseInfo()
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, "Trial", info.LicenseType, "Should have the Trial license type")
	assert.Equal(t, "Sep 30, 2015", info.ValidThrough, "Should be valid through Sep 30th, 2015")
	assert.Equal(t, "User Bob", info.LicensedTo, "Should be licensed to User Bob")
}

func TestInstallLicense(t *testing.T) {
	var buf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		req, _ := ioutil.ReadAll(r.Body)
		buf.Write(req)
		fmt.Fprintf(w, "")
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

	license := InstallLicense{
		LicenseKey: "179b7ea384d0c4655a00dfac7285a21d986a17923",
	}

	expectedJSON, _ := json.Marshal(license)
	err := client.InstallLicense(license, make(map[string]string))
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, string(expectedJSON), buf.String(), "should send license json")
}

func TestGetHALicenseInfo(t *testing.T) {
	responseFile, err := os.Open("assets/test/ha_license_information.json")
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
	info, err := client.GetHALicenseInfo()
	assert.NoError(t, err, "should not return an error")
	assert.Len(t, info.Licenses, 2, "Should have 2 licenses")
	assert.Equal(t, "Enterprise", info.Licenses[0].LicenseType, "Should have the Enterprise license type")
	assert.Equal(t, "May 15, 2018", info.Licenses[0].ValidThrough, "Should be valid through May 15, 2018")
	assert.Equal(t, "JFrog", info.Licenses[0].LicensedTo, "Should be licensed to JFrog")
	assert.Equal(t, "179b7ea384d0c4655a00dfac7285a21d986a17923", info.Licenses[0].LicenseHash, "Should have 179b7ea384d0c4655a00dfac7285a21d986a17923 license hash")
	assert.Equal(t, "art1", info.Licenses[0].NodeID, "Should have art1 node id")
	assert.Equal(t, "http://10.1.16.83:8091/artifactory", info.Licenses[0].NodeURL, "Should have http://10.1.16.83:8091/artifactory node url")
	assert.Equal(t, false, info.Licenses[0].Expired, "Should not be expired")
}

func TestInstallHALicense(t *testing.T) {
	var buf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		req, _ := ioutil.ReadAll(r.Body)
		buf.Write(req)
		fmt.Fprintf(w, "")
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

	licenses := []InstallLicense{
		{LicenseKey: "179b7ea384d0c4655a00dfac7285a21d986a17923"},
		{LicenseKey: "e10b8aa1d1dc5107439ce43debc6e65dfeb71afd3"},
	}

	expectedJSON, _ := json.Marshal(licenses)
	err := client.InstallHALicenses(licenses, make(map[string]string))
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, string(expectedJSON), buf.String(), "should send license json")
}
