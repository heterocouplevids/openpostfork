package commands

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPublicationContentRequiresSource(t *testing.T) {
	_, err := publicationContent("", "")
	if err == nil || !strings.Contains(err.Error(), "--content or --file is required") {
		t.Fatalf("error = %v, want missing content error", err)
	}
}

func TestPublicationUpdateChanged(t *testing.T) {
	cmd := &cobra.Command{}
	flags := publicationFlags{}
	cmd.Flags().StringVar(&flags.title, "title", "", "")
	cmd.Flags().StringVar(&flags.content, "content", "", "")
	cmd.Flags().StringVar(&flags.file, "file", "", "")
	cmd.Flags().StringVar(&flags.sourceURL, "source-url", "", "")
	cmd.Flags().StringVar(&flags.goal, "goal", "", "")
	cmd.Flags().StringVar(&flags.audience, "audience", "", "")
	cmd.Flags().StringVar(&flags.status, "status", "", "")
	cmd.Flags().StringArrayVar(&flags.media, "media", nil, "")

	if publicationUpdateChanged(cmd, false) {
		t.Fatal("publicationUpdateChanged without changed flags = true, want false")
	}
	if err := cmd.Flags().Set("goal", "launch"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	if !publicationUpdateChanged(cmd, false) {
		t.Fatal("publicationUpdateChanged with changed goal = false, want true")
	}
	clearCmd := &cobra.Command{}
	clearFlags := publicationFlags{}
	clearCmd.Flags().StringVar(&clearFlags.title, "title", "", "")
	clearCmd.Flags().StringVar(&clearFlags.content, "content", "", "")
	clearCmd.Flags().StringVar(&clearFlags.file, "file", "", "")
	clearCmd.Flags().StringVar(&clearFlags.sourceURL, "source-url", "", "")
	clearCmd.Flags().StringVar(&clearFlags.goal, "goal", "", "")
	clearCmd.Flags().StringVar(&clearFlags.audience, "audience", "", "")
	clearCmd.Flags().StringVar(&clearFlags.status, "status", "", "")
	clearCmd.Flags().StringArrayVar(&clearFlags.media, "media", nil, "")
	if !publicationUpdateChanged(clearCmd, true) {
		t.Fatal("publicationUpdateChanged with clearMedia = false, want true")
	}
}

func TestJoinOrDash(t *testing.T) {
	if got := joinOrDash(nil); got != "-" {
		t.Fatalf("joinOrDash(nil) = %q, want -", got)
	}
	if got := joinOrDash([]string{"med_1", "med_2"}); got != "med_1,med_2" {
		t.Fatalf("joinOrDash = %q", got)
	}
}
