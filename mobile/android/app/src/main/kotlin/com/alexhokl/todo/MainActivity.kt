package com.alexhokl.todo

import android.content.Intent
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private var channel: MethodChannel? = null
    private var initialIntent: Map<String, String?>? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        initialIntent = toMap(intent)
        channel = MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL).apply {
            setMethodCallHandler { call, result ->
                when (call.method) {
                    "getInitialIntent" -> {
                        result.success(initialIntent)
                        // Consume it so a hot restart or re-read does not replay it
                        initialIntent = null
                    }
                    else -> result.notImplemented()
                }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        toMap(intent)?.let { channel?.invokeMethod("onIntent", it) }
    }

    private fun toMap(intent: Intent?): Map<String, String?>? {
        if (intent == null) return null
        return when (intent.action) {
            Intent.ACTION_PROCESS_TEXT -> mapOf(
                "action" to "process_text",
                "text" to intent.getCharSequenceExtra(Intent.EXTRA_PROCESS_TEXT)?.toString(),
            )
            // Browsers share a page as subject = page title, text = URL
            Intent.ACTION_SEND -> mapOf(
                "action" to "send",
                "text" to intent.getStringExtra(Intent.EXTRA_TEXT),
                "subject" to intent.getStringExtra(Intent.EXTRA_SUBJECT),
            )
            Intent.ACTION_VIEW -> mapOf(
                "action" to "view",
                "uri" to intent.dataString,
            )
            else -> null
        }
    }

    companion object {
        private const val CHANNEL = "com.alexhokl.todo/intents"
    }
}
