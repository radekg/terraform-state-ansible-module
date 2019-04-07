//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package instances

import (
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/count"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/create"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/delete"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/get"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/ip"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/list"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/reboot"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/start"
	"github.com/joyent/triton-go/cmd/triton/cmd/instances/stop"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:     "instances",
		Aliases: []string{"instance", "vms", "machines"},
		Short:   "Instances (aka VMs/Machines/Containers)",
	},

	Setup: func(parent *command.Command) error {

		cmds := []*command.Command{
			list.Cmd,
			create.Cmd,
			delete.Cmd,
			get.Cmd,
			count.Cmd,
			reboot.Cmd,
			start.Cmd,
			stop.Cmd,
			ip.Cmd,
		}

		for _, cmd := range cmds {
			cmd.Setup(cmd)
			parent.Cobra.AddCommand(cmd.Cobra)
		}

		{
			const (
				key          = config.KeyInstanceID
				longName     = "id"
				defaultValue = ""
				description  = "Instance ID"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceName
				longName     = "name"
				shortName    = "n"
				defaultValue = ""
				description  = "Instance Name"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key         = config.KeyInstanceTag
				longName    = "tag"
				shortName   = "t"
				description = "Instance Tags. This flag can be used multiple times"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			flags := parent.Cobra.PersistentFlags()
			flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				switch name {
				case "tag":
					name = "tags"
					break
				}

				return pflag.NormalizedName(name)
			})
		}

		{
			const (
				key          = config.KeyInstanceState
				longName     = "state"
				defaultValue = ""
				description  = "Instance state (e.g. running)"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceBrand
				longName     = "brand"
				defaultValue = ""
				description  = "Instance brand (e.g. lx, kvm)"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}
