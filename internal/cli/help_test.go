package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNestedHelp(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr string
		code    int
	}{
		{
			name: "root",
			args: []string{"help"},
			want: []string{"akme help [command...]", "akme help git", "akme help token", "akme help update"},
		},
		{
			name: "update",
			args: []string{"help", "update"},
			want: []string{"akme update", "releases/latest", "AKME_NO_UPDATE"},
		},
		{
			name: "update help via command",
			args: []string{"update", "--help"},
			want: []string{"akme update", "writable install directory"},
		},
		{
			name: "git overview",
			args: []string{"help", "git"},
			want: []string{"akme git <folder>", "akme help git stat", "activity", "interactive"},
		},
		{
			name: "git activity",
			args: []string{"help", "git", "activity"},
			want: []string{"akme git <folder> [range] [view] [-i]", "first-parent", "52 weeks"},
		},
		{
			name: "git view",
			args: []string{"help", "git", "view"},
			want: []string{"overview", "punchcard", "branch"},
		},
		{
			name: "git range",
			args: []string{"help", "git", "range"},
			want: []string{"year", "30d", "7d"},
		},
		{
			name: "git interactive",
			args: []string{"help", "git", "interactive"},
			want: []string{"tab / ⇧tab", "AKME_BACKGROUND", "-i"},
		},
		{
			name: "git stat",
			args: []string{"help", "git", "stat"},
			want: []string{"akme git stat [--jobs <n>]", "--jobs <n>", "akme git stat ."},
		},
		{
			name: "git help via domain",
			args: []string{"git", "help"},
			want: []string{"akme git <folder>", "akme help git stat"},
		},
		{
			name: "git help stat via domain",
			args: []string{"git", "help", "stat"},
			want: []string{"akme git stat [--jobs <n>]", "origin/HEAD"},
		},
		{
			name: "token overview",
			args: []string{"help", "token"},
			want: []string{"akme token [harness]", "akme help token [topic]", "Nest help for one topic"},
		},
		{
			name: "token tokens alias",
			args: []string{"help", "tokens"},
			want: []string{"akme token [harness]", "Nest help for one topic"},
		},
		{
			name: "token harness topic",
			args: []string{"help", "token", "harness"},
			want: []string{"Claude Code", "~/.claude", "cursor-agent"},
		},
		{
			name: "token view topic",
			args: []string{"help", "token", "view"},
			want: []string{"punchcard", "weekday × hour"},
		},
		{
			name: "token topics index",
			args: []string{"help", "token", "topics"},
			want: []string{"akme help token [topic]", "harness", "examples"},
		},
		{
			name: "version help",
			args: []string{"help", "version"},
			want: []string{"akme version", "-v, --version"},
		},
		{
			name: "help about help",
			args: []string{"help", "help"},
			want: []string{"akme help [command...]", "Keep nesting"},
		},
		{
			name:    "unknown root topic",
			args:    []string{"help", "nope"},
			wantErr: "unknown help topic \"nope\"",
			code:    2,
		},
		{
			name:    "unknown git topic",
			args:    []string{"help", "git", "fetch"},
			wantErr: "unknown git help topic \"fetch\"",
			code:    2,
		},
		{
			name:    "unknown token topic",
			args:    []string{"help", "token", "widgets"},
			wantErr: "unknown token help topic \"widgets\"",
			code:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := New(BuildInfo{Version: "test"}).Run(context.Background(), tt.args, &stdout, &stderr)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				var exitErr ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("error %v is not an ExitError", err)
				}
				if exitErr.Code != tt.code {
					t.Fatalf("exit code = %d, want %d", exitErr.Code, tt.code)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := stdout.String()
			for _, needle := range tt.want {
				if !strings.Contains(got, needle) {
					t.Fatalf("stdout missing %q\n\n%s", needle, got)
				}
			}
		})
	}
}

func TestHelpForDirect(t *testing.T) {
	text, err := helpFor(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "akme help [command...]") {
		t.Fatalf("unexpected root help:\n%s", text)
	}
}
