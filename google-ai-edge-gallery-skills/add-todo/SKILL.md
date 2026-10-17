---
name: add-todo
description: Add an item to the todo app. Use this when the user wants to add, note or remember a task, todo or thing to do, shares a photo of something to do, or wants to open their todo list.
metadata:
  homepage: https://github.com/alexhokl/todo-cli/tree/main/google-ai-edge-gallery-skills/add-todo
---

# Add todo

## Instructions

The todo app opens a pre-filled create form from a link. Your job is to build
that link and show it to the user so they can tap it.

### Adding an item

1. Identify the following from the user's message or photo:
   - `title`: a short imperative phrase for the task, e.g. "Renew passport".
     Required.
   - `description`: extra detail, e.g. a deadline, address or note from the
     photo. Leave it out if there is none.
   - `labels`: label names, only if the user names them, e.g. "label it
     admin" or "tag it home and errand". Leave it out otherwise.
2. Call the `run_js` tool with the following exact parameters:
   - script name: index.html
   - data: A JSON string with the following fields:
     - title: String. The title. Required.
     - description: String. The description. Optional.
     - labels: Array of strings. The label names. Optional.
3. Reply with the returned link as a markdown link, exactly as returned, for
   example:

   [Open create form](todo://item/new?title=Renew%20passport&labels=admin)

   Briefly list the title, description and labels that will be filled in.
   Do not say the item has been added; the user saves it in the app.

### Opening the item list

When the user wants to see their items, do not call any tool. Reply with one
of these markdown links:

- [Open triaged items](todo://items?view=triaged)
- [Open untriaged items](todo://items?view=untriaged)
- [Open time-sensitive items](todo://items?view=time-sensitive)
- [Open done items](todo://items?view=done)

### Sample commands

- "Add a todo to renew my passport, label it admin"
- "Remind me to call the plumber about the leak under the sink"
- "Note down: buy milk and eggs, tag it home and errand"
- "Add this to my todo list" (with a photo)
- "Show my untriaged todos"
