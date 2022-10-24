package uaaclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/lager"
	uaa "github.com/cloudfoundry-community/go-uaa"
)

func newAPI(cfg Config, logger lager.Logger) (*uaa.API, error) {
	if cfg.Port == -1 {
		return nil, errors.New("tls-not-enabled: UAA client requires TLS enabled")
	}

	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.SkipSSLValidation}
	if cfg.CACerts != "" {
		certBytes, err := ioutil.ReadFile(cfg.CACerts)
		if err != nil {
			return nil, fmt.Errorf("Failed to read ca cert file: %s", err.Error())
		}

		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
			return nil, errors.New("Unable to load caCert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := &http.Client{Transport: tr}
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	tokenURL := fmt.Sprintf("https://%s:%d", cfg.TokenEndpoint, cfg.Port)
	if cfg.ClientName != "" && cfg.ClientSecret != "" {
		return uaa.New(tokenURL, uaa.WithClientCredentials(cfg.ClientName, cfg.ClientSecret, uaa.JSONWebToken), uaa.WithClient(httpClient), uaa.WithSkipSSLValidation(cfg.SkipSSLValidation))
	}

	return uaa.New(tokenURL, uaa.WithNoAuthentication(), uaa.WithClient(httpClient), uaa.WithSkipSSLValidation(cfg.SkipSSLValidation))
}