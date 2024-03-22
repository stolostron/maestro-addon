package helpers

import (
	"os"
	"testing"
)

func TestXxx(t *testing.T) {
	// c := NewMaestroAPIClient("http://127.0.0.1:30080")
	// list, _, err := c.DefaultApi.ApiMaestroV1ConsumersGet(context.TODO()).Execute()
	// t.Errorf("%v, %v", list, err)

	certs, err := genCertPairs([]string{"maestro-mqtt", "maestro-mqtt.maestro", "maestro-mqtt.maestro.svc", "localhost"})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("ca.pem", certs.CA, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("server.pem", certs.ServerCert, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("server-key.pem", certs.ServerCertKey, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("client.pem", certs.ClientCert, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("client-key.pem", certs.ClientCertKey, 0644); err != nil {
		t.Fatal(err)
	}
}
