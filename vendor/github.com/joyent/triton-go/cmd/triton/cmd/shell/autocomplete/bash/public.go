//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package bash

import (
	"fmt"

	"os"

	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:   "bash",
		Short: "Generates shell autocompletion file for Triton",
		Long: `Generates a shell autocompletion script for Triton.

By default, the file is written directly to /etc/bash_completion.d
for convenience, and the command may need superuser rights, e.g.:

	$ sudo triton shell autocomplete bash

Add ` + "`--bash-autocomplete-dir=/path/to/file`" + ` flag to set alternative
folder location.

Logout and in again to reload the completion scripts,
or just source them in directly:

	$ . /etc/bash_completion`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := viper.GetString(config.KeyBashAutoCompletionTarget)
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, 0777); err != nil {
					return errors.Wrapf(err, "unable to make bash-autocomplete-target %q", target)
				}
			}
			bashFile := fmt.Sprintf("%s/triton.sh", target)
			err := cmd.Root().GenBashCompletionFile(bashFile)
			if err != nil {
				return err
			}

			log.Info().Msg("Installation completed successfully.")

			return nil
		},
	},
	Setup: func(parent *command.Command) error {

		{
			const (
				key          = config.KeyBashAutoCompletionTarget
				longName     = "bash-autocomplete-dir"
				defaultValue = "/etc/bash_completion.d"
				description  = "autocompletion directory"
			)

			flags := parent.Cobra.PersistentFlags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}
