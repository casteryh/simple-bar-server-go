package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPValidationMatchesUpstreamBodies(t *testing.T) {
	server := newSimpleBarServer()
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   string
	}{
		{
			name: "missing realm",
			path: "/",
			want: "Missing realm name (yabai, skhd, aerospace, widget, missive).",
		},
		{
			name: "unknown realm",
			path: "/unknown",
			want: `Unknown realm "unknown".`,
		},
		{
			name: "widget missing kind",
			path: "/widget",
			want: "Missing kind name.",
		},
		{
			name: "widget missing action",
			path: "/widget/battery",
			want: "You need to specify an action (toggle, enable, disable, refresh).",
		},
		{
			name: "widget unknown kind",
			path: "/widget/nope/refresh",
			want: `Unknown kind "nope".`,
		},
		{
			name: "widget unknown action",
			path: "/widget/battery/nope",
			want: `Unknown action "nope".`,
		},
		{
			name: "widget valid",
			path: "/widget/battery/refresh",
			want: "",
		},
		{
			name: "yabai missing kind",
			path: "/yabai",
			want: "Missing kind name.",
		},
		{
			name: "yabai unknown kind",
			path: "/yabai/nope/refresh",
			want: `Unknown kind "nope".`,
		},
		{
			name: "yabai missing action",
			path: "/yabai/spaces",
			want: "You need to specify an action (refresh).",
		},
		{
			name: "yabai unknown action",
			path: "/yabai/spaces/nope",
			want: `Unknown action "nope".`,
		},
		{
			name: "skhd missing kind",
			path: "/skhd",
			want: "Missing kind name.",
		},
		{
			name: "skhd unknown kind",
			path: "/skhd/nope/refresh",
			want: `Unknown kind "nope".`,
		},
		{
			name: "skhd missing action",
			path: "/skhd/mode",
			want: "You need to specify an action (refresh).",
		},
		{
			name: "skhd unknown action",
			path: "/skhd/mode/nope",
			want: `Unknown action "nope".`,
		},
		{
			name: "aerospace missing kind",
			path: "/aerospace",
			want: "Missing kind name.",
		},
		{
			name: "aerospace unknown kind",
			path: "/aerospace/nope/refresh",
			want: `Unknown kind "nope".`,
		},
		{
			name: "aerospace missing action",
			path: "/aerospace/spaces",
			want: "You need to specify an action (refresh).",
		},
		{
			name: "aerospace unknown action",
			path: "/aerospace/spaces/nope",
			want: `Unknown action "nope".`,
		},
		{
			name: "missive get",
			path: "/missive/push",
			want: "Method Not Allowed.",
		},
		{
			name:   "missive missing action",
			method: http.MethodPost,
			path:   "/missive",
			want:   "You need to specify an action (push).",
		},
		{
			name:   "missive unknown action",
			method: http.MethodPost,
			path:   "/missive/nope",
			want:   `Unknown action "nope".`,
		},
		{
			name:   "missive invalid json",
			method: http.MethodPost,
			path:   "/missive/push",
			body:   "{bad",
			want:   "",
		},
		{
			name:   "missive missing content",
			method: http.MethodPost,
			path:   "/missive/push",
			body:   `{"title":"hello"}`,
			want:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			method := test.method
			if method == "" {
				method = http.MethodGet
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(method, test.path, strings.NewReader(test.body))

			server.ServeHTTP(recorder, request)
			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			if got := recorder.Body.String(); got != test.want {
				t.Fatalf("body = %q, want %q", got, test.want)
			}
			if got := response.Header.Get("Content-Type"); got != "text/plain" {
				t.Fatalf("content-type = %q, want text/plain", got)
			}
		})
	}
}

