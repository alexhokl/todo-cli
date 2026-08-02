package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alexhokl/todo-cli/proto"
)

// writeItemTable renders items as an aligned table. Active items are returned
// in their manual order, so the row order (not an explicit ordinal column)
// reflects the priority ranking.
func writeItemTable(out io.Writer, items []*proto.Item) error {
	if len(items) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "ID\tTITLE\tLABELS\tEFFORT\tBLOCKERS\tCOMMENTS\tLINKED\tLIST\tDUE"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	for _, item := range items {
		list := "-"
		if item.ListId != nil {
			list = fmt.Sprintf("%d", item.GetListId())
		}

		due := "-"
		if item.DueDate != nil {
			due = item.GetDueDate().AsTime().Local().Format(time.DateOnly)
		}

		if _, err := fmt.Fprintf(
			writer,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.GetId(), item.GetTitle(), joinLabelNames(item), effortName(item), joinBlockerDescriptions(item), commentCount(item), joinLinkedItemIDs(item), list, due,
		); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// joinLabelNames renders the labels of an item as a single comma separated
// cell, or a dash when the item carries none.
func joinLabelNames(item *proto.Item) string {
	if len(item.GetLabels()) == 0 {
		return "-"
	}

	names := make([]string, 0, len(item.GetLabels()))
	for _, label := range item.GetLabels() {
		names = append(names, label.GetName())
	}

	return strings.Join(names, ",")
}

// effortName renders the item's effort name, or a dash when it carries none.
func effortName(item *proto.Item) string {
	if item.GetEffort() == nil {
		return "-"
	}
	return item.GetEffort().GetName()
}

// joinBlockerDescriptions renders the item's blockers as a single
// semicolon-separated cell, or a dash when the item carries none.
func joinBlockerDescriptions(item *proto.Item) string {
	if len(item.GetBlockers()) == 0 {
		return "-"
	}
	descriptions := make([]string, 0, len(item.GetBlockers()))
	for _, blocker := range item.GetBlockers() {
		descriptions = append(descriptions, blocker.GetDescription())
	}
	return strings.Join(descriptions, "; ")
}

// commentCount renders the number of comments on the item, or a dash when it
// carries none. Bodies are not inlined because they are free-form and may be
// long; the per-item table is kept narrow on purpose.
func commentCount(item *proto.Item) string {
	if len(item.GetComments()) == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", len(item.GetComments()))
}

// joinLinkedItemIDs renders the ids of the item's linked items as a single
// comma-separated cell, or a dash when the item carries none.
func joinLinkedItemIDs(item *proto.Item) string {
	if len(item.GetLinkedItems()) == 0 {
		return "-"
	}
	ids := make([]string, 0, len(item.GetLinkedItems()))
	for _, linked := range item.GetLinkedItems() {
		ids = append(ids, fmt.Sprintf("%d", linked.GetId()))
	}
	return strings.Join(ids, ",")
}

// writeItemLine renders a single item as a one line confirmation.
func writeItemLine(out io.Writer, item *proto.Item) error {
	details := []string{fmt.Sprintf("id %d", item.GetId())}
	if len(item.GetLabels()) > 0 {
		details = append(details, fmt.Sprintf("labels %s", joinLabelNames(item)))
	}
	if item.GetEffort() != nil {
		details = append(details, fmt.Sprintf("effort %s", item.GetEffort().GetName()))
	}
	if len(item.GetBlockers()) > 0 {
		details = append(details, fmt.Sprintf("blockers %d", len(item.GetBlockers())))
	}
	if len(item.GetComments()) > 0 {
		details = append(details, fmt.Sprintf("comments %d", len(item.GetComments())))
	}
	if len(item.GetLinkedItems()) > 0 {
		details = append(details, fmt.Sprintf("linked %s", joinLinkedItemIDs(item)))
	}
	if item.ListId != nil {
		details = append(details, fmt.Sprintf("list %d", item.GetListId()))
	}
	if item.GetDone() {
		details = append(details, "done")
	}

	if _, err := fmt.Fprintf(out, "%s (%s)\n", item.GetTitle(), strings.Join(details, ", ")); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}