import 'dart:async';

import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/effort_picker.dart';
import 'package:todo/widgets/label_picker.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that creates or edits an item's title and description.
///
/// When [itemId] is null the page runs in **create mode**: it calls
/// [ItemService.createItem] on submit and pops with `true` so the caller can
/// switch to the untriaged view. In create mode the page also lets the user
/// attach labels by picking from the catalogue returned by
/// [ItemService.listLabels]; selected labels accumulate locally and are
/// passed all at once to [ItemService.createItem] (the server creates any
/// unknown label names on the fly).
///
/// When [itemId] is non-null the page runs in **edit mode**: the current
/// values are passed in ([initialTitle], [initialDescription]) so the fields
/// render instantly without a [getItem] round-trip, and on save the page
/// calls [ItemService.updateItem] and pops with `true` so the caller
/// (typically [ItemDetailPage]) can reload the canonical state. Label editing
/// in edit mode is left to [ItemDetailPage], which has the full add/remove
/// flow.
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

  /// All known labels (for the add-label picker), loaded only in create
  /// mode alongside the rest of the page.
  List<Label>? _allLabels;
  String? _labelsError;

  /// Labels selected by the user in create mode, accumulated locally and
  /// passed to [ItemService.createItem] at submit time.
  final List<Label> _selectedLabels = [];

  /// All known efforts (for the edit-effort picker), loaded only in create
  /// mode alongside the rest of the page.
  List<Effort>? _allEfforts;
  String? _effortsError;

  /// Effort selected by the user in create mode. An empty string means "no
  /// effort" (the default). Passed to [ItemService.createItem] at submit
  /// time; an empty string leaves the item without an effort.
  String _selectedEffort = '';

  @override
  void initState() {
    super.initState();
    _titleController = TextEditingController(text: widget.initialTitle);
    _descriptionController =
        TextEditingController(text: widget.initialDescription);
    _service = widget.service;
    _ownsService = widget.service == null;
    if (widget.itemId == null) {
      unawaited(_loadLabels());
      unawaited(_loadEfforts());
    }
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

  /// Loads the catalogue of all known labels for the add-label picker. A
  /// failure is isolated so the rest of the page stays usable; the add-label
  /// button surfaces the error when tapped.
  Future<void> _loadLabels() async {
    setState(() {
      _labelsError = null;
    });
    try {
      _service ??= await _buildService();
      final labels = await _service!.listLabels();
      if (!mounted) return;
      setState(() {
        _allLabels = labels;
      });
    } on ItemException catch (e) {
      if (!mounted) return;
      setState(() {
        _labelsError = e.message;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _labelsError = 'Failed to load labels: $e';
      });
    }
  }

  /// Returns the current label catalogue, reloading it lazily when the
  /// initial fetch failed. Used by [showLabelPickerDialog] as its
  /// [ensureCatalogue] callback.
  Future<LabelCatalogue> _ensureLabelCatalogue() async {
    if (_allLabels == null || _labelsError != null) {
      await _loadLabels();
    }
    return LabelCatalogue(labels: _allLabels, error: _labelsError);
  }

  /// Loads the catalogue of all known efforts for the edit-effort picker. A
  /// failure is isolated so the rest of the page stays usable; the
  /// edit-effort button surfaces the error when tapped.
  Future<void> _loadEfforts() async {
    setState(() {
      _effortsError = null;
    });
    try {
      _service ??= await _buildService();
      final efforts = await _service!.listEfforts();
      if (!mounted) return;
      setState(() {
        _allEfforts = efforts;
      });
    } on ItemException catch (e) {
      if (!mounted) return;
      setState(() {
        _effortsError = e.message;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _effortsError = 'Failed to load efforts: $e';
      });
    }
  }

  /// Returns the current effort catalogue, reloading it lazily when the
  /// initial fetch failed. Used by [showEffortPickerDialog] as its
  /// [ensureCatalogue] callback.
  Future<EffortCatalogue> _ensureEffortCatalogue() async {
    if (_allEfforts == null || _effortsError != null) {
      await _loadEfforts();
    }
    return EffortCatalogue(efforts: _allEfforts, error: _effortsError);
  }

  /// Opens the shared edit-effort picker and stores the selection (which may
  /// be an empty string to mean "No effort") in [_selectedEffort].
  Future<void> _showEffortDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final result = await showEffortPickerDialog(
      context,
      ensureCatalogue: _ensureEffortCatalogue,
    );
    if (result.aborted) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToSetEffort(result.error!))),
      );
      return;
    }
    if (result.dismissed || !mounted) return;
    setState(() => _selectedEffort = result.name!);
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final title = _titleController.text.trim();
    if (title.isEmpty) {
      setState(() => _titleError = l10n.titleRequired);
      return;
    }
    final description = _descriptionController.text;
    final labelNames =
        _selectedLabels.map((l) => l.name).toList(growable: false);

    setState(() {
      _saving = true;
      _titleError = null;
    });
    try {
      _service ??= await _buildService();
      if (widget.itemId == null) {
        await _service!.createItem(
          title: title,
          description: description,
          labels: labelNames.isEmpty ? null : labelNames,
          effort: _selectedEffort,
        );
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

  /// Opens the shared add-label picker and appends the selection (if any) to
  /// [_selectedLabels]. Already-selected labels are excluded.
  Future<void> _showAddLabelDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final result = await showLabelPickerDialog(
      context,
      ensureCatalogue: _ensureLabelCatalogue,
      excludedNames:
          _selectedLabels.map((l) => l.name).toSet(),
    );
    if (result.aborted) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.failedToAddLabel(result.error!))),
      );
      return;
    }
    if (result.label == null || !mounted) return;
    setState(() => _selectedLabels.add(result.label!));
  }

  Widget _sectionLabel(BuildContext context, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Text(
        text,
        style: Theme.of(context).textTheme.titleSmall?.copyWith(
              color: Theme.of(context).colorScheme.primary,
            ),
      ),
    );
  }

  Widget _mutedHint(BuildContext context, String text) {
    final theme = Theme.of(context);
    return Text(
      text,
      style: theme.textTheme.bodyMedium?.copyWith(
        color: theme.colorScheme.onSurfaceVariant.withValues(alpha: 0.6),
        fontStyle: FontStyle.italic,
      ),
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
        child: SingleChildScrollView(
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
              if (isCreate) ...[
                const SizedBox(height: 16),
                _sectionLabel(context, l10n.effort),
                if (_selectedEffort.isEmpty)
                  _mutedHint(context, l10n.noEffort)
                else
                  Text(_selectedEffort),
                Align(
                  alignment: Alignment.centerLeft,
                  child: TextButton.icon(
                    onPressed: _showEffortDialog,
                    icon: const Icon(Icons.edit),
                    label: Text(l10n.editEffort),
                  ),
                ),
                const SizedBox(height: 16),
                _sectionLabel(context, l10n.labels),
                if (_selectedLabels.isEmpty)
                  _mutedHint(context, l10n.noLabels)
                else
                  Wrap(
                    spacing: 8,
                    runSpacing: 4,
                    children: [
                      for (final label in _selectedLabels)
                        labelChip(
                          context,
                          label,
                          onDeleted: () => setState(
                            () => _selectedLabels.remove(label),
                          ),
                        ),
                    ],
                  ),
                Align(
                  alignment: Alignment.centerLeft,
                  child: TextButton.icon(
                    onPressed: _showAddLabelDialog,
                    icon: const Icon(Icons.add),
                    label: Text(l10n.addLabel),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}