package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alexhokl/todo-cli/proto"
)

// writeTodoTable renders todos as an aligned table. Active todos are shown in
// their manual order, so the leading column is the position in that order
// rather than an identifier the user has to interpret.
func writeTodoTable(out io.Writer, todos []*proto.Todo, numbered bool) error {
	if len(todos) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	writer := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(writer, "#\tID\tTITLE\tLABELS\tLIST\tDUE"); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	for i, todo := range todos {
		order := "-"
		if numbered {
			order = fmt.Sprintf("%d", i+1)
		}

		list := "-"
		if todo.ListId != nil {
			list = fmt.Sprintf("%d", todo.GetListId())
		}

		due := "-"
		if todo.DueDate != nil {
			due = todo.GetDueDate().AsTime().Local().Format(time.DateOnly)
		}

		if _, err := fmt.Fprintf(
			writer,
			"%s\t%d\t%s\t%s\t%s\t%s\n",
			order, todo.GetId(), todo.GetTitle(), joinLabelNames(todo), list, due,
		); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// joinLabelNames renders the labels of a todo as a single comma separated
// cell, or a dash when the todo carries none.
func joinLabelNames(todo *proto.Todo) string {
	if len(todo.GetLabels()) == 0 {
		return "-"
	}

	names := make([]string, 0, len(todo.GetLabels()))
	for _, label := range todo.GetLabels() {
		names = append(names, label.GetName())
	}

	return strings.Join(names, ",")
}

// writeTodoLine renders a single todo as a one line confirmation.
func writeTodoLine(out io.Writer, todo *proto.Todo) error {
	details := []string{fmt.Sprintf("id %d", todo.GetId())}
	if len(todo.GetLabels()) > 0 {
		details = append(details, fmt.Sprintf("labels %s", joinLabelNames(todo)))
	}
	if todo.ListId != nil {
		details = append(details, fmt.Sprintf("list %d", todo.GetListId()))
	}
	if todo.GetDone() {
		details = append(details, "done")
	}

	if _, err := fmt.Fprintf(out, "%s (%s)\n", todo.GetTitle(), strings.Join(details, ", ")); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
