//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package command

import "github.com/spf13/cobra"

type SetupFunc func(parent *Command) error

type Command struct {
	Cobra *cobra.Command
	Setup SetupFunc
}
