import 'package:flutter/material.dart';

import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Full-page selector for choosing items to link to a source item.
///
/// The page lists every known item (active and completed) except the source
/// item itself (self-links are server-rejected) and items already linked to
/// it. The user toggles checkboxes for the items to link, then taps the save
/// action in the AppBar. On save the page calls
/// [ItemService.updateItemLinks] with the selected ids and pops with `true` so
/// the caller ([ItemDetailPage]) can reload the canonical state.
///
/// The layout mirrors [EditItemPage]: a [Scaffold] with an AppBar carrying a
/// [FilledButton] save action.
///
/// When [service] is null the page builds one lazily from the persisted backend
/// configuration (the same seam used by [EditItemPage]). Tests inject a fake
/// service so they never touch the network or shared preferences.
class SelectLinkedItemsPage extends StatefulWidget {
  const SelectLinkedItemsPage({
    super.key,
    required this.itemId,
    required this.alreadyLinked,
    this.service,
  });

  final int itemId;

  /// Items already linked to the source item. These are excluded from the
  /// candidate list so the user can only select items they can newly link.
  final List<Item> alreadyLinked;

  final ItemService? service;

  @override
  State<SelectLinkedItemsPage> createState() => _SelectLinkedItemsPageState();
}

class _SelectLinkedItemsPageState extends State<SelectLinkedItemsPage> {
  ItemService? _service;
  bool _ownsService = false;

  List<Item>? _candidates;
  String? _error;
  bool _isLoading = true;

  /// Ids the user has toggled. Starts empty.
  final Set<int> _selected = {};

  /// Current search query, trimmed. Empty means no filtering.
  String _query = '';
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocus = FocusNode();

  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _service = widget.service;
    _ownsService = widget.service == null;
    _load();
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocus.dispose();
    if (_ownsService) {
      _service?.dispose();
    }
    super.dispose();
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      _service ??= await _buildService();
      final result = await _service!.listItems();
      final excluded = <int>{widget.itemId};
      for (final linked in widget.alreadyLinked) {
        excluded.add(linked.id);
      }
      final all = [...result.active, ...result.completed];
      final candidates =
          all.where((item) => !excluded.contains(item.id)).toList();
      if (!mounted) return;
      setState(() {
        _candidates = candidates;
        _isLoading = false;
      });
    } on ItemException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Failed to load items: $e';
        _isLoading = false;
      });
    }
  }

  void _toggle(int id, bool? value) {
    setState(() {
      if (value == true) {
        _selected.add(id);
      } else {
        _selected.remove(id);
      }
    });
  }

  void _onSearchChanged(String value) {
    setState(() {
      _query = value.trim();
    });
  }

  void _clearSearch() {
    _searchController.clear();
    _searchFocus.unfocus();
    setState(() {
      _query = '';
    });
  }

  Future<void> _submit() async {
    if (_selected.isEmpty) {
      Navigator.of(context).pop(false);
      return;
    }
    final l10n = AppLocalizations.of(context)!;
    setState(() => _saving = true);
    try {
      _service ??= await _buildService();
      await _service!.updateItemLinks(
        id: widget.itemId,
        add: _selected.toList(),
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
      SnackBar(content: Text(l10n.failedToAddLinks(message))),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.selectLinkedItems),
        actions: [
          FilledButton(
            onPressed: _saving ? null : _submit,
            child: Text(l10n.save),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildSearchBar(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: TextField(
        controller: _searchController,
        focusNode: _searchFocus,
        onChanged: _onSearchChanged,
        decoration: InputDecoration(
          hintText: l10n.searchItems,
          prefixIcon: const Icon(Icons.search),
          suffixIcon: _query.isEmpty
              ? null
              : IconButton(
                  icon: const Icon(Icons.clear),
                  tooltip: l10n.clearSearch,
                  onPressed: _clearSearch,
                ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(24),
          ),
          contentPadding: const EdgeInsets.symmetric(horizontal: 12),
          isDense: true,
        ),
        textInputAction: TextInputAction.search,
      ),
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
    final candidates = _candidates ?? const <Item>[];
    if (candidates.isEmpty) {
      return Center(child: Text(l10n.noItems));
    }
    final filtered = _query.isEmpty
        ? candidates
        : candidates
            .where((i) => i.title.toLowerCase().contains(_query.toLowerCase()))
            .toList();
    return Column(
      children: [
        _buildSearchBar(context),
        Expanded(
          child: filtered.isEmpty
              ? Center(child: Text(l10n.noMatchingItems))
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: filtered.length,
                  itemBuilder: (context, index) {
                    final item = filtered[index];
                    return CheckboxListTile(
                      title: Text(item.title),
                      value: _selected.contains(item.id),
                      onChanged: (value) => _toggle(item.id, value),
                    );
                  },
                ),
        ),
      ],
    );
  }
}