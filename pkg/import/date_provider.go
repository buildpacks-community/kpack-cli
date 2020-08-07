// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package _import

import "time"

type defaultTimestampProvider struct {
	Format string
}

func DefaultTimestampProvider() defaultTimestampProvider {
	return defaultTimestampProvider{
		Format: time.RFC3339,
	}
}

func (d defaultTimestampProvider) GetTimestamp() string {
	return time.Now().Format(d.Format)
}
