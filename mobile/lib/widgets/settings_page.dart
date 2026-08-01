import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:todo/app_config.dart';
import 'package:todo/l10n/app_localizations.dart';

/// Parsed backend URL configuration.
class BackendConfig {
  final String host;
  final int port;
  final String scheme;

  const BackendConfig({
    required this.host,
    required this.port,
    this.scheme = 'http',
  });

  /// Parse a URL string into host and port. Defaults to port 443 for https,
  /// 80 for http when not specified.
  factory BackendConfig.fromUrl(String url) {
    final uri = Uri.parse(url);
    final host = uri.host;
    final scheme = uri.scheme.isEmpty ? 'http' : uri.scheme;
    int port;
    if (uri.hasPort) {
      port = uri.port;
    } else {
      port = scheme == 'https' ? 443 : 80;
    }
    return BackendConfig(host: host, port: port, scheme: scheme);
  }

  /// Load backend configuration from shared preferences.
  static Future<BackendConfig> load() async {
    final prefs = SharedPreferencesAsync();
    final url = await prefs.getString(SettingsPage.backendUrlKey) ??
        SettingsPage.defaultBackendUrl;
    return BackendConfig.fromUrl(url);
  }
}

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});

  static const String backendUrlKey = 'backend_url';
  static const String defaultBackendUrl = 'http://localhost:8080';

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  final TextEditingController _backendUrlController = TextEditingController();
  late final SharedPreferencesAsync _prefs;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _initPreferences();
  }

  Future<void> _initPreferences() async {
    _prefs = SharedPreferencesAsync();
    final savedUrl = await _prefs.getString(SettingsPage.backendUrlKey);
    setState(() {
      _backendUrlController.text =
          savedUrl ?? SettingsPage.defaultBackendUrl;
      _isLoading = false;
    });
  }

  Future<void> _saveBackendUrl(String url) async {
    await _prefs.setString(SettingsPage.backendUrlKey, url);
  }

  @override
  void dispose() {
    _backendUrlController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(AppLocalizations.of(context)!.settings),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              AppLocalizations.of(context)!.backendConfiguration,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _backendUrlController,
              decoration: InputDecoration(
                labelText: AppLocalizations.of(context)!.backendServiceUrl,
                hintText: AppLocalizations.of(context)!.enterBackendServiceUrl,
                border: const OutlineInputBorder(),
                prefixIcon: const Icon(Icons.cloud),
              ),
              keyboardType: TextInputType.url,
              onChanged: _saveBackendUrl,
            ),
            const SizedBox(height: 24),
            Text(
              AppLocalizations.of(context)!.about,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            TextField(
              controller: TextEditingController(text: AppConfig.gitCommit),
              readOnly: true,
              decoration: InputDecoration(
                labelText: AppLocalizations.of(context)!.gitCommit,
                border: const OutlineInputBorder(),
                prefixIcon: const Icon(Icons.commit),
                suffixIcon: IconButton(
                  icon: const Icon(Icons.copy),
                  onPressed: () {
                    Clipboard.setData(ClipboardData(text: AppConfig.gitCommit));
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(
                          AppLocalizations.of(context)!.gitCommitCopied,
                        ),
                        duration: const Duration(seconds: 2),
                      ),
                    );
                  },
                  tooltip: AppLocalizations.of(context)!.copyToClipboard,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}