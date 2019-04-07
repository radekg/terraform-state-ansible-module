//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package packages

import (
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/joyent/triton-go/cmd/triton/cmd/packages/get"
	"github.com/joyent/triton-go/cmd/triton/cmd/packages/list"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:     "packages",
		Aliases: []string{"package", "pkgs"},
		Short:   "List and get Triton packages.",
		Long: `A package is a collection of attributes -- for example disk quota,
			   amount of RAM -- used when creating an instance. They have a name
			   and ID for identification.`,
	},

	Setup: func(parent *command.Command) error {

		cmds := []*command.Command{
			list.Cmd,
			get.Cmd,
		}

		for _, cmd := range cmds {
			cmd.Setup(cmd)
			parent.Cobra.AddCommand(cmd.Cobra)
		}

		{
			const (
				key          = config.KeyPackageID
				longName     = "id"
				defaultValue = ""
				description  = "Package ID"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageName
				longName     = "name"
				defaultValue = ""
				description  = "Package Name"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}
