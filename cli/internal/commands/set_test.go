package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
)

func TestFindSetByIDNameAndDefault(t *testing.T) {
	sets := []api.SocialSet{
		{ID: "set_1", Name: "Launch"},
		{ID: "set_2", Name: "Evergreen", IsDefault: true},
	}

	for _, selector := range []string{"set_1", "Launch"} {
		t.Run(selector, func(t *testing.T) {
			got, err := findSet(sets, selector)
			if err != nil {
				t.Fatalf("findSet returned error: %v", err)
			}
			if got.ID != "set_1" {
				t.Fatalf("set id = %q, want set_1", got.ID)
			}
		})
	}

	got, err := findSet(sets, "default")
	if err != nil {
		t.Fatalf("findSet default returned error: %v", err)
	}
	if got.ID != "set_2" {
		t.Fatalf("default set id = %q, want set_2", got.ID)
	}
}

func TestFindSetDefaultReportsMissingDefault(t *testing.T) {
	_, err := findSet([]api.SocialSet{{ID: "set_1", Name: "Launch"}}, "default")
	if err == nil || !strings.Contains(err.Error(), "no default") {
		t.Fatalf("error = %v, want missing default error", err)
	}
}

func TestDefaultSet(t *testing.T) {
	sets := []api.SocialSet{
		{ID: "set_1", Name: "Launch"},
		{ID: "set_2", Name: "Evergreen", IsDefault: true},
	}
	got := defaultSet(sets)
	if got == nil || got.ID != "set_2" {
		t.Fatalf("defaultSet = %+v, want set_2", got)
	}
	if got := defaultSet([]api.SocialSet{{ID: "set_1", Name: "Launch"}}); got != nil {
		t.Fatalf("defaultSet without default = %+v, want nil", got)
	}
}

func TestFormatSetAccounts(t *testing.T) {
	got := formatSetAccounts([]api.SetAccount{
		{SocialAccountID: "acc_1", Platform: "x", AccountUsername: "rodrigo"},
		{SocialAccountID: "acc_2", Platform: "bluesky"},
	})
	if got != "x:rodrigo,bluesky:acc_2" {
		t.Fatalf("formatSetAccounts = %q", got)
	}
	if got := formatSetAccounts(nil); got != "-" {
		t.Fatalf("empty formatSetAccounts = %q, want -", got)
	}
}

func TestResolveSocialTargetsUsesDefaultSet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/sets" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
			t.Fatalf("workspace_id = %q, want ws_1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":"set_1",
			"workspace_id":"ws_1",
			"name":"Launch",
			"is_default":true,
			"accounts":[
				{"social_account_id":"acc_1","platform":"x"},
				{"social_account_id":"acc_2","platform":"linkedin"},
				{"social_account_id":"acc_1","platform":"x"}
			]
		}]`))
	}))
	defer srv.Close()

	client := api.New(srv.URL, "")
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	got, err := resolveSocialTargets(cmd, client, "ws_1", "", "", true)
	if err != nil {
		t.Fatalf("resolveSocialTargets returned error: %v", err)
	}
	if got.SetID != "set_1" {
		t.Fatalf("set id = %q, want set_1", got.SetID)
	}
	if strings.Join(got.AccountIDs, ",") != "acc_1,acc_2" {
		t.Fatalf("targets = %#v, want acc_1,acc_2", got)
	}
}

func TestResolveSocialTargetsRejectsAccountsAndSetTogether(t *testing.T) {
	client := api.New("https://openpost.test", "")
	_, err := resolveSocialTargets(&cobra.Command{}, client, "ws_1", "x", "launch", true)
	if err == nil || !strings.Contains(err.Error(), "either --accounts or --set") {
		t.Fatalf("error = %v, want mutual exclusion error", err)
	}
}

func TestIsNextSlotSchedule(t *testing.T) {
	for _, raw := range []string{"next-slot", "next slot", "slot"} {
		t.Run(raw, func(t *testing.T) {
			if !isNextSlotSchedule(raw) {
				t.Fatalf("%q should be a next-slot alias", raw)
			}
		})
	}
	if isNextSlotSchedule("now") {
		t.Fatal("now should not be a next-slot alias")
	}
}

func TestParseScheduleFlagNextSlot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
			t.Fatalf("workspace_id = %q, want ws_1", got)
		}
		if got := r.URL.Query().Get("set_id"); got != "set_1" {
			t.Fatalf("set_id = %q, want set_1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slot_time":"2026-06-16T09:00:00Z","message":"Next available slot found"}`))
	}))
	defer srv.Close()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	got, source, err := parseScheduleFlag(cmd, api.New(srv.URL, ""), "ws_1", "set_1", "next-slot", "UTC")
	if err != nil {
		t.Fatalf("parseScheduleFlag returned error: %v", err)
	}
	if source != "slot" {
		t.Fatalf("source = %q, want slot", source)
	}
	want := "2026-06-16T09:00:00Z"
	if got == nil || got.Format(time.RFC3339) != want {
		t.Fatalf("time = %v, want %s", got, want)
	}
}

func TestParseScheduleFlagNextSlotRequiresAvailableSlot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slot_time":"","message":"No posting schedules configured for this workspace"}`))
	}))
	defer srv.Close()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	_, _, err := parseScheduleFlag(cmd, api.New(srv.URL, ""), "ws_1", "", "slot", "UTC")
	if err == nil || !strings.Contains(err.Error(), "no next posting slot available") {
		t.Fatalf("error = %v, want no slot error", err)
	}
}
