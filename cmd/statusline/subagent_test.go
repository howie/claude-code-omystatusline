package main

import (
	"strings"
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/context"
	"github.com/howie/claude-code-omystatusline/pkg/terminal"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantMode runMode
		wantErr  bool
	}{
		{"no-args-main", nil, modeMain, false},
		{"subagent-flag", []string{"--subagent"}, modeSubagent, false},
		{"unknown-flag", []string{"--bogus"}, modeMain, true},
		{"unexpected-positional", []string{"--subagent", "extra"}, modeMain, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mode, err := parseArgs(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseArgs(%v) err = %v, wantErr %v", tc.args, err, tc.wantErr)
			}
			if err == nil && mode != tc.wantMode {
				t.Errorf("parseArgs(%v) mode = %v, want %v", tc.args, mode, tc.wantMode)
			}
		})
	}
}

func TestRenderSubagentJSON(t *testing.T) {
	context.RenderMode = terminal.ModeASCII

	t.Run("multi-task", func(t *testing.T) {
		raw := []byte(`{"columns":80,"tasks":[
			{"name":"first","status":"running","model":"claude-haiku-4-5","tokenCount":1000},
			{"name":"second","status":"completed","model":"claude-haiku-4-5","tokenCount":2000}]}`)
		out, err := renderSubagentJSON(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Count(out, "\n") != 1 {
			t.Errorf("expected 2 lines, got %q", out)
		}
	})

	t.Run("empty-tasks-zero-bytes", func(t *testing.T) {
		out, err := renderSubagentJSON([]byte(`{"columns":80,"tasks":[]}`))
		if err != nil || out != "" {
			t.Errorf("expected zero bytes, got (%q, %v)", out, err)
		}
	})

	t.Run("malformed-json-errors", func(t *testing.T) {
		if _, err := renderSubagentJSON([]byte(`{bad`)); err == nil {
			t.Error("expected error for malformed JSON")
		}
	})

	t.Run("main-schema-is-safe", func(t *testing.T) {
		// Feeding the flat main-line payload to the subagent path must not error or
		// render garbage — no tasks -> zero bytes.
		raw := []byte(`{"model":{"id":"claude-opus-4-8"},"workspace":{"current_dir":"/x"}}`)
		out, err := renderSubagentJSON(raw)
		if err != nil || out != "" {
			t.Errorf("main-schema should be safe, got (%q, %v)", out, err)
		}
	})
}
