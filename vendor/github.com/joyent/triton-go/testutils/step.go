//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package testutils

type StepAction uint

const (
	Continue StepAction = iota
	Halt
)

const (
	StateCancelled = "cancelled"
	StateHalted    = "halted"
)

type Step interface {
	Run(TritonStateBag) StepAction

	Cleanup(TritonStateBag)
}
