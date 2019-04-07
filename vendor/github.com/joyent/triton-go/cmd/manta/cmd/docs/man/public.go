//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package man

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/joyent/triton-go"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Use:   "man",
		Short: "Generates and installs Joyent Manta cli man pages",
		Long: `This command automatically generates up-to-date man pages of Manta CLI
command-line interface.  By default, it creates the man page files
in the "docs/man" directory under the current directory.`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Manual:  "Manta",
				Section: strconv.Itoa(config.ManSect),
				Source:  strings.Join([]string{"Manta", triton.Version}, " "),
			}

			manDir := viper.GetString(config.KeyDocManDir)

			manSectDir := path.Join(manDir, fmt.Sprintf("man%d", config.ManSect))
			if _, err := os.Stat(manSectDir); os.IsNotExist(err) {
				if err := os.MkdirAll(manSectDir, 0777); err != nil {
					return errors.Wrapf(err, "unable to make mandir %q", manSectDir)
				}
			}

			cmd.Root().DisableAutoGenTag = true
			log.Info().Str("MANDIR", manDir).Int("section", config.ManSect).Msg("Installing man(1) pages")

			err := doc.GenManTree(cmd.Root(), header, manSectDir)
			if err != nil {
				return errors.Wrap(err, "unable to generate man(1) pages")
			}

			log.Info().Msg("Installation completed successfully.")

			return nil
		},
	},
	Setup: func(parent *command.Command) error {

		{
			const (
				key          = config.KeyDocManDir
				longName     = "man-dir"
				shortName    = "m"
				description  = "Specify the MANDIR to use"
				defaultValue = config.DefaultManDir
			)

			flags := parent.Cobra.PersistentFlags()
			flags.StringP(longName, shortName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
			viper.BindEnv(key, "MANDIR")
			viper.SetDefault(key, defaultValue)
		}

		return nil
	},
}
