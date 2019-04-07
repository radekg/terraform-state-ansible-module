//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package list

import (
	"strings"

	"github.com/joyent/triton-go/cmd/agent/compute"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	tc "github.com/joyent/triton-go/compute"
	"github.com/olekukonko/tablewriter"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "list",
		Short:        "list triton instances",
		Aliases:      []string{"ls"},
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

			a, err := compute.NewComputeClient(c)
			if err != nil {
				return err
			}

			instances, err := a.GetInstanceList()
			if err != nil {
				return err
			}

			images, err := a.GetImagesList()
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cons)
			table.SetHeaderAlignment(tablewriter.ALIGN_RIGHT)
			table.SetHeaderLine(false)
			table.SetAutoFormatHeaders(true)

			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")

			table.SetHeader([]string{"SHORTID", "NAME", "IMG", "STATE", "FLAGS", "AGE"})

			var numInstances uint
			for _, instance := range instances {
				table.Append([]string{string(instance.ID[:8]), instance.Name, a.FormatImageName(images, instance.Image), instance.State, formatInstanceFlags(instance), cfg.FormatTime(instance.Created)})
				numInstances++
			}

			table.Render()

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		return nil
	},
}

func formatInstanceFlags(instance *tc.Instance) string {
	flags := []string{}

	if instance.Docker {
		flags = append(flags, "D")
	}
	if strings.ToLower(instance.Brand) == "kvm" {
		flags = append(flags, "K")
	}
	if instance.FirewallEnabled {
		flags = append(flags, "F")
	}

	return strings.Join(flags, "")

}
