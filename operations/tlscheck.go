package operations

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/fail"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/settings"
)

type (
	tlsConfigs []*tls.Config
)

var (
	tlsVersions   = map[uint16]string{}
	knownVersions = []uint16{
		tls.VersionTLS13,
		tls.VersionTLS12,
		tls.VersionTLS11,
		tls.VersionTLS10,
		tls.VersionSSL30,
	}
)

func init() {
	tlsVersions[tls.VersionSSL30] = "SSLv3"
	tlsVersions[tls.VersionTLS10] = "TLS 1.0"
	tlsVersions[tls.VersionTLS11] = "TLS 1.1"
	tlsVersions[tls.VersionTLS12] = "TLS 1.2"
	tlsVersions[tls.VersionTLS13] = "TLS 1.3"
}

func tlsCheckHeadOnly(url string) (*tls.ConnectionState, error) {
	transport := settings.Global.ConfiguredHttpTransport()
	transport.TLSClientConfig.InsecureSkipVerify = true
	transport.TLSClientConfig.MinVersion = tls.VersionSSL30
	// above two setting are needed for TLS checks
	// they weaken security, and that is why this code is only used
	// to get TLS connection state and nothing else
	// this is intentional, so that network diagnosis can detect
	// unsecure certificates, and connections to weaker TLS version
	// [ref: Github CodeQL security warning]
	client := http.Client{Transport: transport}
	response, err := client.Head(url)
	if err != nil {
		return nil, err
	}
	return response.TLS, nil
}

func certificateChain(certificates []*x509.Certificate) string {
	parts := make([]string, 0, len(certificates))
	for at, certificate := range certificates {
		names := strings.Join(certificate.DNSNames, ", ")
		before := certificate.NotBefore.Format("2006-Jan-02")
		after := certificate.NotAfter.Format("2006-Jan-02")
		form := fmt.Sprintf("#%d: [% 02X ...] names [%s] %s...%s %q issued by %q", at, certificate.Signature[:6], names, before, after, certificate.Subject, certificate.Issuer)
		parts = append(parts, form)
	}
	return strings.Join(parts, "; ")
}

func tlsCheckHost(host string, roots map[string]bool) []*common.DiagnosticCheck {
	transport := settings.Global.ConfiguredHttpTransport()
	result := []*common.DiagnosticCheck{}
	supportNetworkUrl := settings.Global.DocsLink("troubleshooting/firewall-and-proxies")
	url := fmt.Sprintf("https://%s/", host)
	state, err := tlsCheckHeadOnly(url)
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
	roots[last.Issuer.String()] = err == nil
	if err != nil {
		result = append(result, &common.DiagnosticCheck{
			Type:     "network",
			Category: common.CategoryNetworkTLSVerify,
			Status:   statusWarning,
			Message:  fmt.Sprintf("TLS verification of %q failed, reason: %v [last issuer: %q]", server, err, last.Issuer),
			Link:     supportNetworkUrl,
		})
		if common.DebugFlag() {
			result = append(result, &common.DiagnosticCheck{
				Type:     "network",
				Category: common.CategoryNetworkTLSChain,
				Status:   statusWarning,
				Message:  fmt.Sprintf("%q certificate chain is {%s}.", host, certificateChain(certificates)),
				Link:     supportNetworkUrl,
			})
		}
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

func configurationVariations(root *x509.CertPool) tlsConfigs {
	configs := make(tlsConfigs, len(knownVersions))
	for at, version := range knownVersions {
		configs[at] = &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            root,
			MinVersion:         version,
			MaxVersion:         version,
		}
	}
	return configs
}

func certificateFingerprint(certificate *x509.Certificate) string {
	if certificate == nil {
		return "[nil]"
	}
	return fmt.Sprintf("[% 02X ...]", certificate.Signature[:10])
}

func probeVersion(serverport string, config *tls.Config, seen map[string]int) {
	dialer := &tls.Dialer{
		Config: config,
	}
	timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
	intermediate, err := dialer.DialContext(timeout, "tcp", serverport)
	if err != nil {
		common.Log("  %s%s failed, reason: %v%s", pretty.Yellow, tlsVersions[config.MinVersion], err, pretty.Reset)
		return
	}
	defer intermediate.Close()
	conn, ok := intermediate.(*tls.Conn)
	if !ok {
		common.Log("  %s%s failed, reason: could not covert to TLS connection.%s", pretty.Yellow, tlsVersions[config.MinVersion], pretty.Reset)
		return
	}
	state := conn.ConnectionState()
	cipher := tls.CipherSuiteName(state.CipherSuite)
	version, ok := tlsVersions[state.Version]
	if !ok {
		version = fmt.Sprintf("unknown: %03x", state.Version)
	}
	server := state.ServerName
	toVerify := x509.VerifyOptions{
		DNSName:       server,
		Roots:         config.RootCAs,
		Intermediates: x509.NewCertPool(),
	}
	common.Log("  %s%s supported, cipher suite %q, server: %q%s", pretty.Green, version, cipher, server, pretty.Reset)
	certificates := state.PeerCertificates
	before := len(seen)
	for at, certificate := range certificates {
		if at > 0 {
			toVerify.Intermediates.AddCert(certificate)
		}
		fingerprint := certificateFingerprint(certificate)
		hit, ok := seen[fingerprint]
		if ok {
			common.Log("    %s#%d: [ID:%d] again %s%s", pretty.Grey, at, hit, fingerprint, pretty.Reset)
			continue
		}
		hit = len(seen) + 1
		seen[fingerprint] = hit
		names := strings.Join(certificate.DNSNames, ", ")
		before := certificate.NotBefore.Format("2006-Jan-02")
		after := certificate.NotAfter.Format("2006-Jan-02")
		common.Log("    #%d: %s[ID:%d]%s %s %s - %s [%s]", at, pretty.Magenta, hit, pretty.Reset, fingerprint, before, after, names)
		common.Log("      + subject %s", certificate.Subject)
		common.Log("      + issuer %s", certificate.Issuer)
	}
	if len(seen) == before {
		return
	}
	_, err = certificates[0].Verify(toVerify)
	if err != nil {
		common.Log("    %s!!! verification failure: %v%s", pretty.Red, err, pretty.Reset)
	}
}

func probeServer(index int, serverport string, variations tlsConfigs, seen map[string]int) {
	common.Log("%s#%d: Server %q%s", pretty.Cyan, index, serverport, pretty.Reset)
	for _, variation := range variations {
		probeVersion(serverport, variation, seen)
	}
}

func TLSProbe(serverports []string) (err error) {
	defer fail.Around(&err)

	root, err := x509.SystemCertPool()
	fail.On(err != nil, "Cannot get system certificate pool, reason: %v", err)

	seen := make(map[string]int)

	variations := configurationVariations(root)
	for at, serverport := range serverports {
		if at > 0 {
			common.Log("--")
		}
		probeServer(at+1, serverport, variations, seen)
	}
	return nil
}
