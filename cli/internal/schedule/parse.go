// Package schedule parses human-friendly --schedule inputs and returns
// a time.Time in the user's resolved timezone.
//
// The precedence for choosing a timezone on ambiguous input is:
//   1. explicit offset/zone in the string itself (e.g. "tomorrow 2pm Europe/Lisbon" or RFC3339 with offset)
//   2. --timezone flag supplied to the command
//   3. workspace timezone (caller passes it in Options.WorkspaceTimezone)
//   4. profile timezone (config)
//   5. local machine timezone (last-resort fallback)
//
// Ambiguous input is treated as an error rather than a silent guess:
//   - "tomorrow"        → error (no time)
//   - "2pm"             → today at 14:00 if future, else tomorrow at 14:00
//   - "03/04/2026 ..."  → parsed as day-first (DD/MM/YYYY) and a warning printed
//
// The parser is intentionally conservative: for a CLI that controls
// real posts, an unexplained guess is worse than a clear error.
package schedule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/markusmobius/go-dateparser"
)

// Options is the input to Parse.
type Options struct {
	Now                time.Time     // when the user typed it (for "2pm")
	DefaultLocation    *time.Location // flag or profile timezone
	WorkspaceTimezone  string         // IANA name; takes precedence over DefaultLocation if non-empty
	AllowPast          bool           // permit times in the past (for testing/back-dating)
}

// Result is the output of Parse.
type Result struct {
	Time    time.Time
	Source  string // "absolute" | "natural" | "alias"
	Original string
	Warning string
}

// Pre-defined aliases. "now" schedules for the next minute so it ends
// up in the same publish-at-time slot as "schedule for now".
var aliases = map[string]time.Duration{
	"now":   0,
	"draft": -1, // sentinel: caller interprets as "no schedule" (omits scheduled_at)
}

// RFC3339 with offset is the strictest; try it first.
var rfc3339 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?(Z|[+-]\d{2}:?\d{2})$`)

// Parse resolves the input. Returns an error for unparseable or
// missing-time inputs unless `now`/`draft`/RFC3339 is supplied.
func Parse(input string, opts Options) (Result, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Result{}, errors.New("empty schedule string")
	}

	// 1) Draft sentinel: caller should treat as no schedule.
	if _, ok := aliases[strings.ToLower(input)]; ok && strings.ToLower(input) == "draft" {
		return Result{Source: "alias", Original: input}, nil
	}

	// 2) "now" → the next 1-minute boundary from now (so a publish
	// worker can pick it up on its next tick).
	if strings.EqualFold(input, "now") {
		now := opts.Now
		if now.IsZero() {
			now = time.Now()
		}
		return Result{Time: now.Truncate(time.Minute).Add(time.Minute), Source: "alias", Original: input}, nil
	}

	// 3) Strict RFC3339.
	if rfc3339.MatchString(input) {
		t, err := time.Parse(time.RFC3339, input)
		if err != nil {
			return Result{}, fmt.Errorf("invalid RFC3339 timestamp: %w", err)
		}
		if !opts.AllowPast && t.Before(opts.Now) && !opts.Now.IsZero() {
			return Result{}, fmt.Errorf("scheduled time %s is in the past", t.Format(time.RFC3339))
		}
		return Result{Time: t, Source: "absolute", Original: input}, nil
	}

	// 4) Common absolute formats (no timezone — use resolved loc).
	absoluteLayouts := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-01-2006 15:04",
		"02/01/2006 15:04",
		"02-01-2006",
		"02/01/2006",
	}
	loc := pickLocation(opts)
	for _, layout := range absoluteLayouts {
		if t, err := time.ParseInLocation(layout, input, loc); err == nil {
			if !opts.AllowPast && t.Before(opts.Now) && !opts.Now.IsZero() {
				return Result{}, fmt.Errorf("scheduled time %s is in the past", t.Format(time.RFC3339))
			}
			return Result{Time: t, Source: "absolute", Original: input}, nil
		}
	}

	// 5) Natural-language: "tomorrow 2pm", "in 3 hours", "next monday 9am".
	cfg := &dateparser.Config{
		CurrentTime:        opts.Now,
		DefaultTimezone:    loc,
		PreferFuture:       true,
		DetectFutureTime:   true,
	}
	t, err := dateparser.Parse(cfg, input)
	if err != nil {
		return Result{}, fmt.Errorf("could not parse %q: %w", input, err)
	}

	// "tomorrow" with no time is a common mistake: reject it.
	if lower := strings.ToLower(input); lower == "tomorrow" || lower == "today" {
		return Result{}, fmt.Errorf("%q has no time; use something like --schedule \"%s 2pm\"", input, lower)
	}

	if !opts.AllowPast && t.Before(opts.Now) && !opts.Now.IsZero() {
		return Result{}, fmt.Errorf("scheduled time %s is in the past", t.Format(time.RFC3339))
	}
	return Result{Time: t, Source: "natural", Original: input}, nil
}

func pickLocation(opts Options) *time.Location {
	if opts.WorkspaceTimezone != "" {
		if loc, err := time.LoadLocation(opts.WorkspaceTimezone); err == nil {
			return loc
		}
	}
	if opts.DefaultLocation != nil {
		return opts.DefaultLocation
	}
	return time.Local
}

// FormatHuman is a small helper used by post/thread commands when
// confirming a parsed time to the user.
func FormatHuman(t time.Time) string {
	if t.IsZero() {
		return "draft"
	}
	return t.Format("Mon, 02 Jan 2006 15:04 MST")
}

// MustLoadLocation wraps time.LoadLocation with a sensible fallback.
func MustLoadLocation(name string) *time.Location {
	if name == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

// Helper for unit tests: convert a HH:MM string to minutes.
func parseHHMM(s string) (int, int, bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}
