package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	confluentkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"gopkg.in/yaml.v2"

	"github.com/stolostron/maestro-addon/pkg/mq"
	"github.com/stolostron/maestro-addon/test/performance/pkg/common"
	"github.com/stolostron/maestro-addon/test/performance/pkg/util"
)

const (
	CertificateBlockType   = "CERTIFICATE"
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
)

var (
	workDir           = flag.String("work-dir", "", "")
	kafkaServer       = flag.String("kafka-server", "", "")
	clusterBeginIndex = flag.Int("cluster-begin-index", 1, "Begin index of the clusters")
	clusterCounts     = flag.Int("cluster-counts", common.DEFAULT_CLUSTER_COUNTS, "Counts of the clusters")
)

func main() {
	flag.Parse()

	// init topics
	brokerConfigPath := filepath.Join(*workDir, "config", "kafka.admin.config")
	mqAuthzCreator, err := mq.NewMessageQueueAuthzCreator(mq.MessageQueueKafka, brokerConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	index := *clusterBeginIndex
	for i := 0; i < *clusterCounts; i++ {
		clusterName := util.ClusterName(index)

		startTime := time.Now()
		if err := mqAuthzCreator.CreateAuthorizations(context.Background(), clusterName); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("the kafka acls is prepared for cluster %s, time=%dms\n",
			clusterName, util.UsedTime(startTime, time.Millisecond))

		if err := prepareKafkaAgentConfig(clusterName); err != nil {
			log.Fatal(err)
		}
		index = index + 1
	}
}

func prepareKafkaAgentConfig(clusterName string) error {
	configPath := filepath.Join(*workDir, "config")
	certPath := filepath.Join(*workDir, "certs")
	kafkaClientCAPath := filepath.Join(*workDir, "certs", "clients-ca.crt")
	kafkaClientCAKeyPath := filepath.Join(*workDir, "certs", "clients-ca.key")

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	certFile, err := os.ReadFile(kafkaClientCAPath)
	if err != nil {
		return err
	}

	pemBlock, _ := pem.Decode(certFile)
	caCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return err
	}

	keyFile, err := os.ReadFile(kafkaClientCAKeyPath)
	if err != nil {
		return err
	}
	keyBlock, _ := pem.Decode(keyFile)
	caKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}

	clientCertDERBytes, err := x509.CreateCertificate(
		rand.Reader,
		&x509.Certificate{
			Subject: pkix.Name{
				CommonName: fmt.Sprintf("system:open-cluster-management:cluster:%s:addon:maestro-addon:agent:maestro-addon-agent", clusterName),
				Organization: []string{
					"system:authenticated",
					"system:open-cluster-management:addon:maestro-addon",
					fmt.Sprintf("system:open-cluster-management:cluster:%s:addon:maestro-addon", clusterName),
				},
			},
			SerialNumber: big.NewInt(1),
			NotBefore:    caCert.NotBefore,
			NotAfter:     caCert.NotBefore.Add(8760 * time.Hour).UTC(),
			KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
		caCert,
		clientKey.Public(),
		caKey,
	)
	if err != nil {
		return err
	}

	clientCert, err := x509.ParseCertificate(clientCertDERBytes)
	if err != nil {
		return err
	}

	if err := os.WriteFile(
		filepath.Join(certPath, fmt.Sprintf("client-%s.crt", clusterName)),
		pem.EncodeToMemory(&pem.Block{
			Type:  CertificateBlockType,
			Bytes: clientCert.Raw,
		}),
		0o600,
	); err != nil {
		return err
	}

	if err := os.WriteFile(
		filepath.Join(certPath, fmt.Sprintf("client-%s.key", clusterName)),
		pem.EncodeToMemory(&pem.Block{
			Type:  RSAPrivateKeyBlockType,
			Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
		}),
		0o600,
	); err != nil {
		return err
	}

	configData, err := yaml.Marshal(confluentkafka.ConfigMap{
		"bootstrapServer": kafkaServer,
		"caFile":          filepath.Join(*workDir, "certs", "cluster-ca.crt"),
		"clientCertFile":  filepath.Join(certPath, fmt.Sprintf("client-%s.crt", clusterName)),
		"clientKeyFile":   filepath.Join(certPath, fmt.Sprintf("client-%s.key", clusterName)),
	})
	if err != nil {
		return err
	}

	configFile := filepath.Join(configPath, fmt.Sprintf("client-%s.config", clusterName))
	if err := os.WriteFile(configFile, configData, 0o600); err != nil {
		return err
	}

	fmt.Printf("The config file %s of cluster %s is prepared\n", configFile, clusterName)
	return nil
}
