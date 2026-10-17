# Google AI Edge Gallery skills

[Agent Skills](https://github.com/google-ai-edge/gallery/blob/main/skills/README.md)
for the [Google AI Edge Gallery](https://github.com/google-ai-edge/gallery)
app. They let an on-device model hand items to the todo Android app through
its `todo://` deep links (see [Android Intents](../README.md#android-intents)).

| Skill | Type | What it does |
|-------|------|--------------|
| [add-todo](add-todo) | JS | Turns "remind me to renew my passport, label it admin" or a photo of something to do into a link that opens the pre-filled create form. Also links to the item list views. |
| [todo-reminder](todo-reminder) | Native | Schedules a one-off or daily notification that opens the create form pre-filled, or an item list view (e.g. a daily triage reminder). |

Gallery cannot open other apps on its own: custom native intents need changes
to Gallery itself, and its webviews do not follow custom URL schemes. The JS
skill cannot reach the gRPC server either. Instead, `add-todo` replies with a
markdown link that you tap, and `todo-reminder` uses the built-in
`schedule_notification` intent, whose `deeplink` is opened when the
notification is tapped. Items are saved in the todo app, not by the skill.

## Installing

Push a skill folder to the device and import it in Gallery:

```sh
adb push google-ai-edge-gallery-skills/add-todo /sdcard/Download/
adb push google-ai-edge-gallery-skills/todo-reminder /sdcard/Download/
```

In Gallery, open **Agent Skills** with a model, tap the **Skills** chip, tap
**(+)**, choose **Import local skill** and select the skill folder.

To load a skill from a URL instead, host this directory on a web host such as
GitHub Pages (with an empty `.nojekyll` file at the root). Raw GitHub URLs
serve files as `text/plain`, which breaks JS skills.

## Testing

The link built by `add-todo` can be checked in a desktop browser by opening
`add-todo/scripts/index.html` and running in the console:

```js
await window.ai_edge_gallery_get_result(JSON.stringify({
  title: "Renew passport", description: "Expires in March", labels: "admin, travel",
}))
// {"result":"todo://item/new?title=Renew%20passport&description=Expires%20in%20March&labels=admin%2Ctravel"}
```

Then open the link on a device with the app installed:

```sh
# quote the whole command so the device shell does not treat & as a separator
adb shell "am start -a android.intent.action.VIEW \
  -d 'todo://item/new?title=Renew%20passport&description=Expires%20in%20March&labels=admin%2Ctravel'"
```

In Gallery, expand the skill execution panel to see the data passed to the
script and its console logs.
