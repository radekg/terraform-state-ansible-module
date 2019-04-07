//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package logger

import (
	"fmt"
	"strings"

	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/spf13/viper"
)

type Format uint

const (
	FormatAuto Format = iota
	FormatZerolog
	FormatHuman
)

func (f Format) String() string {
	switch f {
	case FormatAuto:
		return "auto"
	case FormatZerolog:
		return "zerolog"
	case FormatHuman:
		return "human"
	default:
		panic(fmt.Sprintf("unknown log format: %d", f))
	}
}

func getLogFormat() (Format, error) {
	switch logFormat := strings.ToLower(viper.GetString(config.KeyLogFormat)); logFormat {
	case "auto":
		return FormatAuto, nil
	case "json", "zerolog":
		return FormatZerolog, nil
	case "human":
		return FormatHuman, nil
	default:
		return FormatAuto, fmt.Errorf("unsupported log format: %q", logFormat)
	}
}
