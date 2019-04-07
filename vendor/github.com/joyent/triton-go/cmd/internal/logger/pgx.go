//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package logger

import (
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
)

type PGX struct {
	l zerolog.Logger
}

func NewPGX(l zerolog.Logger) pgx.Logger {
	return &PGX{l: l}
}

func (l PGX) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelDebug:
		l.l.Debug().Fields(data).Msg(msg)
	case pgx.LogLevelInfo:
		l.l.Info().Fields(data).Msg(msg)
	case pgx.LogLevelWarn:
		l.l.Warn().Fields(data).Msg(msg)
	case pgx.LogLevelError:
		l.l.Error().Fields(data).Msg(msg)
	default:
		l.l.Debug().Fields(data).Str("level", level.String()).Msg(msg)
	}
}
