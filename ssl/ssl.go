package ssl

import (
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

// CertificateInfo holds SSL/TLS certificate details
type CertificateInfo struct {
	Subject       string    `json:"subject"`
	Issuer        string    `json:"issuer"`
	NotBefore     time.Time `json:"not_before"`
	NotAfter      time.Time `json:"not_after"`
	DaysRemaining int       `json:"days_remaining"`
	IsExpired     bool      `json:"is_expired"`
	IsValid       bool      `json:"is_valid"`
	Error         string    `json:"error,omitempty"`
}

// GetCertificate connects to a host and retrieves TLS certificate info
func GetCertificate(host string, timeout time.Duration) CertificateInfo {
	info := CertificateInfo{}

	// Try with the given host first, then add port 443 if needed
	hostport := ensurePort(host)

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostport, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		info.Error = fmt.Sprintf("failed to connect: %v", err)
		info.IsValid = false
		return info
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		info.Error = "no peer certificates found"
		info.IsValid = false
		return info
	}

	// Use the leaf certificate (first in chain)
	cert := certs[0]
	info.Subject = cert.Subject.CommonName
	if info.Subject == "" {
		info.Subject = cert.Subject.String()
	}
	info.Issuer = cert.Issuer.CommonName
	if info.Issuer == "" {
		info.Issuer = cert.Issuer.String()
	}
	info.NotBefore = cert.NotBefore
	info.NotAfter = cert.NotAfter

	now := time.Now()
	remaining := cert.NotAfter.Sub(now)
	info.DaysRemaining = int(math.Floor(remaining.Hours() / 24))

	switch {
	case now.Before(cert.NotBefore):
		info.IsExpired = false
		info.IsValid = false
		info.Error = "certificate is not yet valid"
	case now.After(cert.NotAfter):
		info.IsExpired = true
		info.IsValid = false
		info.Error = fmt.Sprintf("certificate expired %d days ago", -info.DaysRemaining)
	default:
		info.IsExpired = false
		info.IsValid = true
	}

	return info
}

// ensurePort adds :443 if no port is present in the host string
func ensurePort(host string) string {
	if strings.Contains(host, ":") {
		return host
	}
	return host + ":443"
}

// ExtractHostname extracts the hostname from a URL for TLS connection
func ExtractHostname(url string) string {
	// Remove protocol prefix
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}

	// Remove path and query, but keep port
	for i := 0; i < len(url); i++ {
		if url[i] == '/' || url[i] == '?' {
			return url[:i]
		}
	}
	return url
}
