//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package identity

import (
	"github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/identity"
	"github.com/pkg/errors"
)

func NewGetIdentityClient(cfg *config.TritonClientConfig) (*identity.IdentityClient, error) {
	identityClient, err := identity.NewClient(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "Error Creating Triton Identity Client")
	}
	return identityClient, nil
}
