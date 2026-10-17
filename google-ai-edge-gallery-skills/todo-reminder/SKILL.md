---
name: todo-reminder
description: Schedule a reminder notification that opens the todo app when tapped. Use this when the user wants to be reminded to do something later, to add a task at a later time, or to review or triage their todo list.
metadata:
  homepage: https://github.com/alexhokl/todo-cli/tree/main/google-ai-edge-gallery-skills/todo-reminder
---

# Todo reminder

## Instructions

To schedule a todo reminder, you must follow these exact steps:

1. If the reminder does not repeat daily, call the `run_intent` tool with
   `intent` as `get_current_date_and_time` and `parameters` as `{}` to get the
   user's local date and time. Then explicitly calculate the date of the
   reminder, rolling over to the next month or year if needed.
2. Build the deep link that the notification opens:
   - To review or triage the todo list, use one of:
     - `todo://items?view=untriaged` for new items that need triaging.
     - `todo://items?view=time-sensitive` for items with a due date.
     - `todo://items?view=triaged` for the prioritised list.
   - Otherwise, to add a task, start with `todo://item/new` and add query
     parameters after `?`, joined with `&`:
     - `title`: a short imperative phrase for the task, e.g. `Pay rent`.
     - `description`: (optional) extra detail.
     - `labels`: (optional) label names the user mentioned, joined with `,`,
       e.g. `home,bills`.
   - Replace every space in a value with `%20` and every `&` with `%26`.
3. Call the `run_intent` tool with the following exact parameters:
   - intent: schedule_notification
   - parameters: A JSON string with the following fields:
     - title: the title of the notification, e.g. "Pay rent". String.
     - message: the message of the notification, e.g. "Tap to add it to your todo list". String.
     - hour: the hour of the day (0-23). Integer.
     - minute: the minute of the hour (0-59). Integer.
     - deeplink: the deep link from step 2. String.
     - year: (optional) the year of the reminder. Integer.
     - month: (optional) the month (1-12) of the reminder. Integer.
     - day: (optional) the day of the month (1-31) of the reminder. Integer.
     - repeat_daily: (optional) true if the reminder repeats every day. Boolean.
4. Tell the user when the reminder will appear and what tapping it opens.

Reminders can only repeat daily. If the user asks for a weekly or monthly
reminder, schedule the next occurrence only and tell the user.

### Examples

- "Remind me every day at 9am to triage my todos":

  ```json
  {"title": "Triage todos", "message": "Time to triage new items", "hour": 9, "minute": 0, "deeplink": "todo://items?view=untriaged", "repeat_daily": true}
  ```

- "Remind me on 1 November at 10am to pay rent, label it bills":

  ```json
  {"title": "Pay rent", "message": "Tap to add it to your todo list", "hour": 10, "minute": 0, "year": 2026, "month": 11, "day": 1, "deeplink": "todo://item/new?title=Pay%20rent&labels=bills"}
  ```

- "Remind me tomorrow at 6pm to check what's due":

  ```json
  {"title": "Check due items", "message": "Review your time-sensitive items", "hour": 18, "minute": 0, "year": 2026, "month": 10, "day": 18, "deeplink": "todo://items?view=time-sensitive"}
  ```
