package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
	"github.com/openpost/cli/internal/config"
)

type publicationFlags struct {
	title      string
	content    string
	file       string
	sourceURL  string
	goal       string
	audience   string
	status     string
	media      []string
	mediaAlt   []string
	clearMedia bool
	limit      int
	offset     int
}

func newPublicationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publication",
		Short: "Create and manage source publications",
		Long: "Create and manage source publications: the canonical idea, brief, or source material\n" +
			"that posts, threads, and assistant workflows can reference through --publication.",
	}
	cmd.AddCommand(newPublicationCreateCmd())
	cmd.AddCommand(newPublicationListCmd())
	cmd.AddCommand(newPublicationViewCmd())
	cmd.AddCommand(newPublicationUpdateCmd())
	return cmd
}

func newPublicationCreateCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a source publication",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, client, workspaceID, err := publicationRuntime(cmd)
			if err != nil {
				return err
			}
			title := strings.TrimSpace(flags.title)
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			sourceContent, err := publicationContent(flags.content, flags.file)
			if err != nil {
				return err
			}
			mediaIDs, err := publicationMediaIDs(cmd, client, workspaceID, flags)
			if err != nil {
				return err
			}
			publication, err := client.CreatePublication(cmd.Context(), api.CreatePublicationInput{
				WorkspaceID:   workspaceID,
				Title:         title,
				SourceContent: sourceContent,
				SourceURL:     strings.TrimSpace(flags.sourceURL),
				Goal:          strings.TrimSpace(flags.goal),
				Audience:      strings.TrimSpace(flags.audience),
				MediaIDs:      mediaIDs,
			})
			if err != nil {
				return err
			}
			return printPublicationSummary(cfg, publication)
		},
	}
	addPublicationWriteFlags(cmd, &flags, true)
	return cmd
}

func newPublicationListCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List source publications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, client, workspaceID, err := publicationRuntime(cmd)
			if err != nil {
				return err
			}
			publications, err := client.ListPublications(cmd.Context(), api.ListPublicationsInput{
				WorkspaceID: workspaceID,
				Status:      strings.TrimSpace(flags.status),
				Limit:       flags.limit,
				Offset:      flags.offset,
			})
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(publications)
			}
			if len(publications) == 0 {
				p.Printf("No source publications found for this workspace.")
				return nil
			}
			rows := make([][]string, 0, len(publications))
			for _, publication := range publications {
				rows = append(rows, []string{
					publication.ID,
					publication.Status,
					publication.Title,
					strconv.Itoa(len(publication.MediaIDs)),
					emptyDash(publication.UpdatedAt),
				})
			}
			p.Table([]string{"ID", "STATUS", "TITLE", "MEDIA", "UPDATED"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status: draft, ready, scheduled, published, failed")
	cmd.Flags().IntVar(&flags.limit, "limit", 0, "maximum number of publications to return")
	cmd.Flags().IntVar(&flags.offset, "offset", 0, "pagination offset")
	return cmd
}

func newPublicationViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <publication-id>",
		Short: "View a source publication",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			publication, err := client.GetPublication(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(publication)
			}
			p.Table([]string{"FIELD", "VALUE"}, [][]string{
				{"id", publication.ID},
				{"workspace_id", publication.WorkspaceID},
				{"created_by", publication.CreatedBy},
				{"status", publication.Status},
				{"title", publication.Title},
				{"source_url", emptyDash(publication.SourceURL)},
				{"goal", emptyDash(publication.Goal)},
				{"audience", emptyDash(publication.Audience)},
				{"media_ids", joinOrDash(publication.MediaIDs)},
				{"created_at", emptyDash(publication.CreatedAt)},
				{"updated_at", emptyDash(publication.UpdatedAt)},
				{"source_content", publication.SourceContent},
			})
			return nil
		},
	}
}

