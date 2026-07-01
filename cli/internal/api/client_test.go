package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadyChecksReadinessEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ready" {
			t.Fatalf("path = %s, want /api/v1/ready", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready","database":"ok"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.Ready(context.Background())
	if err != nil {
		t.Fatalf("Ready returned error: %v", err)
	}
	if got.Status != "ready" || got.Database != "ok" {
		t.Fatalf("readiness = %+v", got)
	}
}

func TestReadyRejectsUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"status":"starting","database":"ok"}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	_, err := c.Ready(context.Background())
	if err == nil {
		t.Fatal("Ready returned nil error")
	}
}

// TestListAccounts_WireFormat verifies that ListAccounts decodes a raw
// JSON array from the server. The server's Huma output type is
// ListAccountsOutput { Body []AccountResponse } and Huma flattens the
// Body field on the wire, so the response is `[...]`, not
// `{body: [...]}`.
//
// This is a regression guard. A previous version of this client
// decoded `{body: [...]}` and failed with "cannot unmarshal array
// into Go value of type struct { Body []SocialAccount }".
func TestListAccounts_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"acc_1","slug":"x","platform":"x","account_id":"x_handle","account_username":"@rodrigo","is_active":true},
			{"id":"acc_2","slug":"bluesky","platform":"bluesky","account_id":"did:plc:abc","account_username":"rodrigo.bsky.social","is_active":true}
		]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListAccounts(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("ListAccounts returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(got))
	}
	if got[0].ID != "acc_1" || got[0].Slug != "x" || got[0].Platform != "x" {
		t.Errorf("account[0] wrong: %+v", got[0])
	}
	if got[1].ID != "acc_2" || got[1].Slug != "bluesky" || got[1].Platform != "bluesky" {
		t.Errorf("account[1] wrong: %+v", got[1])
	}
}

