package internal

import (
	"encoding/json"

	"github.com/cloudflare/cfssl/api/client"
	"github.com/cloudflare/cfssl/auth"
	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/info"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/remote"
)

func newClientCSR(user string) *csr.CertificateRequest {
	return &csr.CertificateRequest{
		CN:         user,
		KeyRequest: csr.NewBasicKeyRequest(),
	}
}

func newNodeCSR(hosts []string) *csr.CertificateRequest {
	return &csr.CertificateRequest{
		CN:         "node",
		Hosts:      hosts,
		KeyRequest: csr.NewBasicKeyRequest(),
	}
}

func createCertificateAndKey(
	address, profile, authKey string,
	req *csr.CertificateRequest) (key []byte, cert []byte, err error) {

	generator := &csr.Generator{Validator: genkey.Validator}
	csrBytes, key, err := generator.ProcessRequest(req)
	if err != nil {
		key = nil
		return
	}

	provider, err := auth.New(authKey, nil)
	if err != nil {
		key = nil
		return
	}

	signingProfile := &config.SigningProfile{
		RemoteName:     "server",
		RemoteServer:   address,
		RemoteProvider: provider,
	}
	s, err := remote.NewSigner(&config.Signing{
		Profiles: map[string]*config.SigningProfile{},
		Default:  signingProfile,
	})

	if err != nil {
		return
	}

	signReq := signer.SignRequest{
		Request: string(csrBytes),
		Profile: profile,
		Hosts:   nil,
	}

	cert, err = s.Sign(signReq)
	if err != nil {
		cert, key = nil, nil
		return
	}
	return key, cert, err
}

func getCACertificate(address string) (cert []byte, err error) {
	req := new(info.Req)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return
	}

	serv := client.NewServerTLS(address, helpers.CreateTLSConfig(nil, nil))
	resp, err := serv.Info(reqJSON)
	if err != nil {
		return
	}

	cert = []byte(resp.Certificate)
	return
}
