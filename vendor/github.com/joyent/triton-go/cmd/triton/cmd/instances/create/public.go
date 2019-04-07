//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package create

import (
	"errors"
	"fmt"

	"github.com/joyent/triton-go/cmd/agent/compute"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:          "create",
		Short:        "create a triton instance",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cfg.GetPkgName() == "" && cfg.GetPkgID() == "" {
				return errors.New("Either `pkg-name` or `pkg-id` must be specified for Create Instance")
			}

			if cfg.GetImgName() == "" && cfg.GetImgID() == "" {
				return errors.New("Either `img-name` or `img-id` must be specified for Create Instance")
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

			instance, err := a.CreateInstance()
			if err != nil {
				return err
			}

			cons.Write([]byte(fmt.Sprintf("Created instance %q (%s)", instance.Name, instance.ID)))

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		{
			parent.Cobra.MarkFlagRequired("name")
		}

		{
			const (
				key          = config.KeyInstanceWait
				longName     = "wait"
				shortName    = "w"
				defaultValue = false
				description  = "Block until instance state indicates the action is complete. (defaults to false)"
			)

			flags := parent.Cobra.Flags()
			flags.BoolP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key          = config.KeyPackageID
				longName     = "pkg-id"
				defaultValue = ""
				description  = "Package id (defaults to ''). This takes precedence over 'pkg-name'"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageName
				longName     = "pkg-name"
				defaultValue = ""
				description  = "Package name (defaults to '')"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyImageId
				longName     = "img-id"
				defaultValue = ""
				description  = "Image id (defaults to ''). This takes precedence over 'img-name'"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyImageName
				longName     = "img-name"
				defaultValue = ""
				description  = "Image name (defaults to '')"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceFirewall
				longName     = "firewall"
				defaultValue = false
				description  = "Enable Cloud Firewall on this instance (defaults to false)"
			)

			flags := parent.Cobra.Flags()
			flags.Bool(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))

			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key         = config.KeyInstanceNetwork
				longName    = "networks"
				shortName   = "N"
				description = "One or more comma-separated networks IDs. This option can be used multiple times."
			)

			flags := parent.Cobra.Flags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key         = config.KeyInstanceMetadata
				longName    = "metadata"
				shortName   = "m"
				description = `Add metadata when creating the instance. Metadata are key/value
			       pairs available on the instance API object as the "metadata"
			       field, and inside the instance via the "mdata-*" commands. DATA
			       is one of: a "key=value" string (bool and numeric "value" are
				   converted to that type). This option can be used multiple times.`
			)

			flags := parent.Cobra.Flags()
			flags.StringSliceP(longName, shortName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key         = config.KeyInstanceAffinityRule
				longName    = "affinity"
				description = `Affinity rules for selecting a server for this instance. Rules
have one of the following forms: "instance==INST" (the new
instance must be on the same server as INST), "instance!=INST"
(new inst must *not* be on the same server as INST),
"instance==~INST"" (*attempt* to place on the same server as
INST), or "instance!=~INST" (*attempt* to place on a server
other than INST's). "INST" is an existing instance name or id.
There are two shortcuts: "inst" may be used instead of
"instance" and "instance==INST" can be shortened to just "INST".
This option can be used multiple times.`
			)

			flags := parent.Cobra.Flags()
			flags.StringSlice(longName, nil, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyInstanceUserdata
				longName     = "userdata"
				defaultValue = ""
				description  = "A custom script which will be executed by the instance right after creation, and on every instance reboot."
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}
