import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Page that lists every effort and lets the user create, rename, and delete
/// efforts. An effort is a per-user named level of effort (e.g. "low",
/// "medium", "high"); unlike a label it has no colour and an item carries at
/// most one effort via a belongs-to foreign key.
///
/// When [service] is null the page builds one lazily from the persisted backend
/// configuration (the same seam used by [ItemList]). Tests inject a fake
/// service so they never touch the network or shared preferences.
class EffortsPage extends StatefulWidget {
  const EffortsPage({super.key, this.service});

  final ItemService? service;

  @override
  State<EffortsPage> createState() => _EffortsPageState();
}

class _EffortsPageState extends State<EffortsPage> {
  ItemService? _service;
  List<Effort>? _efforts;
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
      final efforts = await _service!.listEfforts();
      setState(() {
        _efforts = efforts;
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
      builder: (_) => _EffortDialog(
        title: l10n.addEffort,
        service: _service!,
      ),
    );
    if (created == true && mounted) {
      await _load();
      _showSnackbar(l10n.effortCreated);
    }
  }

  Future<void> _showEditDialog(Effort effort) async {
    final l10n = AppLocalizations.of(context)!;
    final updated = await showDialog<bool>(
      context: context,
      builder: (_) => _EffortDialog(
        title: l10n.editEffort,
        service: _service!,
        effortId: effort.id,
        initialName: effort.name,
      ),
    );
    if (updated == true && mounted) {
      await _load();
      _showSnackbar(l10n.effortRenamed);
    }
  }

  Future<void> _confirmDelete(Effort effort) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(l10n.delete),
        content: Text(l10n.confirmDeleteEffort(effort.name)),
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
      await _service!.deleteEffort(effort.id);
      if (!mounted) return;
      await _load();
      _showSnackbar(l10n.effortDeleted);
    } on ItemException catch (e) {
      _showSnackbar(l10n.failedToDeleteEffort(e.message));
    } catch (e) {
      _showSnackbar(l10n.failedToDeleteEffort(e.toString()));
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
      appBar: AppBar(title: Text(l10n.efforts)),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreateDialog,
        tooltip: l10n.addEffort,
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
    final efforts = _efforts ?? const <Effort>[];
    if (efforts.isEmpty) {
      return Center(child: Text(l10n.noEfforts));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        itemCount: efforts.length,
        itemBuilder: (context, index) {
          final effort = efforts[index];
          return ListTile(
            leading: const Icon(Icons.bolt),
            title: Text(effort.name),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.edit),
                  tooltip: l10n.editEffort,
                  onPressed: () => _showEditDialog(effort),
                ),
                IconButton(
                  icon: const Icon(Icons.delete),
                  tooltip: l10n.delete,
                  onPressed: () => _confirmDelete(effort),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

/// Dialog used for both creating and editing an effort. Validates the name is
/// non-empty. On confirm it calls the matching [ItemService] method and pops
/// with `true` when the operation succeeds. Errors are surfaced as a SnackBar
/// and the dialog stays open so the user can retry.
class _EffortDialog extends StatefulWidget {
  const _EffortDialog({
    required this.title,
    required this.service,
    this.effortId,
    this.initialName,
  });

  final String title;
  final ItemService service;
  final int? effortId;
  final String? initialName;

  @override
  State<_EffortDialog> createState() => _EffortDialogState();
}

class _EffortDialogState extends State<_EffortDialog> {
  late final TextEditingController _nameController;
  String? _nameError;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.initialName ?? '');
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  bool get _isCreate => widget.effortId == null;

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    final name = _nameController.text.trim();

    String? nameError;
    if (name.isEmpty) {
      nameError = l10n.enterEffortName;
    }
    if (nameError != null) {
      setState(() {
        _nameError = nameError;
      });
      return;
    }

    setState(() {
      _saving = true;
      _nameError = null;
    });

    try {
      if (_isCreate) {
        await widget.service.createEffort(name);
      } else {
        await widget.service.renameEffort(
          widget.effortId!,
          name: name,
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
              ? l10n.failedToCreateEffort(message)
              : l10n.failedToRenameEffort(message),
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
              labelText: l10n.effortName,
              hintText: l10n.enterEffortName,
              errorText: _nameError,
              border: const OutlineInputBorder(),
            ),
            textCapitalization: TextCapitalization.words,
            autofocus: true,
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
          child: Text(_isCreate ? l10n.addEffort : l10n.editEffort),
        ),
      ],
    );
  }
}