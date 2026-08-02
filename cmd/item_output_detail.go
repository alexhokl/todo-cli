package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alexhokl/todo-cli/proto"
)

// writeItemDetail renders a single item as a multi-line details view. The
// header line distinguishes done, untriaged, and active triaged items. Empty
// optional fields are omitted to keep the output compact for sparse items.
// Blocker and comment sub-tables are rendered only when the item carries any.
func writeItemDetail(out io.Writer, item *proto.Item) error {
	header := itemDetailHeader(item)
	if _, err := fmt.Fprintf(out, "%s\n\n", header); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	if desc := item.GetDescription(); desc != "" {
		if err := writeDetailRow(writer, "Description:", desc); err != nil {
			return err
		}
	}
	if item.GetDueDate() != nil {
		due := item.GetDueDate().AsTime().Local().Format(time.DateOnly)
		if err := writeDetailRow(writer, "Due:", due); err != nil {
			return err
		}
	}
	if item.ListId != nil {
		if err := writeDetailRow(writer, "List:", fmt.Sprintf("%d", item.GetListId())); err != nil {
			return err
		}
	}
	if !item.GetDone() {
		priority := "-"
		if item.Priority != nil {
			priority = formatPriority(*item.Priority)
		}
		if err := writeDetailRow(writer, "Priority:", priority); err != nil {
			return err
		}
	}
	if item.GetEffort() != nil {
		if err := writeDetailRow(writer, "Effort:", item.GetEffort().GetName()); err != nil {
			return err
		}
	}
	if len(item.GetLabels()) > 0 {
		if err := writeDetailRow(writer, "Labels:", joinLabelNames(item)); err != nil {
			return err
		}
	}
	if len(item.GetLinkedItems()) > 0 {
		if err := writeDetailRow(writer, "Linked:", joinLinkedItemSummaries(item)); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if err := writeBlockerDetail(out, item); err != nil {
		return err
	}
	if err := writeCommentDetail(out, item); err != nil {
		return err
	}

	return nil
}

// itemDetailHeader builds the leading line. A done item carries "(done)"; an
// active but untriaged item carries "(untriaged)"; an active triaged item
// carries no qualifier.
func itemDetailHeader(item *proto.Item) string {
	qualifier := ""
	switch {
	case item.GetDone():
		qualifier = " (done)"
	case item.Priority == nil:
		qualifier = " (untriaged)"
	}
	return fmt.Sprintf("Item %d%s: %s", item.GetId(), qualifier, item.GetTitle())
}

// writeDetailRow writes a single "label: value" line via the tabwriter so the
// colons align across all detail rows.
func writeDetailRow(w *tabwriter.Writer, label, value string) error {
	if _, err := fmt.Fprintf(w, "  %s\t%s\n", label, value); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}

// formatPriority renders the manual ordering rank. The stored value is a
// sparse fractional double; trim it to a compact form for display.
func formatPriority(p float64) string {
	trimmed := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", p), "0"), ".")
	if trimmed == "" || trimmed == "-" {
		return "0"
	}
	return trimmed
}

// joinLinkedItemSummaries renders the linked items as "<id> - <title>" pairs,
// comma-separated, mirroring the detail view's convention.
func joinLinkedItemSummaries(item *proto.Item) string {
	parts := make([]string, 0, len(item.GetLinkedItems()))
	for _, linked := range item.GetLinkedItems() {
		parts = append(parts, fmt.Sprintf("%d - %s", linked.GetId(), linked.GetTitle()))
	}
	return strings.Join(parts, ", ")
}

// writeBlockerDetail renders the blocker sub-table when the item carries any.
func writeBlockerDetail(out io.Writer, item *proto.Item) error {
	blockers := item.GetBlockers()
	if len(blockers) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(out, "  Blockers (%d):\n", len(blockers)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, blocker := range blockers {
		if _, err := fmt.Fprintf(writer, "    %d\t%s\n", blocker.GetId(), blocker.GetDescription()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// writeCommentDetail renders the comment sub-table when the item carries any.
func writeCommentDetail(out io.Writer, item *proto.Item) error {
	comments := item.GetComments()
	if len(comments) == 0 {
		return nil
	}

	if _, err := fmt.Fprintf(out, "  Comments (%d):\n", len(comments)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, comment := range comments {
		created := "-"
		if comment.GetCreatedAt() != nil {
			created = comment.GetCreatedAt().AsTime().Local().Format(time.DateTime)
		}
		author := comment.GetAuthor()
		if author == "" {
			author = "-"
		}
		if _, err := fmt.Fprintf(writer, "    %d\t%s\t%s\t%s\n", comment.GetId(), created, author, comment.GetBody()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}