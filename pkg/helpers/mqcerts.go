package helpers

import (
	"context"
	"crypto/rand"

	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/stolostron/maestro-addon/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	serverCertKey    = "server.crt"
	serverPrivKeyKey = "server.key"
	clientCertKey    = "client.crt"
	clientPrivKeyKey = "client.key"
)

const keySize = 2048

const duration365d = time.Hour * 24 * 365

type messageQueueCerts struct {
	CA            []byte
	CAKey         []byte
	ServerCert    []byte
	ServerCertKey []byte
	ClientCert    []byte
	ClientCertKey []byte
}

func PrepareCerts(ctx context.Context, kubeClient kubernetes.Interface, namespace, addOnManagerNamespace string) error {
	certs, err := genCertPairs([]string{"maestro-mqtt", "maestro-mqtt.maestro", "maestro-mqtt.maestro.svc", "localhost"})
	if err != nil {
		return err
	}

	if err := prepareMQSecret(ctx, kubeClient, namespace, certs); err != nil {
		return err
	}

	if err := prepareMQSecretForAddOnManager(ctx, kubeClient, addOnManagerNamespace, certs); err != nil {
		return err
	}

	return nil
}

func prepareMQSecret(ctx context.Context,
	kubeClient kubernetes.Interface, namespace string, certs *messageQueueCerts) error {
	logger := klog.FromContext(ctx)
	_, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, common.MessageQueueCertsSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		if _, err := kubeClient.CoreV1().Secrets(namespace).Create(
			ctx,
			newMQCertsSecret(namespace, certs),
			metav1.CreateOptions{},
		); err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("secret %s/%s is created", namespace, common.MessageQueueCertsSecretName))
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to get secret %s/%s: %w", namespace, common.MessageQueueCertsSecretName, err)
	}

	return nil
}

func prepareMQSecretForAddOnManager(ctx context.Context,
	kubeClient kubernetes.Interface, addOnManagerNamespace string, certs *messageQueueCerts) error {
	logger := klog.FromContext(ctx)
	_, err := kubeClient.CoreV1().Secrets(addOnManagerNamespace).Get(ctx, common.MessageQueueCertsSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		if _, err := kubeClient.CoreV1().Secrets(addOnManagerNamespace).Create(
			ctx,
			newAddOnManagerSigningCertSecret(addOnManagerNamespace, certs),
			metav1.CreateOptions{},
		); err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("secret %s/%s is created", addOnManagerNamespace, common.MessageQueueCertsSecretName))
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to get secret %s/%s: %w", addOnManagerNamespace, common.MessageQueueCertsSecretName, err)
	}

	return nil
}

func newMQCertsSecret(namespace string, certs *messageQueueCerts) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      common.MessageQueueCertsSecretName,
		},
		Data: map[string][]byte{
			common.MessageQueueCAKey: certs.CA,
			serverCertKey:            certs.ServerCert,
			serverPrivKeyKey:         certs.ServerCertKey,
			clientCertKey:            certs.ClientCert,
			clientPrivKeyKey:         certs.ClientCertKey,
		},
	}
}

func newAddOnManagerSigningCertSecret(namespace string, certs *messageQueueCerts) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      common.MessageQueueCertsSecretName,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       certs.CA,
			corev1.TLSPrivateKeyKey: certs.CAKey,
		},
	}
}

func genCertPairs(hosts []string) (*messageQueueCerts, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// sign ca
	serial, err := genSerial()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Maestro AddOn", Organization: []string{"ACM"}},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	rootDer, err := x509.CreateCertificate(rand.Reader, template, template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, err
	}
	rootCert, err := x509.ParseCertificate(rootDer)
	if err != nil {
		return nil, err
	}

	// sign server cert
	serial, err = genSerial()
	if err != nil {
		return nil, err
	}

	serverTemplate := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "maestro-mq-server"},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              hosts,
		BasicConstraintsValid: true,
	}
	serverDer, err := x509.CreateCertificate(rand.Reader, serverTemplate, rootCert, serverPrivateKey.Public(), privateKey)
	if err != nil {
		return nil, err
	}

	serverCert, err := x509.ParseCertificate(serverDer)
	if err != nil {
		return nil, err
	}

	// sign client cert
	serial, err = genSerial()
	if err != nil {
		return nil, err
	}

	clientTemplate := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "maestro-mq-client"},
		NotBefore:             now.UTC(),
		NotAfter:              now.Add(duration365d).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{"maestro"},
		BasicConstraintsValid: true,
	}
	clientDer, err := x509.CreateCertificate(rand.Reader, clientTemplate, rootCert, clientPrivateKey.Public(), privateKey)
	if err != nil {
		return nil, err
	}

	clientCert, err := x509.ParseCertificate(clientDer)
	if err != nil {
		return nil, err
	}

	return &messageQueueCerts{
		CA:            pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCert.Raw}),
		CAKey:         pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}),
		ServerCert:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCert.Raw}),
		ServerCertKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)}),
		ClientCert:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCert.Raw}),
		ClientCertKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)}),
	}, nil
}

func genSerial() (*big.Int, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, err
	}
	return new(big.Int).Add(serial, big.NewInt(1)), nil
}
