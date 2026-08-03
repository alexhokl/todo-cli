import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that creates or edits an item's title and description.
///
/// When [itemId] is null the page runs in **create mode**: it calls
/// [ItemService.createItem] on submit and pops with `true` so the caller can
/// switch to the untriaged view. When [itemId] is non-null the page runs in
/// **edit mode**: the current values are passed in ([initialTitle],
/// [initialDescription]) so the fields render instantly without a [getItem]
/// round-trip, and on save the page calls [ItemService.updateItem] and pops
/// with `true` so the caller (typically [ItemDetailPage]) can reload the
/// canonical state.
///
/// When [service] is null the page builds one lazily from the persisted
/// backend configuration (the same seam used by [ItemDetailPage] and
/// [CommentsPage]). Tests inject a fake service so they never touch the
/// network or shared preferences.
class EditItemPage extends StatefulWidget {
  const EditItemPage({
    super.key,
    this.itemId,
    this.initialTitle = '',
    this.initialDescription = '',
    this.service,
  });

  final int? itemId;
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
      if (widget.itemId == null) {
        await _service!.createItem(title: title, description: description);
      } else {
        await _service!.updateItem(
          id: widget.itemId!,
          title: title,
          description: description,
        );
      }
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
    final text = widget.itemId == null
        ? l10n.failedToCreateItem(message)
        : l10n.failedToUpdateItem(message);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(text)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isCreate = widget.itemId == null;
    return Scaffold(
      appBar: AppBar(
        title: Text(isCreate ? l10n.createItem : l10n.editItem),
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