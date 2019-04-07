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

	humanize "github.com/dustin/go-humanize"
	"github.com/joyent/triton-go/cmd/agent/compute"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/olekukonko/tablewriter"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "list",
		Short:        "list triton packages",
		Aliases:      []string{"ls"},
		SilenceUsage: true,
		Example: `
$ triton package list
$ triton package ls
$ triton package list --disk=102400
$ triton package list --memory=8192
$ triton package list --vpcu=4
`,
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

			packages, err := a.GetPackagesList()
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cons)
			table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
			table.SetHeaderLine(false)
			table.SetAutoFormatHeaders(true)

			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")

			table.SetHeader([]string{"SHORTID", "NAME", "MEMORY", "SWAP", "DISK", "VCPUS"})

			for _, pkg := range packages {
				table.Append([]string{string(pkg.ID[:8]), pkg.Name, humanize.Bytes(uint64(pkg.Memory * 1000 * 1000)), humanize.Bytes(uint64(pkg.Swap * 1000 * 1000)), humanize.Bytes(uint64(pkg.Disk * 1000 * 1000)), parseVPCs(pkg.VCPUs)})
			}

			table.Render()

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		{
			const (
				key          = config.KeyPackageMemory
				longName     = "memory"
				defaultValue = ""
				description  = "Package Memory (in MiB)"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageDisk
				longName     = "disk"
				defaultValue = ""
				description  = "Package Disk (in MiB)"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageSwap
				longName     = "swap"
				defaultValue = ""
				description  = "Package Swap(in MiB)"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		{
			const (
				key          = config.KeyPackageVPCUs
				longName     = "vcpu"
				defaultValue = ""
				description  = "Package VCPUs"
			)

			flags := parent.Cobra.Flags()
			flags.String(longName, defaultValue, description)
			viper.BindPFlag(key, flags.Lookup(longName))
		}

		return nil
	},
}

func parseVPCs(count int64) string {
	if count == 0 {
		return "-"
	}

	return fmt.Sprintf("%d", count)
}
