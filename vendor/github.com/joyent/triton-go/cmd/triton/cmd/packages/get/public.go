//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package get

import (
	"encoding/json"
	"errors"

	"bytes"

	"github.com/joyent/triton-go/cmd/agent/compute"
	cfg "github.com/joyent/triton-go/cmd/config"
	"github.com/joyent/triton-go/cmd/internal/command"
	"github.com/sean-/conswriter"
	"github.com/spf13/cobra"
)

var Cmd = &command.Command{
	Cobra: &cobra.Command{
		Args:         cobra.NoArgs,
		Use:          "get",
		Short:        "get triton package",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cfg.GetPkgID() == "" && cfg.GetPkgName() == "" {
				return errors.New("Either `id` or `name` must be specified")
			}

			if cfg.GetPkgID() != "" && cfg.GetPkgName() != "" {
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

			pkg, err := a.GetPackage()
			if err != nil {
				return err
			}

			bytes, err := json.Marshal(pkg)
			if err != nil {
				return err
			}

			output, _ := prettyPrintJSON(bytes)

			cons.Write(output)

			return nil
		},
	},
	Setup: func(parent *command.Command) error {
		return nil
	},
}

func prettyPrintJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "    ")
	return out.Bytes(), err
}
