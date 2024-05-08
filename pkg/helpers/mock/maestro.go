package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/openshift-online/maestro/pkg/api/openapi"
)

const Consumer = "maestro-build-in-consumer"

type MaestroMockServer struct {
	server *httptest.Server
}

func NewMaestroMockServer() *MaestroMockServer {
	return &MaestroMockServer{
		server: httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				list := openapi.ConsumerList{}

				if strings.Contains(r.URL.RawQuery, Consumer) {
					consumer := openapi.NewConsumer()
					consumer.SetName(Consumer)
					list.Items = []openapi.Consumer{*consumer}
				}

				data, _ := json.Marshal(list)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(data)
			case http.MethodPost:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
			default:
				w.WriteHeader(http.StatusNotImplemented)
			}
		})),
	}
}

func (m *MaestroMockServer) URL() string {
	return m.server.URL
}

func (m *MaestroMockServer) Start() {
	m.server.Start()
}

func (m *MaestroMockServer) Stop() {
	m.server.Close()
}