func TestUpdateAccount_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/api/v1/accounts/acc_1" {
			t.Errorf("path = %s, want /api/v1/accounts/acc_1", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["slug"] != "main-x" {
			t.Errorf("slug body = %q, want main-x", body["slug"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"acc_1","slug":"main-x","platform":"x","account_id":"x_handle","account_username":"@rodrigo","is_active":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.UpdateAccount(context.Background(), "acc_1", UpdateAccountInput{Slug: "main-x"})
	if err != nil {
		t.Fatalf("UpdateAccount returned error: %v", err)
	}
	if got.ID != "acc_1" || got.Slug != "main-x" || got.Platform != "x" {
		t.Errorf("account wrong: %+v", got)
	}
}

func TestListSets_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("workspace_id") != "ws_1" {
			t.Errorf("workspace_id = %q, want ws_1", r.URL.Query().Get("workspace_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":"set_1",
			"workspace_id":"ws_1",
			"name":"Launch",
			"is_default":true,
			"created_at":"2026-06-16T10:00:00Z",
			"accounts":[{"social_account_id":"acc_1","platform":"x","account_username":"rodrigo","is_main":false}]
		}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListSets(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("ListSets returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 set, got %d", len(got))
	}
	if got[0].ID != "set_1" || got[0].Name != "Launch" || !got[0].IsDefault {
		t.Errorf("set wrong: %+v", got[0])
	}
	if len(got[0].Accounts) != 1 || got[0].Accounts[0].SocialAccountID != "acc_1" {
		t.Errorf("accounts wrong: %+v", got[0].Accounts)
	}
}

func TestSetMutations_WireFormat(t *testing.T) {
	var sawCreate, sawUpdate, sawAdd, sawRemove, sawDelete bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sets":
			sawCreate = true
			var body CreateSetInput
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.WorkspaceID != "ws_1" || body.Name != "Launch" || !body.IsDefault || len(body.AccountIDs) != 1 || body.AccountIDs[0] != "acc_1" {
				t.Errorf("create body wrong: %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":"set_1","workspace_id":"ws_1","name":"Launch","is_default":true,"created_at":"2026-06-16T10:00:00Z"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/sets/set_1":
			sawUpdate = true
			var body UpdateSetInput
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body.Name == nil || *body.Name != "Renamed" {
				t.Errorf("update body wrong: %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":"set_1","workspace_id":"ws_1","name":"Renamed","is_default":true,"created_at":"2026-06-16T10:00:00Z"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sets/set_1/accounts":
			sawAdd = true
			var body AddSetAccountsInput
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode add body: %v", err)
			}
			if len(body.AccountIDs) != 1 || body.AccountIDs[0] != "acc_2" {
				t.Errorf("add body wrong: %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":"set_1","workspace_id":"ws_1","name":"Renamed","is_default":true,"created_at":"2026-06-16T10:00:00Z","accounts":[{"social_account_id":"acc_2","platform":"x"}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/sets/set_1/accounts/acc_2":
			sawRemove = true
			_, _ = w.Write([]byte(`{"id":"set_1","workspace_id":"ws_1","name":"Renamed","is_default":true,"created_at":"2026-06-16T10:00:00Z","accounts":[]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/sets/set_1":
			sawDelete = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	if _, err := c.CreateSet(context.Background(), CreateSetInput{WorkspaceID: "ws_1", Name: "Launch", IsDefault: true, AccountIDs: []string{"acc_1"}}); err != nil {
		t.Fatalf("CreateSet returned error: %v", err)
	}
	name := "Renamed"
	if _, err := c.UpdateSet(context.Background(), "set_1", UpdateSetInput{Name: &name}); err != nil {
		t.Fatalf("UpdateSet returned error: %v", err)
	}
	if _, err := c.AddSetAccounts(context.Background(), "set_1", AddSetAccountsInput{AccountIDs: []string{"acc_2"}}); err != nil {
		t.Fatalf("AddSetAccounts returned error: %v", err)
	}
	if _, err := c.RemoveSetAccount(context.Background(), "set_1", "acc_2"); err != nil {
		t.Fatalf("RemoveSetAccount returned error: %v", err)
	}
	if err := c.DeleteSet(context.Background(), "set_1"); err != nil {
		t.Fatalf("DeleteSet returned error: %v", err)
	}
	if !sawCreate || !sawUpdate || !sawAdd || !sawRemove || !sawDelete {
		t.Fatalf("missing request create=%t update=%t add=%t remove=%t delete=%t", sawCreate, sawUpdate, sawAdd, sawRemove, sawDelete)
	}
}

// TestListMedia_WireFormat verifies that ListMedia decodes the
// server's `{media: [...], total: N}` shape, not a `{body: {media,
// total}}` envelope.
func TestListMedia_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"media":[{"id":"m_1","workspace_id":"ws_1","mime_type":"image/png","size":1024,"original_filename":"x.png","width":800,"height":600,"alt_text":"","is_favorite":false,"created_at":"2026-06-15T10:00:00Z","url":"/media/m_1","thumbnail_url":"/media/m_1/thumb/sm","usage_count":0,"can_delete":true,"processing_status":"ready"}],"total":1}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListMedia(context.Background(), "ws_1", 50)
	if err != nil {
		t.Fatalf("ListMedia returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(got))
	}
	if got[0].ID != "m_1" || got[0].URL != "/media/m_1" {
		t.Errorf("media[0] wrong: %+v", got[0])
	}
}

// TestListMedia_EmptyResponse_DoesNotSilentlySucceed guards against
// the prior bug where the client decoded `{media: null, total: 0}`
// into `{body: {media, total}}`, which silently produced a nil
// slice — making the user believe there was no media when in fact
// the response was just missing the `body` wrapper. With the fix,
// the decode now matches the wire format and returns the empty
// list directly.
func TestListMedia_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"media":[],"total":0}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListMedia(context.Background(), "ws_1", 50)
	if err != nil {
		t.Fatalf("ListMedia returned error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 media items, got %d", len(got))
	}
}

func TestPublicationEndpoints_WireFormat(t *testing.T) {
	var sawCreate, sawList, sawGet, sawUpdate bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/publications":
			sawCreate = true
			var body CreatePublicationInput
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if body.WorkspaceID != "ws_1" || body.Title != "Launch" || body.SourceContent != "Ship the CLI publication flow" || len(body.MediaIDs) != 1 || body.MediaIDs[0] != "med_1" {
				t.Errorf("create body wrong: %+v", body)
			}
			_, _ = w.Write([]byte(`{"id":"pub_1","workspace_id":"ws_1","created_by":"u_1","title":"Launch","source_content":"Ship the CLI publication flow","status":"draft","release_plan_json":"{}","media_ids":["med_1"],"created_at":"2026-06-30T10:00:00Z","updated_at":"2026-06-30T10:00:00Z"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/publications":
			sawList = true
			if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
				t.Errorf("workspace_id = %q, want ws_1", got)
			}
			if got := r.URL.Query().Get("status"); got != "draft" {
				t.Errorf("status = %q, want draft", got)
			}
			if got := r.URL.Query().Get("limit"); got != "25" {
				t.Errorf("limit = %q, want 25", got)
			}
			if got := r.URL.Query().Get("offset"); got != "10" {
				t.Errorf("offset = %q, want 10", got)
			}
			_, _ = w.Write([]byte(`[{"id":"pub_1","workspace_id":"ws_1","created_by":"u_1","title":"Launch","source_content":"Ship the CLI publication flow","status":"draft","release_plan_json":"{}","media_ids":["med_1"],"created_at":"2026-06-30T10:00:00Z","updated_at":"2026-06-30T10:00:00Z"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/publications/pub_1":
			sawGet = true
			_, _ = w.Write([]byte(`{"id":"pub_1","workspace_id":"ws_1","created_by":"u_1","title":"Launch","source_content":"Ship the CLI publication flow","status":"draft","release_plan_json":"{}","media_ids":["med_1"],"created_at":"2026-06-30T10:00:00Z","updated_at":"2026-06-30T10:00:00Z"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/publications/pub_1":
			sawUpdate = true
			var body UpdatePublicationInput
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if body.Status == nil || *body.Status != "ready" {
				t.Errorf("status body wrong: %+v", body.Status)
			}
			if body.MediaIDs == nil || len(*body.MediaIDs) != 0 {
				t.Errorf("media_ids body wrong: %+v", body.MediaIDs)
			}
			_, _ = w.Write([]byte(`{"id":"pub_1","workspace_id":"ws_1","created_by":"u_1","title":"Launch","source_content":"Ship the CLI publication flow","status":"ready","release_plan_json":"{}","created_at":"2026-06-30T10:00:00Z","updated_at":"2026-06-30T11:00:00Z"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	if _, err := c.CreatePublication(context.Background(), CreatePublicationInput{WorkspaceID: "ws_1", Title: "Launch", SourceContent: "Ship the CLI publication flow", MediaIDs: []string{"med_1"}}); err != nil {
		t.Fatalf("CreatePublication returned error: %v", err)
	}
	listed, err := c.ListPublications(context.Background(), ListPublicationsInput{WorkspaceID: "ws_1", Status: "draft", Limit: 25, Offset: 10})
	if err != nil {
		t.Fatalf("ListPublications returned error: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "pub_1" || listed[0].MediaIDs[0] != "med_1" {
		t.Fatalf("listed publication wrong: %+v", listed)
	}
	got, err := c.GetPublication(context.Background(), "pub_1")
	if err != nil {
		t.Fatalf("GetPublication returned error: %v", err)
	}
	if got.ID != "pub_1" || got.Title != "Launch" {
		t.Fatalf("publication wrong: %+v", got)
	}
	status := "ready"
	emptyMedia := []string{}
	updated, err := c.UpdatePublication(context.Background(), "pub_1", UpdatePublicationInput{Status: &status, MediaIDs: &emptyMedia})
	if err != nil {
		t.Fatalf("UpdatePublication returned error: %v", err)
	}
	if updated.Status != "ready" {
		t.Fatalf("updated status = %q, want ready", updated.Status)
	}
	if !sawCreate || !sawList || !sawGet || !sawUpdate {
		t.Fatalf("missing request create=%t list=%t get=%t update=%t", sawCreate, sawList, sawGet, sawUpdate)
	}
}

// TestListPosts_WireFormat: server returns a raw array of posts.
func TestListPosts_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
			t.Fatalf("workspace_id query = %q, want ws_1", got)
		}
		if got := r.URL.Query().Get("status"); got != "scheduled" {
			t.Fatalf("status query = %q, want scheduled", got)
		}
		if got := r.URL.Query().Get("date"); got != "2026-06-16" {
			t.Fatalf("date query = %q, want 2026-06-16", got)
		}
		if got := r.URL.Query().Get("limit"); got != "25" {
			t.Fatalf("limit query = %q, want 25", got)
		}
		if got := r.URL.Query().Get("offset"); got != "50" {
			t.Fatalf("offset query = %q, want 50", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"p_1","workspace_id":"ws_1","created_by":"u_1","content":"Hello","status":"scheduled","scheduled_at":"2026-06-16T09:00:00Z","created_at":"2026-06-15T10:00:00Z","random_delay_minutes":0}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListPosts(context.Background(), ListPostsInput{WorkspaceID: "ws_1", Status: "scheduled", Date: "2026-06-16", Limit: 25, Offset: 50})
	if err != nil {
		t.Fatalf("ListPosts returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 post, got %d", len(got))
	}
	if got[0].ID != "p_1" || got[0].Content != "Hello" {
		t.Errorf("post wrong: %+v", got[0])
	}
}

// TestListJobs_WireFormat: server returns a raw array of jobs.
func TestListJobs_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
			t.Fatalf("workspace_id query = %q, want ws_1", got)
		}
		if got := r.URL.Query().Get("status"); got != "pending" {
			t.Fatalf("status query = %q, want pending", got)
		}
		if got := r.URL.Query().Get("limit"); got != "25" {
			t.Fatalf("limit query = %q, want 25", got)
		}
		if got := r.URL.Query().Get("offset"); got != "50" {
			t.Fatalf("offset query = %q, want 50", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"j_1","type":"publish","payload":"{}","status":"queued","run_at":"2026-06-16T09:00:00Z","attempts":0}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListJobs(context.Background(), ListJobsInput{WorkspaceID: "ws_1", Status: "pending", Limit: 25, Offset: 50})
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 job, got %d", len(got))
	}
	if got[0].ID != "j_1" || got[0].Type != "publish" {
		t.Errorf("job wrong: %+v", got[0])
	}
}

// TestGetWorkspaceSettings_WireFormat: server returns a flat object.
func TestGetWorkspaceSettings_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"timezone":"Europe/Lisbon","week_start":1,"media_cleanup_days":30,"random_delay_minutes":5,"draft_gap_minutes":60,"slot_start_hour":9,"slot_end_hour":18,"slot_interval_minutes":30}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.GetWorkspaceSettings(context.Background(), "ws_1")
	if err != nil {
		t.Fatalf("GetWorkspaceSettings returned error: %v", err)
	}
	if got.Timezone != "Europe/Lisbon" {
		t.Errorf("expected timezone Europe/Lisbon, got %q", got.Timezone)
	}
	if got.WeekStart != 1 {
		t.Errorf("expected week_start 1, got %d", got.WeekStart)
	}
}

func TestNextAvailableSlot_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/posting-schedules/next-slot" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("workspace_id"); got != "ws_1" {
			t.Errorf("workspace_id = %q, want ws_1", got)
		}
		if got := r.URL.Query().Get("set_id"); got != "set_1" {
			t.Errorf("set_id = %q, want set_1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"slot":{"id":"slot_1","workspace_id":"ws_1","set_id":"set_1","utc_hour":9,"utc_minute":0,"day_of_week":2,"local_hour":9,"local_minute":0,"local_day_of_week":2,"label":"Morning","is_active":true,"created_at":"2026-06-16T08:00:00Z"},
			"slot_time":"2026-06-16T09:00:00Z",
			"message":"Next available slot found"
		}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.NextAvailableSlot(context.Background(), NextAvailableSlotInput{WorkspaceID: "ws_1", SetID: "set_1"})
	if err != nil {
		t.Fatalf("NextAvailableSlot returned error: %v", err)
	}
	if got.SlotTime != "2026-06-16T09:00:00Z" {
		t.Fatalf("slot_time = %q", got.SlotTime)
	}
	if got.Slot == nil || got.Slot.ID != "slot_1" || got.Slot.SetID != "set_1" {
		t.Fatalf("slot = %+v", got.Slot)
	}
}

// TestCreatePost_WireFormat: server returns the Post object directly.
func TestCreatePost_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p_new","workspace_id":"ws_1","created_by":"u_1","publication_id":"pub_1","content":"Hi","status":"draft","scheduled_at":"","created_at":"2026-06-15T10:00:00Z","random_delay_minutes":0}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.CreatePost(context.Background(), CreatePostInput{
		WorkspaceID:      "ws_1",
		Content:          "Hi",
		SocialAccountIDs: []string{"acc_1"},
	})
	if err != nil {
		t.Fatalf("CreatePost returned error: %v", err)
	}
	if got.ID != "p_new" {
		t.Errorf("expected id p_new, got %q", got.ID)
	}
	if got.Content != "Hi" {
		t.Errorf("expected content Hi, got %q", got.Content)
	}
	if got.PublicationID != "pub_1" {
		t.Errorf("expected publication_id pub_1, got %q", got.PublicationID)
	}
}

// TestCreateAPIToken_WireFormat: server returns `{token, item}`.
func TestCreateAPIToken_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"op_cli_abc_secret","item":{"id":"t_1","name":"laptop","token_prefix":"op_cli_","scope":"cli:full","created_at":"2026-06-15T10:00:00Z"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.CreateAPIToken(context.Background(), CreateAPITokenInput{Name: "laptop"})
	if err != nil {
		t.Fatalf("CreateAPIToken returned error: %v", err)
	}
	if got.RawToken != "op_cli_abc_secret" {
		t.Errorf("expected raw token, got %q", got.RawToken)
	}
	if got.Item.Name != "laptop" {
		t.Errorf("expected item.name laptop, got %q", got.Item.Name)
	}
}
