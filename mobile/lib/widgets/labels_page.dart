import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that lists every label and lets the user create, rename, and delete
/// labels. Each label is shown with a colour swatch derived from its
/// `colour` field (a `#RRGGBB` hex string).
///
/// When [service] is null the page builds one lazily from the persisted backend
/// configuration (the same seam used by [ItemList]). Tests inject a fake
/// service so they never touch the network or shared preferences.
class LabelsPage extends StatefulWidget {
  const LabelsPage({super.key, this.service});

  final ItemService? service;

  @override
  State<LabelsPage> createState() => _LabelsPageState();
}

class _LabelsPageState extends State<LabelsPage> {
  ItemService? _service;
  List<Label>? _labels;
  String? _error;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _service = widget.service;
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      _service ??= await _buildService();
      final labels = await _service!.listLabels();
      setState(() {
        _labels = labels;
        _isLoading = false;
      });
    } on ItemException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> _showCreateDialog() async {
    final l10n = AppLocalizations.of(context)!;
    final created = await showDialog<bool>(
      context: context,
      builder: (_) => _LabelDialog(
        title: l10n.addLabel,
        service: _service!,
        defaultColour: '#FFFF00',
      ),
    );
    if (created == true && mounted) {
      await _load();
      _showSnackbar(l10n.labelCreated);
    }
  }

  Future<void> _showEditDialog(Label label) async {
    final l10n = AppLocalizations.of(context)!;
    final updated = await showDialog<bool>(
      context: context,
      builder: (_) => _LabelDialog(
        title: l10n.editLabel,
        service: _service!,
        labelId: label.id,
        initialName: label.name,
        initialColour: label.hasColour() ? label.colour : null,
        defaultColour: '#FFFF00',
      ),
    );
    if (updated == true && mounted) {
      await _load();
      _showSnackbar(l10n.labelRenamed);
    }
  }

  Future<void> _confirmDelete(Label label) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(l10n.delete),
        content: Text(l10n.confirmDeleteLabel(label.name)),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: Text(l10n.delete),
          ),
        ],
      ),
    );
    if (confirmed != true) {
      return;
    }
    try {
      await _service!.deleteLabel(label.id);
      if (!mounted) return;
      await _load();
      _showSnackbar(l10n.labelDeleted);
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToDeleteLabel(e.message));
    } catch (e) {
      _showSnackbar(l10n.failedToDeleteLabel(e.toString()));
    }
  }

  void _showSnackbar(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), duration: const Duration(seconds: 2)),
    );
  }

  @override
  void dispose() {
    if (widget.service == null) {
      _service?.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.labels)),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateDialog,
        tooltip: l10n.addLabel,
        child: const Icon(Icons.add),
      ),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildBody(BuildContext context, AppLocalizations l10n) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error!, style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: _load,
              icon: const Icon(Icons.refresh),
              label: Text(l10n.retry),
            ),
          ],
        ),
      );
    }
    final labels = _labels ?? const <Label>[];
    if (labels.isEmpty) {
      return Center(child: Text(l10n.noLabels));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        itemCount: labels.length,
        itemBuilder: (context, index) {
          final label = labels[index];
          return ListTile(
            leading: _ColourSwatch(colour: label.colour),
            title: Text(label.name),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.edit),
                  tooltip: l10n.editLabel,
                  onPressed: () => _showEditDialog(label),
                ),
                IconButton(
                  icon: const Icon(Icons.delete),
                  tooltip: l10n.delete,
                  onPressed: () => _confirmDelete(label),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

/// A circular swatch painted from a `#RRGGBB` hex string. Falls back to a
/// neutral grey when the colour cannot be parsed.
class _ColourSwatch extends StatelessWidget {
  const _ColourSwatch({required this.colour});

  final String colour;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 24,
      height: 24,
      decoration: BoxDecoration(
        color: _parseColour(colour),
        shape: BoxShape.circle,
        border: Border.all(color: Theme.of(context).dividerColor, width: 1),
      ),
    );
  }

  static Color _parseColour(String hex) {
    if (hex.length == 7 && hex.startsWith('#')) {
      final value = int.tryParse(hex.substring(1), radix: 16);
      if (value != null) {
        return Color(0xFF000000 | value);
      }
    }
    return Colors.grey;
  }
}

/// Dialog used for both creating and editing a label. Validates the name is
/// non-empty and that the colour, when entered, is in canonical #RRGGBB form.
/// On confirm it calls the matching [ItemService] method and pops with `true`
/// when the operation succeeds. Errors are surfaced as a SnackBar and the
/// dialog stays open so the user can retry.
class _LabelDialog extends StatefulWidget {
  const _LabelDialog({
    required this.title,
    required this.service,
    required this.defaultColour,
    this.labelId,
    this.initialName,
    this.initialColour,
  });

  final String title;
  final ItemService service;
  final String defaultColour;
  final int? labelId;
  final String? initialName;
  final String? initialColour;

  @override
  State<_LabelDialog> createState() => _LabelDialogState();
}

class _LabelDialogState extends State<_LabelDialog> {
  late final TextEditingController _nameController;
  late final TextEditingController _colourController;
  String? _nameError;
  String? _colourError;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.initialName ?? '');
    _colourController =
        TextEditingController(text: widget.initialColour ?? widget.defaultColour);
  }

  @override
  void dispose() {
    _nameController.dispose();
    _colourController.dispose();
    super.dispose();
  }

  bool get _isCreate => widget.labelId == null;

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final name = _nameController.text.trim();
    final colour = _colourController.text.trim();

    String? nameError;
    String? colourError;
    if (name.isEmpty) {
      nameError = l10n.enterLabelName;
    }
    try {
      ItemService.normaliseColour(colour);
    } on ItemException {
      colourError = l10n.invalidColour;
    }
    if (nameError != null || colourError != null) {
      setState(() {
        _nameError = nameError;
        _colourError = colourError;
      });
      return;
    }

    setState(() {
      _saving = true;
      _nameError = null;
      _colourError = null;
    });

    try {
      if (_isCreate) {
        await widget.service.createLabel(name, colour: colour);
      } else {
        await widget.service.renameLabel(
          widget.labelId!,
          name: name,
          colour: colour,
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
    setState(() {
      _saving = false;
    });
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          _isCreate
              ? l10n.failedToCreateLabel(message)
              : l10n.failedToRenameLabel(message),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return AlertDialog(
      title: Text(widget.title),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _nameController,
            decoration: InputDecoration(
              labelText: l10n.labelName,
              hintText: l10n.enterLabelName,
              errorText: _nameError,
              border: const OutlineInputBorder(),
            ),
            textCapitalization: TextCapitalization.words,
            autofocus: true,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _colourController,
            decoration: InputDecoration(
              labelText: l10n.labelColour,
              hintText: l10n.enterLabelColour,
              errorText: _colourError,
              border: const OutlineInputBorder(),
              prefixIcon: ValueListenableBuilder<TextEditingValue>(
                valueListenable: _colourController,
                builder: (context, value, _) =>
                    Icon(Icons.circle, color: _ColourSwatch._parseColour(value.text)),
              ),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: _saving ? null : () => Navigator.of(context).pop(false),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: _saving ? null : _submit,
          child: Text(_isCreate ? l10n.addLabel : l10n.editLabel),
        ),
      ],
    );
  }
}