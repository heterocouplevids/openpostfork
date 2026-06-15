package commands

import (
	"strings"
	"testing"
)

func TestParseThreadMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   threadFrontMatter
		wantBody []string
		wantErr  string
	}{
		{
			name:  "front-matter plus three posts",
			input: "---\nworkspace: personal\naccounts: x,linkedin\nschedule: tomorrow 2pm\nrandom_delay: 7\n---\nOne\n---\nTwo\n---\nThree\n",
			wantFM: threadFrontMatter{
				Workspace:   "personal",
				Accounts:    "x,linkedin",
				Schedule:    "tomorrow 2pm",
				RandomDelay: 7,
			},
			wantBody: []string{"One", "Two", "Three"},
		},
		{
			name:     "no front-matter plus two posts",
			input:    "One\n---\nTwo\n",
			wantBody: []string{"One", "Two"},
		},
		{
			name:     "embedded dashes inside post body",
			input:    "One --- still one\n---\nTwo\n",
			wantBody: []string{"One --- still one", "Two"},
		},
		{
			name:    "empty segment rejected",
			input:   "One\n---\n \n---\nThree\n",
			wantErr: "thread segment 2 is empty",
		},
		{
			name:     "mixed CRLF and LF",
			input:    "One\r\n---\nTwo\r\n---\nThree",
			wantBody: []string{"One", "Two", "Three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFM, gotBody, err := parseThreadMarkdown(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFM != tt.wantFM {
				t.Fatalf("frontmatter = %#v, want %#v", gotFM, tt.wantFM)
			}
			if strings.Join(gotBody, "\n---\n") != strings.Join(tt.wantBody, "\n---\n") {
				t.Fatalf("segments = %#v, want %#v", gotBody, tt.wantBody)
			}
		})
	}
}
