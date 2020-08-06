// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfirmationProvider(t *testing.T) {
	spec.Run(t, "TestDefaultConfirmationProvider", testDefaultConfirmationProvider)
}

func testDefaultConfirmationProvider(t *testing.T, when spec.G, it spec.S) {
	var (
		confirmationProvider defaultConfirmationProvider
	)

	when("message is empty", func() {
		it("errors with no message provided", func() {
			_, err := confirmationProvider.Confirm("")
			require.Error(t, err, "Error: confirmation message cannot be empty")
		})
	})

	when("message is not empty", func() {
		const (
			message = "some-confirmation-message"
		)
		var (
			buffer = &bytes.Buffer{}
		)

		it.Before(func() {
			confirmationProvider.writer = buffer
		})

		it("uses provided message for confirmation", func() {
			confirmationProvider.reader = strings.NewReader("")
			_, err := confirmationProvider.Confirm(message)
			require.NoError(t, err)
			require.Equal(t, message, buffer.String())
		})

		when("okayResponses are provided", func() {
			var (
				okayResponses = []string{"y"}
			)

			when("user provides an okay response", func() {
				it.Before(func() {
					confirmationProvider.reader = strings.NewReader(okayResponses[0])
				})
				it("returns confirmed as true and does not error", func() {
					confirmed, err := confirmationProvider.Confirm(message, okayResponses...)
					require.NoError(t, err)
					require.True(t, confirmed)
				})
			})

			when("user does not provides an okay response", func() {
				it.Before(func() {
					confirmationProvider.reader = strings.NewReader("not-okay-response")
				})
				it("returns confirmed as false and does not error", func() {
					confirmed, err := confirmationProvider.Confirm(message, okayResponses...)
					require.NoError(t, err)
					require.False(t, confirmed)
				})
			})
		})

		when("okayResponses are not provided", func() {
			it.Before(func() {
				defaultOkayResponses := []string{"yes"}
				confirmationProvider.defaultOkayResponses = defaultOkayResponses
				confirmationProvider.reader = strings.NewReader(defaultOkayResponses[0])
			})
			it("uses the default set of okay responses to confirm", func() {
				confirmed, err := confirmationProvider.Confirm(message)
				require.NoError(t, err)
				require.True(t, confirmed)
			})
		})
	})
}
