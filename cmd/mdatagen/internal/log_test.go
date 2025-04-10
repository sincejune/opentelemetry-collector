// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestLogNameRender(t *testing.T) {
	for _, tt := range []struct {
		name               LogName
		success            bool
		expectedExported   string
		expectedUnExported string
	}{
		{"", false, "", ""},
		{"otel.val", true, "OtelVal", "otelVal"},
		{"otel_val_2", true, "OtelVal2", "otelVal2"},
	} {
		exported, err := tt.name.Render()
		if tt.success {
			require.Nil(t, err)
			assert.Equal(t, tt.expectedExported, exported)
		} else {
			require.NotNil(t, err)
		}

		unexported, err := tt.name.RenderUnexported()
		if tt.success {
			require.Nil(t, err)
			assert.Equal(t, tt.expectedUnExported, unexported)
		} else {
			require.NotNil(t, err)
		}
	}
}
