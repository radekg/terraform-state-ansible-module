//
//  Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
//  This Source Code Form is subject to the terms of the Mozilla Public
//  License, v. 2.0. If a copy of the MPL was not distributed with this
//  file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute

import (
	"sort"

	"github.com/joyent/triton-go/compute"
)

type instanceSort []*compute.Instance

func sortInstances(instances []*compute.Instance) []*compute.Instance {
	sortInstances := instances
	sort.Sort(instanceSort(sortInstances))
	return sortInstances
}

func (a instanceSort) Len() int {
	return len(a)
}

func (a instanceSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a instanceSort) Less(i, j int) bool {
	itime := a[i].Created
	jtime := a[j].Created
	return itime.Unix() < jtime.Unix()
}
