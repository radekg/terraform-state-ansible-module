//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package cmd

import (
	"os"

	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/joyent/triton-go/cmd/internal/logger"
	"github.com/joyent/triton-go/cmd/manta/cmd/docs"
	"github.com/joyent/triton-go/cmd/manta/cmd/list"
	"github.com/joyent/triton-go/cmd/manta/cmd/shell"
	"github.com/joyent/triton-go/cmd/manta/cmd/version"
	isatty "github.com/mattn/go-isatty"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var subCommands = []*command.Command{
	version.Cmd,
	list.Cmd,
	docs.Cmd,
	shell.Cmd,
}

var rootCmd = &command.Command{
	Cobra: &cobra.Command{
		Use:   "manta",
		Short: "Joyent Manta CLI and client (https://www.joyent.com/triton)",
	},
	Setup: func(parent *command.Command) error {
		{
			const (
				key         = config.KeyUsePager
				longName    = "use-pager"
				shortName   = "P"
				description = "Use a pager to read the output (defaults to $PAGER, less(1), or more(1))"
			)
			var defaultValue bool
			if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
				defaultValue = true
			}

			flags := parent.Cobra.PersistentFlags()
			flags.BoolP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key          = config.KeyLogLevel
				longOpt      = "log-level"
				shortOpt     = "l"
				defaultValue = "INFO"
				description  = "Change the log level being sent to stdout"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longOpt, shortOpt, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longOpt))
			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key         = config.KeyLogFormat
				longOpt     = "log-format"
				shortOpt    = "F"
				description = `Specify the log format ("auto", "zerolog", or "human")`
			)
			defaultValue := logger.FormatAuto.String()

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longOpt, shortOpt, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longOpt))
			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key         = config.KeyLogTermColor
				longOpt     = "use-color"
				shortOpt    = ""
				description = "Use ASCII colors"
			)
			defaultValue := false
			if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
				defaultValue = true
			}

			flags := parent.Cobra.PersistentFlags()
			flags.BoolP(longOpt, shortOpt, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longOpt))
			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key          = config.KeyUseUTC
				longName     = "utc"
				shortName    = "Z"
				defaultValue = false
				description  = "Display times in UTC"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.BoolP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
			viper.SetDefault(key, defaultValue)
		}

		{
			const (
				key          = config.KeyMantaAccount
				longName     = "account"
				shortName    = "A"
				defaultValue = ""
				description  = "Account (login name). If not specified, the environment variable MANTA_USER will be used"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyMantaURL
				longName     = "url"
				shortName    = "U"
				defaultValue = ""
				description  = "Manta URL. If not specified, the environment variable MANTA_URL will be used"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyTritonSSHKeyID
				longName     = "key-id"
				shortName    = "K"
				defaultValue = ""
				description  = "This is the fingerprint of the public key matching the key specified in key_path. It can be obtained via the command ssh-keygen -l -E md5 -f /path/to/key. It can be provided via the SDC_KEY_ID or TRITON_KEY_ID environment variables."
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyTritonSSHKeyMaterial
				longName     = "key-material"
				defaultValue = ""
				description  = "This is the private key of an SSH key associated with the Triton account to be used. If this is not set, the private key corresponding to the fingerprint in key_id must be available via an SSH Agent. It can be provided via the SDC_KEY_MATERIAL or TRITON_KEY_MATERIAL environment variables."
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}

func Execute() error {

	rootCmd.Setup(rootCmd)

	conswriter.UsePager(viper.GetBool(config.KeyUsePager))

	if err := logger.Setup(); err != nil {
		return err
	}

	for _, cmd := range subCommands {
		rootCmd.Cobra.AddCommand(cmd.Cobra)
		cmd.Setup(cmd)
	}

	if err := rootCmd.Cobra.Execute(); err != nil {
		return err
	}

	//switch argv[0] {
	//case "mls":
	//	return list.Cmd
	//}

	return nil
}
