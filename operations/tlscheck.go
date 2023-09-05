package operations

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/settings"
)

var (
	tlsVersions = map[uint16]string{}
)

func init() {
	tlsVersions[tls.VersionSSL30] = "SSLv3"
	tlsVersions[tls.VersionTLS10] = "TLS 1.0"
	tlsVersions[tls.VersionTLS11] = "TLS 1.1"
	tlsVersions[tls.VersionTLS12] = "TLS 1.2"
	tlsVersions[tls.VersionTLS13] = "TLS 1.3"
}

func get(url string) (*tls.ConnectionState, error) {
	transport := settings.Global.ConfiguredHttpTransport()
	transport.TLSClientConfig.InsecureSkipVerify = true
	client := http.Client{Transport: transport}
	response, err := client.Head(url)
	if err != nil {
		return nil, err
	}
	return response.TLS, nil
}

func tlsCheckHost(host string) []*common.DiagnosticCheck {
	transport := settings.Global.ConfiguredHttpTransport()
	result := []*common.DiagnosticCheck{}
	supportNetworkUrl := settings.Global.DocsLink("troubleshooting/firewall-and-proxies")
	url := fmt.Sprintf("https://%s/", host)
	state, err := get(url)
	if err != nil {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkLink,
			Status:   statusWarning,
			Message:  fmt.Sprintf("%s -> %v", url, err),
			Link:     supportNetworkUrl,
		})
		return result
	}
	server := state.ServerName
	version, ok := tlsVersions[state.Version]
	if !ok {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVersion,
			Status:   statusWarning,
			Message:  fmt.Sprintf("unknown TLS version: %q -> %03x", host, state.Version),
			Link:     supportNetworkUrl,
		})
	} else {
		tlsStatus := statusOk
		if state.Version < tls.VersionTLS12 {
			tlsStatus = statusWarning
		}
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVersion,
			Status:   tlsStatus,
			Message:  fmt.Sprintf("TLS version: %q -> %s", host, version),
			Link:     supportNetworkUrl,
		})
	}
	toVerify := x509.VerifyOptions{
		DNSName:       server,
		Roots:         transport.TLSClientConfig.RootCAs,
		Intermediates: x509.NewCertPool(),
	}
	certificates := state.PeerCertificates
	if len(certificates) == 0 {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVerify,
			Status:   statusWarning,
			Message:  fmt.Sprintf("no certificates for %s", server),
			Link:     supportNetworkUrl,
		})
		return result
	}
	last := certificates[0]
	for _, certificate := range certificates[1:] {
		toVerify.Intermediates.AddCert(certificate)
		last = certificate
	}
	_, err = certificates[0].Verify(toVerify)
	if err != nil {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVerify,
			Status:   statusWarning,
			Message:  fmt.Sprintf("TLS verification of %q failed, reason: %v [last issuer: %q]", server, err, last.Issuer),
			Link:     supportNetworkUrl,
		})
	} else {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVerify,
			Status:   statusOk,
			Message:  fmt.Sprintf("TLS verification of %q passed with certificate issued by %q", server, last.Issuer),
			Link:     supportNetworkUrl,
		})
	}
	return result
}
