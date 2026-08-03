import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that edits an item's title and description.
///
/// The current values are passed in ([initialTitle], [initialDescription]) so
/// the fields render instantly without a [getItem] round-trip. On save the
/// page calls [ItemService.updateItem] and pops with `true` so the caller
/// (typically [ItemDetailPage]) can reload the canonical state.
///
/// When [service] is null the page builds one lazily from the persisted
/// backend configuration (the same seam used by [ItemDetailPage] and
/// [CommentsPage]). Tests inject a fake service so they never touch the
/// network or shared preferences.
class EditItemPage extends StatefulWidget {
  const EditItemPage({
    super.key,
    required this.itemId,
    required this.initialTitle,
    required this.initialDescription,
    this.service,
  });

  final int itemId;
  final String initialTitle;
  final String initialDescription;
  final ItemService? service;

  @override
  State<EditItemPage> createState() => _EditItemPageState();
}

class _EditItemPageState extends State<EditItemPage> {
  late final TextEditingController _titleController;
  late final TextEditingController _descriptionController;
  ItemService? _service;
  bool _ownsService = false;
  String? _titleError;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _titleController = TextEditingController(text: widget.initialTitle);
    _descriptionController =
        TextEditingController(text: widget.initialDescription);
    _service = widget.service;
    _ownsService = widget.service == null;
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    if (_ownsService) {
      _service?.dispose();
    }
    super.dispose();
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final title = _titleController.text.trim();
    if (title.isEmpty) {
      setState(() => _titleError = l10n.titleRequired);
      return;
    }
    final description = _descriptionController.text;

    setState(() {
      _saving = true;
      _titleError = null;
    });
    try {
      _service ??= await _buildService();
      await _service!.updateItem(
        id: widget.itemId,
        title: title,
        description: description,
      );
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } on ItemException catch (e) {
      _handleFailure(l10n, e.message);
    } catch (e) {
      _handleFailure(l10n, e.toString());
    }
  }

  void _handleFailure(AppLocalizations l10n, String message) {
    setState(() => _saving = false);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(l10n.failedToUpdateItem(message))),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.editItem),
        actions: [
          FilledButton(
            onPressed: _saving ? null : _submit,
            child: Text(l10n.save),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(
              controller: _titleController,
              decoration: InputDecoration(
                labelText: l10n.titleLabel,
                hintText: l10n.enterTitle,
                errorText: _titleError,
                border: const OutlineInputBorder(),
              ),
              textCapitalization: TextCapitalization.sentences,
              autofocus: true,
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _descriptionController,
              decoration: InputDecoration(
                labelText: l10n.descriptionLabel,
                hintText: l10n.enterDescription,
                border: const OutlineInputBorder(),
              ),
              textCapitalization: TextCapitalization.sentences,
              maxLines: 16,
            ),
          ],
        ),
      ),
    );
  }
}
