package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

// TestListPosts_WireFormat: server returns a raw array of posts.
func TestListPosts_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"p_1","workspace_id":"ws_1","created_by":"u_1","content":"Hello","status":"scheduled","scheduled_at":"2026-06-16T09:00:00Z","created_at":"2026-06-15T10:00:00Z","random_delay_minutes":0}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListPosts(context.Background(), ListPostsInput{WorkspaceID: "ws_1"})
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"j_1","type":"publish","payload":"{}","status":"queued","run_at":"2026-06-16T09:00:00Z","attempts":0}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "")
	got, err := c.ListJobs(context.Background(), ListJobsInput{WorkspaceID: "ws_1"})
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

// TestCreatePost_WireFormat: server returns the Post object directly.
func TestCreatePost_WireFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p_new","workspace_id":"ws_1","created_by":"u_1","content":"Hi","status":"draft","scheduled_at":"","created_at":"2026-06-15T10:00:00Z","random_delay_minutes":0}`))
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