func TestWidgetRoutingMatchesTargetAndUserWidgetIndex(t *testing.T) {
	server := newSimpleBarServer()
	plainBattery := make(chan string, 2)
	indexSeven := make(chan string, 2)
	indexEight := make(chan string, 2)
	cpu := make(chan string, 2)

	server.hub.add(fakeClient("battery", "", plainBattery))
	server.hub.add(fakeClient("battery", "7", indexSeven))
	server.hub.add(fakeClient("battery", "8", indexEight))
	server.hub.add(fakeClient("cpu", "", cpu))

	perform(server, http.MethodGet, "/widget/battery/refresh/7", "")
	assertMessage(t, indexSeven, `{"action":"refresh"}`)
	assertNoMessage(t, plainBattery)
	assertNoMessage(t, indexEight)
	assertNoMessage(t, cpu)

	perform(server, http.MethodGet, "/widget/battery/toggle", "")
	assertMessage(t, plainBattery, `{"action":"toggle"}`)
	assertMessage(t, indexSeven, `{"action":"toggle"}`)
	assertMessage(t, indexEight, `{"action":"toggle"}`)
	assertNoMessage(t, cpu)
}

func TestYabaiRefreshIsDelayedAndCoalesced(t *testing.T) {
	server := newSimpleBarServer()
	messages := make(chan string, 2)
	server.hub.add(fakeClient("windows", "", messages))

	perform(server, http.MethodGet, "/yabai/windows/refresh", "")
	perform(server, http.MethodGet, "/yabai/windows/refresh", "")

	assertNoMessageWithin(t, messages, 5*time.Millisecond)
	assertMessage(t, messages, `{"action":"refresh"}`)
	assertNoMessageWithin(t, messages, 40*time.Millisecond)
}

func TestSkhdModeRefreshBroadcasts(t *testing.T) {
	server := newSimpleBarServer()
	messages := make(chan string, 1)
	server.hub.add(fakeClient("mode", "", messages))

	perform(server, http.MethodGet, "/skhd/mode/refresh", "")

	assertMessage(t, messages, `{"action":"refresh"}`)
}

func TestAerospaceKeepsLastSpaceUntilQueueCleanup(t *testing.T) {
	server := newSimpleBarServer()
	messages := make(chan string, 2)
	server.hub.add(fakeClient("spaces", "", messages))

	perform(server, http.MethodGet, "/aerospace/spaces/refresh?space=1", "")
	assertMessage(t, messages, `{"action":"refresh","data":{"space":"1"}}`)

	perform(server, http.MethodGet, "/aerospace/spaces/refresh", "")
	assertMessage(t, messages, `{"action":"refresh","data":{"space":"1"}}`)
}

func TestMissivePostRespondsEmptyAndBroadcastsValidContent(t *testing.T) {
	server := newSimpleBarServer()
	messages := make(chan string, 1)
	server.hub.add(fakeClient("missive", "", messages))

	recorder := perform(server, http.MethodPost, "/missive/push", `{"content":"hello","z":1}`)

	if got := recorder.Body.String(); got != "" {
		t.Fatalf("body = %q, want empty", got)
	}
	assertMessage(t, messages, `{"action":"push","data":{"content":"hello","z":1}}`)
}

func perform(server *simpleBarServer, method, path, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	server.ServeHTTP(recorder, request)
	return recorder
}

func fakeClient(target, userWidgetIndex string, messages chan<- string) *client {
	return &client{
		target:          target,
		userWidgetIndex: userWidgetIndex,
		send: func(payload string) error {
			messages <- payload
			return nil
		},
	}
}

func assertMessage(t *testing.T, messages <-chan string, want string) {
	t.Helper()
	select {
	case got := <-messages:
		if got != want {
			t.Fatalf("message = %q, want %q", got, want)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timed out waiting for message %q", want)
	}
}

func assertNoMessage(t *testing.T, messages <-chan string) {
	t.Helper()
	assertNoMessageWithin(t, messages, 10*time.Millisecond)
}

func assertNoMessageWithin(t *testing.T, messages <-chan string, duration time.Duration) {
	t.Helper()
	select {
	case got := <-messages:
		t.Fatalf("unexpected message %q", got)
	case <-time.After(duration):
	}
}
