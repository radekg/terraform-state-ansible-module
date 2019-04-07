//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package get

import (
	"fmt"

	"github.com/joyent/triton-go/cmd/agent/account"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "get",
		Short:        "Show account information",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cons := conswriter.GetTerminal()

			c, err := cfg.NewTritonConfig()
			if err != nil {
				return err
			}

			a, err := account.NewAccountClient(c)
			if err != nil {
				return err
			}

			accDetails, err := a.Get()
			if err != nil {
				return err
			}

			cons.Write([]byte(fmt.Sprintf("id: %s", accDetails.ID)))
			cons.Write([]byte(fmt.Sprintf("\nlogin: %s", accDetails.Login)))
			cons.Write([]byte(fmt.Sprintf("\nemail: %s", accDetails.Email)))
			cons.Write([]byte(fmt.Sprintf("\ncompanyName: %s", accDetails.CompanyName)))
			cons.Write([]byte(fmt.Sprintf("\nfirstName: %s", accDetails.FirstName)))
			cons.Write([]byte(fmt.Sprintf("\nlastName: %s", accDetails.LastName)))
			cons.Write([]byte(fmt.Sprintf("\npostalCode: %s", accDetails.PostalCode)))
			cons.Write([]byte(fmt.Sprintf("\ntriton_cns_enabled: %t", accDetails.TritonCNSEnabled)))
			cons.Write([]byte(fmt.Sprintf("\naddress: %s", accDetails.Address)))
			cons.Write([]byte(fmt.Sprintf("\ncity: %s", accDetails.City)))
			cons.Write([]byte(fmt.Sprintf("\nstate: %s", accDetails.State)))
			cons.Write([]byte(fmt.Sprintf("\ncountry: %s", accDetails.Country)))
			cons.Write([]byte(fmt.Sprintf("\nphone: %s", accDetails.Phone)))
			cons.Write([]byte(fmt.Sprintf("\nupdated: %s (%s)", accDetails.Updated.String(), cfg.FormatTime(accDetails.Updated))))
			cons.Write([]byte(fmt.Sprintf("\ncreated: %s (%s)", accDetails.Created.String(), cfg.FormatTime(accDetails.Created))))

			return nil
		},
	},

	Setup: func(parent *command.Command) error {
		return nil
	},
}
