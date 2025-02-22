package e2e

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/logging"
)

// TestSlackServer ... Mock server for testing slack alerts
type TestSlackServer struct {
	Server   *httptest.Server
	Payloads []*client.SlackPayload
}

// NewTestSlackServer ... Creates a new mock slack server
func NewTestSlackServer(url string, port int) *TestSlackServer { //nolint:dupl //This will be addressed
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
	if err != nil {
		panic(err)
	}

	ss := &TestSlackServer{
		Payloads: []*client.SlackPayload{},
	}

	ss.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			ss.mockSlackPost(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	err = ss.Server.Listener.Close()
	if err != nil {
		panic(err)
	}
	ss.Server.Listener = l
	ss.Server.Start()

	logging.NoContext().Info("Test slack server started", zap.String("url", url), zap.Int("port", port))

	return ss
}

// Close ... Closes the server
func (svr *TestSlackServer) Close() {
	svr.Server.Close()
}

// mockSlackPost ... Mocks a slack post request
func (svr *TestSlackServer) mockSlackPost(w http.ResponseWriter, r *http.Request) {
	var alert *client.SlackPayload

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"", "error":"could not decode slack payload"}`))
		return
	}

	svr.Payloads = append(svr.Payloads, alert)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"ok", "error":""}`))
}

// SlackAlerts ... Returns the slack alerts
func (svr *TestSlackServer) SlackAlerts() []*client.SlackPayload {
	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestSlackServer) ClearAlerts() {
	svr.Payloads = []*client.SlackPayload{}
}
