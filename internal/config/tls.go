package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

/* Bacground for TLS :
1) Decides what cipher client and server will use, autheticate server, generate session keys
	a) How it does authenticate server is that server has some private key and anyone with the public key can decipher it to
	verify it's the server

	The handshake also handles authentication, which usually consists of the server proving its identity to the client. This is done using public keys. Public keys are encryption keys that use one-way encryption, meaning that anyone with the public key can unscramble the data encrypted with the server's private key to ensure its authenticity, but only the original sender can encrypt data with the private key. The server's public key is part of its TLS certificate.

2)Use mTLS i.e that is both client and server need to prove their identity as opposed top just server. This is usally used in districuted systes
with machine to machine communication

3) We need certs to do the tls handshake. They can be obtained from third part CA(vertificate authority) but we will be our own
certificate authority in this case ?? (Don't know why that works)

4)ca-csr.json will be used to configure our ca's certificate

5)ca-config.json specifies what kind of certificates our CA will issue. Signing section defines CA's signing policy
Our configuration file says that the CA can generate client and server certificates that will expire after a year and the
certificates may be used for digital signatures, encrypting keys, and auth.

6)server-csr.json will be used to configure our server's certs.

7) Various tls config's generated below
	a) Client Tls with just Root CA setup : Client to verify server's certificate(via CA)
	b) Client Tls with Root CA and CertKey and CertValue setup.: Client to verify server and server to verify client
	c) Server Tls with clientCA,Certificate and Client Auth setup : Again nboth way verification

*/

func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(
			cfg.CertFile,
			cfg.KeyFile,
		)
		if err != nil {
			return nil, err
		}
	}
	if cfg.CAFile != "" {
		b, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(b))
		if !ok {
			return nil, fmt.Errorf(
				"failed to parse root certificate: %q",
				cfg.CAFile,
			)
		}
		if cfg.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}
		tlsConfig.ServerName = cfg.ServerAddress
	}
	return tlsConfig, nil
}

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}
