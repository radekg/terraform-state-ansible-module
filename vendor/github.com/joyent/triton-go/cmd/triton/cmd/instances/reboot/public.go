//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package reboot

import (
	"errors"
	"fmt"

	"github.com/joyent/triton-go/cmd/agent/compute"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "reboot",
		Short:        "reboot instance",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cfg.GetMachineID() == "" && cfg.GetMachineName() == "" {
				return errors.New("Either `id` or `name` must be specified")
			}

			if cfg.GetMachineID() != "" && cfg.GetMachineName() != "" {
				return errors.New("Only 1 of `id` or `name` must be specified")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cons := conswriter.GetTerminal()

			c, err := cfg.NewTritonConfig()
			if err != nil {
				return err
			}

			a, err := compute.NewComputeClient(c)
			if err != nil {
				return err
			}

			instance, err := a.RebootInstance()
			if err != nil {
				return err
			}

			cons.Write([]byte(fmt.Sprintf("Rebooted instance %q", instance.Name)))

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		return nil
	},
}
