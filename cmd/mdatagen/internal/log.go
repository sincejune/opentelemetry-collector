// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/cmd/mdatagen/internal"

type LogName string

func (mn LogName) Render() (string, error) {
	return FormatIdentifier(string(mn), true)
}

func (mn LogName) RenderUnexported() (string, error) {
	return FormatIdentifier(string(mn), false)
}
