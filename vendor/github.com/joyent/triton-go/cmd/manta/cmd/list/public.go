//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package list

import (
	"fmt"

	"github.com/joyent/triton-go/cmd/agent/storage"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.MaximumNArgs(3),
		Use:          "ls",
		Short:        "list directory contents",
		SilenceUsage: true,
		Example: `
$ manta ls
$ manta ls /stor
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cons := conswriter.GetTerminal()

			c, err := cfg.NewMantaConfig()
			if err != nil {
				return err
			}

			s, err := storage.NewStorageClient(c)
			if err != nil {
				return err
			}

			directoryOutput, err := s.GetDirectoryListing(args)
			if err != nil {
				return err
			}

			cons.Write([]byte(fmt.Sprintf("Found %d directory entries", directoryOutput.ResultSetSize)))

			for _, entry := range directoryOutput.Entries {
				cons.Write([]byte(fmt.Sprintf("\n%s/", entry.Name)))
			}

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		return nil
	},
}
