package platform

import "testing"

func TestValidateMediaBlueskyRejectsMixedVideo(t *testing.T) {
	RegisterAllMediaValidators()
	issues := ValidateMedia("bluesky", []MediaItem{
		{ID: "video", MimeType: "video/mp4"},
		{ID: "image", MimeType: "image/png"},
	})

	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Severity != "error" {
		t.Fatalf("expected error severity, got %q", issues[0].Severity)
	}
}

func TestValidateMediaThreadsUsesMimeTypeForVideo(t *testing.T) {
	issues := ValidateMedia("threads", []MediaItem{
		{ID: "video-without-extension", MimeType: "video/mp4"},
	})

	if len(issues) != 0 {
		t.Fatalf("expected no issues for mp4 video, got %#v", issues)
	}
}

func TestValidateMediaLinkedInWarnsForMultipleAttachments(t *testing.T) {
	issues := ValidateMedia("linkedin", []MediaItem{
		{ID: "first", MimeType: "image/png"},
		{ID: "second", MimeType: "image/png"},
	})

	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Severity != "warning" {
		t.Fatalf("expected warning severity, got %q", issues[0].Severity)
	}
}
