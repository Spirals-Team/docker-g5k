package docker

import (
	"fmt"
	"io/ioutil"

	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnutils"
)

// GenerateNewClientCertificate generate a new Docker client certificate and put it in the host storage directory
func GenerateNewClientCertificate(h *host.Host, certName string) error {
	authOptions := h.AuthOptions()

	// create cert/key paths
	certFile := fmt.Sprintf("%s/%s.pem", authOptions.StorePath, certName)
	keyFile := fmt.Sprintf("%s/%s-key.pem", authOptions.StorePath, certName)

	// configure key length (bits) and organization name for certificate
	org := mcnutils.GetUsername() + "." + h.Name
	bits := 2048

	// generate a new client certificate
	err := cert.GenerateCert(&cert.Options{
		Hosts:       []string{""},
		CertFile:    certFile,
		KeyFile:     keyFile,
		CAFile:      authOptions.CaCertPath,
		CAKeyFile:   authOptions.CaPrivateKeyPath,
		Org:         org,
		Bits:        bits,
		SwarmMaster: false,
	})

	if err != nil {
		return fmt.Errorf("Error generating client certificate '%s'", certName)
	}

	return nil
}

// CopyCertificateToRemoteHost copy a certificate (cert + private key) to given remote host '/etc/docker' directory
func CopyCertificateToRemoteHost(h *host.Host, certName string) error {
	authOptions := h.AuthOptions()

	// read certificate file
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.pem", authOptions.StorePath, certName))
	if err != nil {
		return err
	}

	// read private key file
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s-key.pem", authOptions.StorePath, certName))
	if err != nil {
		return err
	}

	// create cert/key remote paths
	certPathDst := fmt.Sprintf("/etc/docker/%s.pem", certName)
	keyPathDst := fmt.Sprintf("/etc/docker/%s-key.pem", certName)

	// command format used to copy cert/key (PEM format) to remote host via SSH
	certTransferCmdFmt := "printf '%%s' '%s' | sudo tee %s"

	// copy certificate to remote host
	if _, err := h.RunSSHCommand(fmt.Sprintf(certTransferCmdFmt, string(cert), certPathDst)); err != nil {
		return err
	}

	// copy private key to remote host
	if _, err := h.RunSSHCommand(fmt.Sprintf(certTransferCmdFmt, string(key), keyPathDst)); err != nil {
		return err
	}

	return nil
}
