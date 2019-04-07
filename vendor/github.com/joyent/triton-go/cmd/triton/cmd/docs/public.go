//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package docs

import (
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/triton/cmd/docs/man"
	"github.com/joyent/triton-go/cmd/triton/cmd/docs/md"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:     "doc",
		Aliases: []string{"docs", "documentation"},
		Short:   "Documentation for Triton cli",
	},

	Setup: func(parent *command.Command) error {
		cmds := []*command.Command{
			man.Cmd,
			md.Cmd,
		}

		for _, cmd := range cmds {
			cmd.Setup(cmd)
			parent.Cobra.AddCommand(cmd.Cobra)
		}

		return nil
	},
}
