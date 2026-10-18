import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter_helper/config.dart';
import 'package:flutter_helper/ui.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/services/item_service.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});

  static const String backendUrlKey = BackendConfig.storageKey;
  static const String defaultBackendUrl = ItemService.defaultBackendUrl;

  /// Storage key and default for the "show labels in item list" toggle.
  static const String showLabelsKey = 'show_labels_in_item_list';
  static const bool defaultShowLabels = true;

  /// Loads the persisted "show labels in item list" setting, defaulting to
  /// [defaultShowLabels] when unset. Mirrors [BackendConfig.load] so consumers
  /// read the same way as the backend URL.
  static Future<bool> loadShowLabels() async {
    final prefs = SharedPreferencesAsync();
    return await prefs.getBool(showLabelsKey) ?? defaultShowLabels;
  }

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  final TextEditingController _backendUrlController = TextEditingController();
  late final SharedPreferencesAsync _prefs;
  bool _isLoading = true;
  bool _showLabels = SettingsPage.defaultShowLabels;

  @override
  void initState() {
    super.initState();
    _initPreferences();
  }

  Future<void> _initPreferences() async {
    _prefs = SharedPreferencesAsync();
    final savedUrl = await _prefs.getString(SettingsPage.backendUrlKey);
    final savedShowLabels = await _prefs.getBool(SettingsPage.showLabelsKey);
    setState(() {
      _backendUrlController.text =
          savedUrl ?? SettingsPage.defaultBackendUrl;
      _showLabels = savedShowLabels ?? SettingsPage.defaultShowLabels;
      _isLoading = false;
    });
  }

  Future<void> _saveBackendUrl(String url) async {
    await BackendConfig.save(url);
  }

  Future<void> _saveShowLabels(bool value) async {
    await _prefs.setBool(SettingsPage.showLabelsKey, value);
    setState(() => _showLabels = value);
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
              AppLocalizations.of(context)!.display,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            SizedBox(
              width: double.infinity,
              child: SwitchListTile(
                title: Text(AppLocalizations.of(context)!.showLabelsInItemList),
                subtitle: Text(
                  AppLocalizations.of(context)!.showLabelsInItemListDescription,
                ),
                value: _showLabels,
                onChanged: _saveShowLabels,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              AppLocalizations.of(context)!.about,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            GitCommitField(
              label: AppLocalizations.of(context)!.gitCommit,
              copiedMessage: AppLocalizations.of(context)!.gitCommitCopied,
              copyTooltip: AppLocalizations.of(context)!.copyToClipboard,
            ),
          ],
        ),
      ),
    );
  }
}