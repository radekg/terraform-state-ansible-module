//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package errors

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
)

func TestCheckIsSpecificError(t *testing.T) {
	t.Run("API error", func(t *testing.T) {
		err := &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "ResourceNotFound",
			Message:    "Resource Not Found", // note dosesn't matter
		}

		if !IsSpecificError(err, "ResourceNotFound") {
			t.Fatalf("Expected `ResourceNotFound`, got %v", err.Code)
		}

		if IsSpecificError(err, "IncorrectCode") {
			t.Fatalf("Expected `IncorrectCode`, got %v", err.Code)
		}
	})

	t.Run("Non Specific Error Type", func(t *testing.T) {
		err := errors.New("This is a new error")

		if IsSpecificError(err, "ResourceNotFound") {
			t.Fatalf("Specific Error Type Found")
		}
	})
}

func TestCheckIsSpecificStatusCode(t *testing.T) {
	t.Run("API error", func(t *testing.T) {
		err := &APIError{
			StatusCode: http.StatusNotFound,
			Code:       "ResourceNotFound",
			Message:    "Resource Not Found", // note dosesn't matter
		}

		if !IsSpecificStatusCode(err, http.StatusNotFound) {
			t.Fatalf("Expected `404`, got %v", err.StatusCode)
		}

		if IsSpecificStatusCode(err, http.StatusNoContent) {
			t.Fatalf("Expected `404`, got %v", err.Code)
		}
	})

	t.Run("Non Specific Error Type", func(t *testing.T) {
		err := errors.New("This is a new error")

		if IsSpecificStatusCode(err, http.StatusNotFound) {
			t.Fatalf("Specific Error Type Found")
		}
	})
}
