package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

const queueDelay = 20 * time.Millisecond

type simpleBarServer struct {
	hub             *hub
	yabaiQueues     map[string]*queue
	aerospaceQueues map[string]*queue
	aerospaceSpaceQ *queue
}

func newSimpleBarServer() *simpleBarServer {
	return &simpleBarServer{
		hub: newHub(),
		yabaiQueues: map[string]*queue{
			"spaces":   {},
			"windows":  {},
			"displays": {},
		},
		aerospaceQueues: map[string]*queue{
			"spaces": {},
		},
		aerospaceSpaceQ: &queue{},
	}
}

func (s *simpleBarServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	segments := pathSegments(r)
	realm := segment(segments, 0)
	kind := segment(segments, 1)
	action := segment(segments, 2)
	userWidgetIndex := segment(segments, 3)

	if realm == "" {
		writeText(w, "Missing realm name ("+strings.Join(realms, ", ")+").")
		return
	}

	if !contains(realms, realm) {
		writeText(w, `Unknown realm "`+realm+`".`)
		return
	}

	switch realm {
	case "widget":
		s.widgetAction(w, kind, action, userWidgetIndex)
	case "yabai":
		s.yabaiAction(w, kind, action)
	case "skhd":
		s.skhdAction(w, kind, action)
	case "aerospace":
		s.aerospaceAction(w, r, kind, action)
	case "missive":
		s.missiveAction(w, r, kind)
	}
}

func pathSegments(r *http.Request) []string {
	path := r.URL.EscapedPath()
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] == "" {
		return parts[1:]
	}
	return parts
}

func segment(segments []string, index int) string {
	if index >= len(segments) {
		return ""
	}
	return segments[index]
}

func writeText(w http.ResponseWriter, text string) {
	_, _ = io.WriteString(w, text)
}

func (s *simpleBarServer) widgetAction(w http.ResponseWriter, kind, action, userWidgetIndex string) {
	if kind == "" {
		writeText(w, "Missing kind name.")
		return
	}

	if !contains(widgetKinds, kind) {
		writeText(w, `Unknown kind "`+kind+`".`)
		return
	}

	if action == "" {
		writeText(w, "You need to specify an action ("+strings.Join(widgetActions, ", ")+").")
		return
	}

	if !contains(widgetActions, action) {
		writeText(w, `Unknown action "`+action+`".`)
		return
	}

	payload := actionPayload(action)
	s.hub.broadcast(func(client *client) bool {
		return client.target == kind && (userWidgetIndex == "" || client.userWidgetIndex == userWidgetIndex)
	}, payload)
}

func (s *simpleBarServer) yabaiAction(w http.ResponseWriter, kind, action string) {
	if kind == "" {
		writeText(w, "Missing kind name.")
		return
	}

	if !contains(yabaiKinds, kind) {
		writeText(w, `Unknown kind "`+kind+`".`)
		return
	}

	if action == "" {
		writeText(w, "You need to specify an action ("+strings.Join(yabaiActions, ", ")+").")
		return
	}

	if !contains(yabaiActions, action) {
		writeText(w, `Unknown action "`+action+`".`)
		return
	}

	q := s.yabaiQueues[kind]
	q.enqueue(action)

	time.AfterFunc(queueDelay, func() {
		lastAction := q.peekAndEmpty()
		if lastAction != "" {
			payload := actionPayload(lastAction)
			s.hub.broadcast(func(client *client) bool {
				return client.target == kind
			}, payload)
		}
	})
}

func (s *simpleBarServer) skhdAction(w http.ResponseWriter, kind, action string) {
	if kind == "" {
		writeText(w, "Missing kind name.")
		return
	}

	if !contains(skhdKinds, kind) {
		writeText(w, `Unknown kind "`+kind+`".`)
		return
	}

	if action == "" {
		writeText(w, "You need to specify an action ("+strings.Join(skhdActions, ", ")+").")
		return
	}

	if !contains(skhdActions, action) {
		writeText(w, `Unknown action "`+action+`".`)
		return
	}

	payload := actionPayload(action)
	s.hub.broadcast(func(client *client) bool {
		return client.target == kind
	}, payload)
}

func (s *simpleBarServer) aerospaceAction(w http.ResponseWriter, r *http.Request, kind, action string) {
	if kind == "" {
		writeText(w, "Missing kind name.")
		return
	}

	if !contains(aerospaceKinds, kind) {
		writeText(w, `Unknown kind "`+kind+`".`)
		return
	}

	if action == "" {
		writeText(w, "You need to specify an action ("+strings.Join(aerospaceActions, ", ")+").")
		return
	}

	if !contains(aerospaceActions, action) {
		writeText(w, `Unknown action "`+action+`".`)
		return
	}

	space := r.URL.Query().Get("space")
	q := s.aerospaceQueues[kind]
	q.enqueue(action)
	if space != "" && strings.TrimSpace(space) != "" {
		s.aerospaceSpaceQ.enqueue(space)
	}

	time.AfterFunc(queueDelay, func() {
		lastAction := q.peekAndEmpty()
		space := s.aerospaceSpaceQ.get()
		if lastAction != "" && space != "" {
			payload := aerospaceSpacePayload(lastAction, space)
			s.hub.broadcast(func(client *client) bool {
				return client.target == kind
			}, payload)
		} else if lastAction != "" {
			payload := actionPayload(lastAction)
			s.hub.broadcast(func(client *client) bool {
				return client.target == kind
			}, payload)
		}

		if s.aerospaceSpaceQ.length() > 100 {
			s.aerospaceSpaceQ.empty()
		}
	})
}

func (s *simpleBarServer) missiveAction(w http.ResponseWriter, r *http.Request, action string) {
	if r.Method != http.MethodPost {
		writeText(w, "Method Not Allowed.")
		return
	}

	if action == "" {
		writeText(w, "You need to specify an action ("+strings.Join(missiveActions, ", ")+").")
		return
	}

	if !contains(missiveActions, action) {
		writeText(w, `Unknown action "`+action+`".`)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return
	}

	object, ok := parsed.(map[string]any)
	if !ok || !truthy(object["content"]) {
		return
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, body); err != nil {
		return
	}

	payload := missivePayload(action, compact.Bytes())
	s.hub.broadcast(func(client *client) bool {
		return client.target == "missive"
	}, payload)
}

func truthy(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	case float64:
		return value != 0
	case string:
		return value != ""
	default:
		return true
	}
}

func actionPayload(action string) string {
	payload, _ := json.Marshal(struct {
		Action string `json:"action"`
	}{Action: action})
	return string(payload)
}

func aerospaceSpacePayload(action, space string) string {
	payload, _ := json.Marshal(struct {
		Action string `json:"action"`
		Data   struct {
			Space string `json:"space"`
		} `json:"data"`
	}{
		Action: action,
		Data: struct {
			Space string `json:"space"`
		}{Space: space},
	})
	return string(payload)
}

func missivePayload(action string, data []byte) string {
	payload, _ := json.Marshal(struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}{Action: action, Data: json.RawMessage(data)})
	return string(payload)
}
