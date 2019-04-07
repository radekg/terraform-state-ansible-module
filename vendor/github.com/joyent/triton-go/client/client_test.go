//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package client

import (
	"os"
	"strings"
	"testing"

	auth "github.com/joyent/triton-go/authentication"
)

const BadURL = "**ftp://man($$"

func TestNew(t *testing.T) {
	mantaURL := "https://us-east.manta.joyent.com"
	tsgEnv := "http://tsg.test.org"
	jpcTritonURL := "https://us-east-1.api.joyent.com"
	spcTritonURL := "https://us-east-1.api.samsungcloud.io"
	jpcServiceURL := "https://tsg.us-east-1.svc.joyent.zone"
	spcServiceURL := "https://tsg.us-east-1.svc.samsungcloud.zone"
	privateInstallUrl := "https://myinstall.mycompany.com"

	accountName := "test.user"
	signer, _ := auth.NewTestSigner()

	tests := []struct {
		name        string
		tritonURL   string
		mantaURL    string
		tsgEnv      string
		servicesURL string
		accountName string
		signer      auth.Signer
		err         interface{}
	}{
		{"default", jpcTritonURL, mantaURL, "", jpcServiceURL, accountName, signer, nil},
		{"in samsung", spcTritonURL, mantaURL, "", spcServiceURL, accountName, signer, nil},
		{"env TSG", jpcTritonURL, mantaURL, tsgEnv, tsgEnv, accountName, signer, nil},
		{"missing url", "", "", "", "", accountName, signer, ErrMissingURL},
		{"bad tritonURL", BadURL, mantaURL, "", "", accountName, signer, InvalidTritonURL},
		{"bad mantaURL", jpcTritonURL, BadURL, "", jpcServiceURL, accountName, signer, InvalidMantaURL},
		{"bad TSG", jpcTritonURL, mantaURL, BadURL, "", accountName, signer, InvalidServicesURL},
		{"missing accountName", jpcTritonURL, mantaURL, "", jpcServiceURL, "", signer, ErrAccountName},
		{"missing signer", jpcTritonURL, mantaURL, "", jpcServiceURL, accountName, nil, ErrDefaultAuth},
		{"private install", privateInstallUrl, mantaURL, "", "", accountName, signer, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv("TRITON_KEY_ID")
			os.Unsetenv("SDC_KEY_ID")
			os.Unsetenv("MANTA_KEY_ID")
			os.Unsetenv("SSH_AUTH_SOCK")
			os.Unsetenv("TRITON_TSG_URL")

			if test.tsgEnv != "" {
				os.Setenv("TRITON_TSG_URL", test.tsgEnv)
			}

			c, err := New(
				test.tritonURL,
				test.mantaURL,
				test.accountName,
				test.signer,
			)

			// test generation of TSG URL for all non-error cases
			if err == nil {
				if c.ServicesURL.String() != test.servicesURL {
					t.Errorf("expected ServicesURL to be set to %q: got %q (%s)",
						test.servicesURL, c.ServicesURL.String(), test.name)
					return
				}
			}

			if test.err != nil {
				if err == nil {
					t.Error("expected error not to be nil")
					return
				}

				switch test.err.(type) {
				case error:
					testErr := test.err.(error)
					if err.Error() != testErr.Error() {
						t.Errorf("expected error: received %v", err)
					}
				case string:
					testErr := test.err.(string)
					if !strings.Contains(err.Error(), testErr) {
						t.Errorf("expected error: received %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Errorf("expected error to be nil: received %v", err)
			}
		})
	}

	t.Run("default SSH agent auth", func(t *testing.T) {
		os.Unsetenv("SSH_AUTH_SOCK")
		err := os.Setenv("TRITON_KEY_ID", auth.Dummy.Fingerprint)
		defer os.Unsetenv("TRITON_KEY_ID")
		if err != nil {
			t.Errorf("expected error to not be nil: received %v", err)
		}

		_, err = New(
			jpcTritonURL,
			mantaURL,
			accountName,
			nil,
		)
		if err == nil {
			t.Error("expected error to not be nil")
		}
		if !strings.Contains(err.Error(), "unable to initialize NewSSHAgentSigner") {
			t.Errorf("expected error to be from NewSSHAgentSigner: received '%v'", err)
		}
	})
}
