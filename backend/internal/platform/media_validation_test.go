package platform

import "testing"

func TestValidateMediaBlueskyRejectsMixedVideo(t *testing.T) {
	RegisterAllMediaValidators()
	issues := ValidateMedia(providerBluesky, []MediaItem{
		{ID: "video", MimeType: "video/mp4"},
		{ID: "image", MimeType: "image/png"},
	})

	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Severity != severityError {
		t.Fatalf("expected error severity, got %q", issues[0].Severity)
	}
}

func TestValidateMediaThreadsUsesMimeTypeForVideo(t *testing.T) {
	issues := ValidateMedia(providerThreads, []MediaItem{
		{ID: "video-without-extension", MimeType: "video/mp4"},
	})

	if len(issues) != 0 {
		t.Fatalf("expected no issues for mp4 video, got %#v", issues)
	}
}

func TestValidateMediaLinkedInWarnsForMultipleAttachments(t *testing.T) {
	issues := ValidateMedia(providerLinkedIn, []MediaItem{
		{ID: "first", MimeType: "image/png"},
		{ID: "second", MimeType: "image/png"},
	})

	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Severity != severityWarning {
		t.Fatalf("expected warning severity, got %q", issues[0].Severity)
	}
}

func TestValidateMediaTikTokRequiresOneVideo(t *testing.T) {
	RegisterAllMediaValidators()

	issues := ValidateMedia(providerTikTok, []MediaItem{{ID: "image", MimeType: "image/png"}})
	if len(issues) != 1 {
		t.Fatalf("expected one image issue, got %d", len(issues))
	}
	if issues[0].Severity != severityError {
		t.Fatalf("expected error severity, got %q", issues[0].Severity)
	}

	issues = ValidateMedia(providerTikTok, []MediaItem{{ID: "video", MimeType: "video/mp4"}})
	if len(issues) != 0 {
		t.Fatalf("expected no issues for one video, got %#v", issues)
	}
}

func TestValidateMediaFacebookRejectsMultipleAttachments(t *testing.T) {
	RegisterAllMediaValidators()

	issues := ValidateMedia(providerFacebook, []MediaItem{
		{ID: "first", MimeType: "image/png"},
		{ID: "second", MimeType: "image/png"},
	})
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Severity != severityError {
		t.Fatalf("expected error severity, got %q", issues[0].Severity)
	}

	issues = ValidateMedia(providerFacebook, []MediaItem{{ID: "image", MimeType: "image/png"}})
	if len(issues) != 0 {
		t.Fatalf("expected no issues for one attachment, got %#v", issues)
	}
}

func TestValidateMediaInstagramRequiresOneImageOrVideo(t *testing.T) {
	RegisterAllMediaValidators()

	issues := ValidateMedia(providerInstagram, nil)
	if len(issues) != 1 {
		t.Fatalf("expected one missing-media issue, got %d", len(issues))
	}

	issues = ValidateMedia(providerInstagram, []MediaItem{{ID: "file", MimeType: "application/pdf"}})
	if len(issues) != 1 {
		t.Fatalf("expected one unsupported-media issue, got %d", len(issues))
	}

	issues = ValidateMedia(providerInstagram, []MediaItem{{ID: "image", MimeType: "image/png"}})
	if len(issues) != 0 {
		t.Fatalf("expected no issues for one image, got %#v", issues)
	}
}
