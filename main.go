package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
)

const (
	httpAddr = "127.0.0.1:7776"
	wsAddr   = "127.0.0.1:7777"
)

func main() {
	srv := newSimpleBarServer()

	wsListener, err := net.Listen("tcp", wsAddr)
	if err != nil {
		exitWithError(err)
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		_ = wsListener.Close()
		exitWithError(err)
	}

	wsServer := &http.Server{Handler: http.HandlerFunc(srv.serveWebSocket)}
	go func() {
		if err := wsServer.Serve(wsListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			exitWithError(err)
		}
	}()

	fmt.Println("simple-bar-server running at http://localhost:7776")
	refreshUebersicht()

	httpServer := &http.Server{Handler: srv}
	if err := httpServer.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		exitWithError(err)
	}
}

func refreshUebersicht() {
	cmd := exec.Command("osascript", "-e", `tell application id "tracesOf.Uebersicht" to refresh`)
	if err := cmd.Start(); err == nil {
		go func() {
			_ = cmd.Wait()
		}()
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
