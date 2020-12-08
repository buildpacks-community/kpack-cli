// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"

	"github.com/pivotal/build-service-cli/pkg/registry"
)

func TestTLSConfig(t *testing.T) {
	spec.Run(t, "Test TLSConfig", testTLSConfig)
}

func testTLSConfig(t *testing.T, when spec.G, it spec.S) {
	it("adds the cert to the cert pool and sets skip tls verify", func() {
		certPath := filepath.Join("testdata", "ca.crt")
		certData, err := ioutil.ReadFile(certPath)
		require.NoError(t, err)

		block, _ := pem.Decode(certData)
		require.NotNil(t, block)

		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)
		expectedSubject := cert.RawSubject

		fetcher := registry.TLSConfig{
			CaCertPath:  certPath,
			VerifyCerts: false,
		}

		transport, err := fetcher.Transport()
		require.NoError(t, err)
		subjects := transport.TLSClientConfig.RootCAs.Subjects()

		found := false
		for _, s := range subjects {
			if string(s) == string(expectedSubject) {
				found = true
			}
		}
		require.True(t, found, "cert pool did not contain expected cert")
		require.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})

	it("sets skip verify to false when verify certs is true", func() {
		fetcher := registry.TLSConfig{
			CaCertPath:  "",
			VerifyCerts: true,
		}

		transport, err := fetcher.Transport()
		require.NoError(t, err)
		require.False(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}
