package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alexhokl/todo-cli/proto"
)

// writeItemTable renders items as an aligned table. Active items are shown in
// their manual order, so the leading column is the priority in that order
// rather than an identifier the user has to interpret.
func writeItemTable(out io.Writer, items []*proto.Item, numbered bool) error {
	if len(items) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "#\tID\tTITLE\tLABELS\tLIST\tDUE"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	for i, item := range items {
		order := "-"
		if numbered {
			order = fmt.Sprintf("%d", i+1)
		}

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
			"%s\t%d\t%s\t%s\t%s\t%s\n",
			order, item.GetId(), item.GetTitle(), joinLabelNames(item), list, due,
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

// writeItemLine renders a single item as a one line confirmation.
func writeItemLine(out io.Writer, item *proto.Item) error {
	details := []string{fmt.Sprintf("id %d", item.GetId())}
	if len(item.GetLabels()) > 0 {
		details = append(details, fmt.Sprintf("labels %s", joinLabelNames(item)))
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