func newPublicationUpdateCmd() *cobra.Command {
	var flags publicationFlags
	cmd := &cobra.Command{
		Use:   "update <publication-id>",
		Short: "Update a source publication",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := publicationRuntime(cmd)
			if err != nil {
				return err
			}
			in := api.UpdatePublicationInput{}
			if cmd.Flags().Changed("title") {
				title := strings.TrimSpace(flags.title)
				if title == "" {
					return fmt.Errorf("--title cannot be empty")
				}
				in.Title = &title
			}
			if cmd.Flags().Changed("content") || cmd.Flags().Changed("file") {
				sourceContent, err := publicationContent(flags.content, flags.file)
				if err != nil {
					return err
				}
				in.SourceContent = &sourceContent
			}
			if cmd.Flags().Changed("source-url") {
				sourceURL := strings.TrimSpace(flags.sourceURL)
				in.SourceURL = &sourceURL
			}
			if cmd.Flags().Changed("goal") {
				goal := strings.TrimSpace(flags.goal)
				in.Goal = &goal
			}
			if cmd.Flags().Changed("audience") {
				audience := strings.TrimSpace(flags.audience)
				in.Audience = &audience
			}
			if cmd.Flags().Changed("status") {
				status := strings.TrimSpace(flags.status)
				if status == "" {
					return fmt.Errorf("--status cannot be empty")
				}
				in.Status = &status
			}
			if flags.clearMedia && cmd.Flags().Changed("media") {
				return fmt.Errorf("use either --media or --clear-media, not both")
			}
			if flags.clearMedia {
				mediaIDs := []string{}
				in.MediaIDs = &mediaIDs
			} else if cmd.Flags().Changed("media") {
				mediaIDs, err := publicationMediaIDs(cmd, client, workspaceID, flags)
				if err != nil {
					return err
				}
				in.MediaIDs = &mediaIDs
			}
			if !publicationUpdateChanged(cmd, flags.clearMedia) {
				return fmt.Errorf("no publication fields were changed")
			}
			publication, err := client.UpdatePublication(cmd.Context(), args[0], in)
			if err != nil {
				return err
			}
			return printPublicationSummary(cfg, publication)
		},
	}
	addPublicationWriteFlags(cmd, &flags, false)
	cmd.Flags().StringVar(&flags.status, "status", "", "status: draft, ready, scheduled, published, failed")
	cmd.Flags().BoolVar(&flags.clearMedia, "clear-media", false, "remove all source media attachments")
	return cmd
}

func addPublicationWriteFlags(cmd *cobra.Command, flags *publicationFlags, create bool) {
	cmd.Flags().StringVar(&flags.title, "title", "", "short internal title")
	cmd.Flags().StringVar(&flags.content, "content", "", "source idea, brief, announcement, notes, or canonical material")
	cmd.Flags().StringVar(&flags.file, "file", "", "read source content from a file")
	cmd.Flags().StringVar(&flags.sourceURL, "source-url", "", "source URL related to the publication")
	cmd.Flags().StringVar(&flags.goal, "goal", "", "goal such as announce, explain, launch, ask for feedback, or promote article")
	cmd.Flags().StringVar(&flags.audience, "audience", "", "intended audience")
	cmd.Flags().StringArrayVar(&flags.media, "media", nil, "media id or local file path to attach; repeatable")
	cmd.Flags().StringArrayVar(&flags.mediaAlt, "media-alt", nil, "alt text for the matching uploaded --media")
	if create {
		_ = cmd.MarkFlagRequired("title")
	}
}

func publicationRuntime(cmd *cobra.Command) (*config.Runtime, *api.Client, string, error) {
	cfg, err := runtimeFrom(cmd)
	if err != nil {
		return nil, nil, "", err
	}
	client, err := clientFrom(cfg)
	if err != nil {
		return nil, nil, "", err
	}
	workspaceID, err := activeWorkspaceID(cmd, client)
	if err != nil {
		return nil, nil, "", err
	}
	return cfg, client, workspaceID, nil
}

func publicationContent(content, file string) (string, error) {
	sourceContent, err := contentFromFlags(content, file)
	if err != nil {
		return "", err
	}
	sourceContent = strings.TrimSpace(sourceContent)
	if sourceContent == "" {
		return "", fmt.Errorf("--content or --file is required")
	}
	return sourceContent, nil
}

func publicationMediaIDs(cmd *cobra.Command, client *api.Client, workspaceID string, flags publicationFlags) ([]string, error) {
	if len(flags.media) == 0 {
		return nil, nil
	}
	return resolveMedia(cmd, client, workspaceID, flags.media, flags.mediaAlt)
}

func publicationUpdateChanged(cmd *cobra.Command, clearMedia bool) bool {
	for _, name := range []string{"title", "content", "file", "source-url", "goal", "audience", "status", "media"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return clearMedia
}

func printPublicationSummary(cfg *config.Runtime, publication *api.Publication) error {
	p := printerFrom(cfg)
	if cfg.AsJSON {
		return p.PrintJSON(publication)
	}
	p.Table([]string{"ID", "STATUS", "TITLE", "MEDIA", "UPDATED"}, [][]string{{
		publication.ID,
		publication.Status,
		publication.Title,
		strconv.Itoa(len(publication.MediaIDs)),
		emptyDash(publication.UpdatedAt),
	}})
	return nil
}

func joinOrDash(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ",")
}
