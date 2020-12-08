package registry

import (
	"crypto/tls"
	"crypto/x509"
	fmt "fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"time"
)

type TLSConfig struct {
	CaCertPath  string
	VerifyCerts bool
}

func (t *TLSConfig) Transport() (*http.Transport, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	if t.CaCertPath != "" {
		if cert, err := ioutil.ReadFile(t.CaCertPath); err != nil {
			return nil, fmt.Errorf("reading CA certificate from '%s': %s", t.CaCertPath, err)
		} else if ok := pool.AppendCertsFromPEM(cert); !ok {
			return nil, fmt.Errorf("adding CA certificate from '%s': failed", t.CaCertPath)
		}
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: t.VerifyCerts == false,
		},
	}

	// Do not set RootCAs when custom CA is not set on windows
	// https://github.com/golang/go/issues/16736
	if runtime.GOOS == "windows" && t.CaCertPath == "" {
		transport.TLSClientConfig.RootCAs = nil
	}

	return transport, nil
}
