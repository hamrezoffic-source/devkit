package main

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "object", input: `{"b":2,"a":[1,true,null]}`, want: "{\n  \"a\": [\n    1,\n    true,\n    null\n  ],\n  \"b\": 2\n}\n"},
		{name: "large integer remains exact", input: `{"n":9007199254740993}`, want: "9007199254740993", wantErr: false},
		{name: "empty object", input: `{}`, want: "{}\n"},
		{name: "trailing data", input: `{} {}`, wantErr: true},
		{name: "malformed", input: `{"a":}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("FormatJSON() expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("FormatJSON() error = %v", err)
			}
			if tt.name == "large integer remains exact" {
				if !bytes.Contains(got, []byte(tt.want)) {
					t.Fatalf("FormatJSON() = %s, want exact integer %s", got, tt.want)
				}
				return
			}
			if string(got) != tt.want {
				t.Fatalf("FormatJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONYAMLRoundTrip(t *testing.T) {
	input := []byte(`{"name":"devkit","enabled":true,"count":3,"items":["a",null],"nested":{"ok":false}}`)
	yamlOutput, err := JSONToYAML(input)
	if err != nil {
		t.Fatalf("JSONToYAML() error = %v", err)
	}
	jsonOutput, err := YAMLToJSON(yamlOutput)
	if err != nil {
		t.Fatalf("YAMLToJSON() error = %v", err)
	}
	for _, want := range []string{"devkit", "enabled", "count", "items", "nested"} {
		if !bytes.Contains(jsonOutput, []byte(want)) {
			t.Fatalf("YAMLToJSON() = %s, missing %q", jsonOutput, want)
		}
	}
}

func TestYAMLToJSONEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "null and booleans", input: "a: null\nb: true\nc: false\n", want: "null"},
		{name: "sequence", input: "- red\n- green\n- blue\n", want: "red"},
		{name: "quoted string", input: "value: \"0123\"\n", want: "0123"},
		{name: "empty document", input: "", want: "null"},
		{name: "infinite float", input: "value: .inf\n", wantErr: "not representable"},
		{name: "non string key", input: "1: value\n", wantErr: "not a string"},
		{name: "duplicate key", input: "a: 1\na: 2\n", wantErr: "duplicate"},
		{name: "malformed", input: "a: [one, two\n", wantErr: "invalid YAML"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YAMLToJSON([]byte(tt.input))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("YAMLToJSON() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("YAMLToJSON() error = %v", err)
			}
			if !bytes.Contains(got, []byte(tt.want)) {
				t.Fatalf("YAMLToJSON() = %s, want containing %q", got, tt.want)
			}
		})
	}
}

func TestCheckTCPPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot listen: %v", err)
	}
	defer listener.Close()

	done := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
		close(done)
	}()
	port := listener.Addr().(*net.TCPAddr).Port
	if err := CheckTCPPort("127.0.0.1", port, time.Second); err != nil {
		t.Fatalf("CheckTCPPort() error = %v", err)
	}
	<-done

	if err := CheckTCPPort("127.0.0.1", 1, 20*time.Millisecond); err == nil {
		t.Fatal("CheckTCPPort() unexpectedly succeeded")
	}
}

func TestCLIEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("subprocess executable path is covered by the build matrix")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "devkit")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = "."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	tests := []struct {
		name    string
		args    []string
		input   string
		want    string
		wantErr bool
	}{
		{name: "format stdin", args: []string{"json", "format", "-"}, input: `{"ok":true}`, want: "\"ok\": true"},
		{name: "json to yaml alias", args: []string{"json-to-yaml", "-"}, input: `{"ok":true}`, want: "ok: true"},
		{name: "yaml to json", args: []string{"yaml", "to-json", "-"}, input: "ok: true\n", want: "\"ok\": true"},
		{name: "bad JSON", args: []string{"json", "format", "-"}, input: `{bad`, wantErr: true},
		{name: "help", args: []string{"--help"}, want: "devkit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			cmd.Stdin = strings.NewReader(tt.input)
			output, err := cmd.CombinedOutput()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("command unexpectedly succeeded: %s", output)
				}
				return
			}
			if err != nil {
				t.Fatalf("command failed: %v\n%s", err, output)
			}
			if !bytes.Contains(output, []byte(tt.want)) {
				t.Fatalf("output = %q, want containing %q", output, tt.want)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
