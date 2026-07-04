package main

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebSocketAcceptMatchesRFCExample(t *testing.T) {
	got := websocketAccept("dGhlIHNhbXBsZSBub25jZQ==")
	want := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got != want {
		t.Fatalf("accept = %q, want %q", got, want)
	}
}

func TestWebSocketConnSendsTextFrame(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	ws := &websocketConn{conn: serverConn}
	errs := make(chan error, 1)
	go func() {
		errs <- ws.sendText(`{"action":"refresh"}`)
	}()

	opcode, payload, err := readFrame(bufio.NewReader(clientConn))
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if opcode != 1 {
		t.Fatalf("opcode = %d, want 1", opcode)
	}
	if got, want := string(payload), `{"action":"refresh"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	if err := <-errs; err != nil {
		t.Fatalf("send text: %v", err)
	}
}

func TestServeWebSocketRegistersClientAndReceivesBroadcast(t *testing.T) {
	server := newSimpleBarServer()
	recorder, clientConn := newHijackRecorder()
	defer clientConn.Close()
	_ = clientConn.SetDeadline(time.Now().Add(time.Second))

	request := httptest.NewRequest(http.MethodGet, "http://localhost/?target=battery", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	done := make(chan struct{})
	go func() {
		server.serveWebSocket(recorder, request)
		close(done)
	}()

	reader := bufio.NewReader(clientConn)
	status, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read handshake status: %v", err)
	}
	if !strings.Contains(status, "101 Switching Protocols") {
		t.Fatalf("handshake status = %q, want 101", status)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read handshake header: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}

	waitForClient(t, server)
	requestDone := make(chan struct{})
	go func() {
		perform(server, http.MethodGet, "/widget/battery/refresh", "")
		close(requestDone)
	}()

	opcode, payload, err := readFrame(reader)
	if err != nil {
		t.Fatalf("read broadcast frame: %v", err)
	}
	if opcode != 1 {
		t.Fatalf("opcode = %d, want 1", opcode)
	}
	if got, want := string(payload), `{"action":"refresh"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("request did not return after websocket broadcast")
	}

	_ = clientConn.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("websocket handler did not return after client close")
	}
}

type hijackRecorder struct {
	header http.Header
	conn   net.Conn
	body   bytes.Buffer
	status int
}

func newHijackRecorder() (*hijackRecorder, net.Conn) {
	serverConn, clientConn := net.Pipe()
	return &hijackRecorder{header: http.Header{}, conn: serverConn}, clientConn
}

func (r *hijackRecorder) Header() http.Header {
	return r.header
}

func (r *hijackRecorder) Write(payload []byte) (int, error) {
	return r.body.Write(payload)
}

func (r *hijackRecorder) WriteHeader(status int) {
	r.status = status
}

func (r *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.conn, bufio.NewReadWriter(bufio.NewReader(r.conn), bufio.NewWriter(r.conn)), nil
}

func waitForClient(t *testing.T, server *simpleBarServer) {
	t.Helper()
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(server.hub.snapshot()) == 1 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for websocket client registration")
}
