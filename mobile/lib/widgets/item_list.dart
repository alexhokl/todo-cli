import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/colour.dart';
import 'package:todo/widgets/item_detail_page.dart';
import 'package:todo/widgets/settings_page.dart';
import 'package:todo/widgets/status_icon.dart';

/// Read-only list of todo items fetched from the gRPC server.
///
/// A chip bar at the top selects the bucket shown (triaged by default,
/// plus untriaged, time-sensitive, and completed). When [service] is null
/// the page builds one lazily from the persisted backend configuration
/// (the same seam used by [LabelsPage] and [CommentsPage]). Tests inject a
/// fake service so they never touch the network or shared preferences.
class ItemList extends StatefulWidget {
  const ItemList({super.key, this.service});

  final ItemService? service;

  @override
  State<ItemList> createState() => ItemListState();
}

class ItemListState extends State<ItemList> {
  /// Currently selected bucket. Defaults to triaged active items.
  ItemView _view = ItemView.ITEM_VIEW_TRIAGED;

  /// Whether the chip bar is expanded to reveal all four bucket chips.
  bool _chipsExpanded = false;

  List<Item>? _items;
  String? _error;
  bool _isLoading = true;
  ItemService? _service;

  /// All known labels (for the label filter bar), loaded alongside the items.
  List<Label> _allLabels = const [];

  /// Currently selected label names (AND semantics, matching the CLI
  /// `--label` repeatable flag). Persist across bucket switches.
  final Set<String> _selectedLabels = {};

  /// Current search query, trimmed. Empty means no filtering.
  String _query = '';
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocus = FocusNode();

  @override
  void initState() {
    super.initState();
    _service = widget.service;
    _load();
    _loadLabels();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      _service ??= await _buildService();
      final result = await _service!.listItems(
        view: _view,
        labels: _selectedLabels.isEmpty ? null : _selectedLabels.toList(),
      );
      setState(() {
        _items = result.active;
        _isLoading = false;
      });
    } on ItemException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Failed to load items: $e';
        _isLoading = false;
      });
    }
  }

  /// Loads the label catalogue for the filter bar. Failures are non-fatal:
  /// the bar is a secondary affordance, so an empty catalogue just hides it
  /// without surfacing a SnackBar on every refresh.
  Future<void> _loadLabels() async {
    try {
      _service ??= await _buildService();
      final labels = await _service!.listLabels();
      if (mounted) setState(() => _allLabels = labels);
    } catch (_) {
      if (mounted) setState(() => _allLabels = const []);
    }
  }

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> retryLoading() async {
    await _service?.dispose();
    _service = widget.service;
    _items = null;
    await _load();
    await _loadLabels();
  }

  /// Selects a bucket, collapses the chip bar, clears any active search, and
  /// reloads the list for the new view. Public so [HomePage] can drive it via a
  /// [GlobalKey] (e.g. to switch to the untriaged view after creating an item).
  void selectView(ItemView view) {
    setState(() {
      _view = view;
      _chipsExpanded = false;
      _query = '';
      _searchController.clear();
    });
    _searchFocus.unfocus();
    _load();
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

  /// Toggles membership of [name] in [_selectedLabels] and reloads the list.
  /// The server applies AND semantics: only items carrying every selected
  /// label are returned (matching the CLI `--label` repeatable flag).
  void _toggleLabel(String name) {
    setState(() {
      if (_selectedLabels.contains(name)) {
        _selectedLabels.remove(name);
      } else {
        _selectedLabels.add(name);
      }
    });
    _load();
  }

  void _openDetail(BuildContext context, Item item) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => ItemDetailPage(
          itemId: item.id,
          service: _service,
          onItemChanged: _load,
        ),
      ),
    );
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocus.dispose();
    _service?.dispose();
    super.dispose();
  }

  String _viewLabel(BuildContext context, ItemView view) {
    final l10n = AppLocalizations.of(context)!;
    switch (view) {
      case ItemView.ITEM_VIEW_TRIAGED:
        return l10n.triaged;
      case ItemView.ITEM_VIEW_UNTRIAGED:
        return l10n.untriaged;
      case ItemView.ITEM_VIEW_TIME_SENSITIVE:
        return l10n.timeSensitive;
      case ItemView.ITEM_VIEW_DONE:
        return l10n.completed;
      default:
        return l10n.triaged;
    }
  }

  Widget _buildChipBar(BuildContext context) {
    final views = const [
      ItemView.ITEM_VIEW_UNTRIAGED,
      ItemView.ITEM_VIEW_TRIAGED,
      ItemView.ITEM_VIEW_TIME_SENSITIVE,
      ItemView.ITEM_VIEW_DONE,
    ];

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: AnimatedSize(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
        alignment: Alignment.topCenter,
        child: _chipsExpanded
            ? Wrap(
                spacing: 8,
                runSpacing: 4,
                children: [
                  for (final v in views)
                    FilterChip(
                      label: Text(_viewLabel(context, v)),
                      selected: _view == v,
                      onSelected: (_) => selectView(v),
                    ),
                ],
              )
            : ActionChip(
                label: Text(_viewLabel(context, _view)),
                avatar: const Icon(Icons.filter_list),
                onPressed: () => setState(() => _chipsExpanded = true),
              ),
      ),
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

  /// A horizontally-scrollable bar of [FilterChip]s, one per known label,
  /// allowing the user to narrow the list by label (AND semantics, matching
  /// the CLI `--label` repeatable flag). Each chip shows the label's colour
  /// as a small [CircleAvatar] avatar. Returns a zero-height widget when the
  /// label catalogue is empty so the list layout is unchanged.
  Widget _buildLabelFilterBar(BuildContext context) {
    if (_allLabels.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            for (final label in _allLabels)
              Padding(
                padding: const EdgeInsets.only(right: 8),
                child: FilterChip(
                  label: Text(label.name),
                  avatar: parseLabelColour(label.colour) != null
                      ? CircleAvatar(
                          backgroundColor: parseLabelColour(label.colour),
                          maxRadius: 6,
                        )
                      : null,
                  selected: _selectedLabels.contains(label.name),
                  onSelected: (_) => _toggleLabel(label.name),
                ),
              ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
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
              onPressed: retryLoading,
              icon: const Icon(Icons.refresh),
              label: Text(AppLocalizations.of(context)!.retry),
            ),
          ],
        ),
      );
    }
    final all = _items ?? const <Item>[];
    final items = _query.isEmpty
        ? all
        : all
            .where((i) =>
                i.title.toLowerCase().contains(_query.toLowerCase()) ||
                i.description.toLowerCase().contains(_query.toLowerCase()))
            .toList();
    final l10n = AppLocalizations.of(context)!;
    // Reordering is only meaningful for the triaged view and only when no
    // search or label filter is active: anchors sent to the server reference
    // items that must be adjacent in the full ordering, which a filtered list
    // cannot guarantee.
    final canReorder = _view == ItemView.ITEM_VIEW_TRIAGED &&
        _query.isEmpty &&
        _selectedLabels.isEmpty;
    return Column(
      children: [
        _buildChipBar(context),
        _buildSearchBar(context),
        _buildLabelFilterBar(context),
        Expanded(
          child: items.isEmpty
              ? Center(
                  child: Text(
                    _query.isEmpty ? l10n.noItems : l10n.noMatchingItems,
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                )
              : RefreshIndicator(
                  onRefresh: retryLoading,
                  child: canReorder
                      ? ReorderableListView.builder(
                          itemCount: items.length,
                          buildDefaultDragHandles: false,
                          onReorderItem: (oldIndex, newIndex) =>
                              _handleReorder(items, oldIndex, newIndex),
                          itemBuilder: (context, index) {
                            final item = items[index];
                            return _buildItemTile(
                              context,
                              item,
                              index,
                              reorderable: true,
                            );
                          },
                        )
                      : ListView.builder(
                          itemCount: items.length,
                          itemBuilder: (context, index) {
                            final item = items[index];
                            return _buildItemTile(
                              context,
                              item,
                              index,
                              reorderable: false,
                            );
                          },
                        ),
                ),
        ),
      ],
    );
  }

  /// Builds a single item row. When [reorderable] is true (triaged view) the
  /// row is keyed by item id and prefixed with a drag handle wrapped in a
  /// [ReorderableDragStartListener]; otherwise it is a plain [ListTile].
  /// Tapping the row body opens [ItemDetailPage].
  Widget _buildItemTile(
    BuildContext context,
    Item item,
    int index, {
    required bool reorderable,
  }) {
    final statusIcon = statusIconFor(item);
    final labelChips = item.labels.isEmpty
        ? null
        : Wrap(
            spacing: 4,
            runSpacing: 0,
            children: [
              for (final label in item.labels)
                InputChip(
                  label: Text(label.name),
                  avatar: parseLabelColour(label.colour) != null
                      ? CircleAvatar(
                          backgroundColor: parseLabelColour(label.colour),
                          maxRadius: 6,
                        )
                      : null,
                  materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  visualDensity: VisualDensity.compact,
                ),
            ],
          );

    if (!reorderable) {
      return ListTile(
        leading: statusIcon,
        title: Text(item.title),
        subtitle: labelChips,
        trailing: item.hasDueDate() ? const Icon(Icons.timer) : null,
        onTap: () => _openDetail(context, item),
      );
    }

    return ReorderableDelayedDragStartListener(
      index: index,
      key: ValueKey(item.id),
      child: ListTile(
        leading: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            ReorderableDragStartListener(
              index: index,
              key: ValueKey('drag-handle-${item.id}'),
              child: const Icon(Icons.drag_indicator),
            ),
            const SizedBox(width: 4),
            statusIcon,
          ],
        ),
        title: Text(item.title),
        subtitle: labelChips,
        trailing: item.hasDueDate() ? const Icon(Icons.timer) : null,
        onTap: () => _openDetail(context, item),
      ),
    );
  }

  /// Handles a drag-and-drop reorder for the triaged view. The list is
  /// optimistically reordered, then [ItemService.moveItem] is called with the
  /// appropriate anchor. On failure a localised SnackBar is shown and the list
  /// is reloaded from the server to revert the optimistic change.
  ///
  /// [newIndex] is already normalised (the item at [oldIndex] conceptually
  /// removed) by [ReorderableListView.onReorderItem].
  Future<void> _handleReorder(
    List<Item> items,
    int oldIndex,
    int newIndex,
  ) async {
    if (oldIndex == newIndex) return;

    final moved = items[oldIndex];
    final reordered = List<Item>.from(items);
    reordered.removeAt(oldIndex);
    reordered.insert(newIndex, moved);

    setState(() {
      _items = reordered;
    });

    // Compute the anchor the server expects: insert before the item now at
    // newIndex, or after the previous one when dropped at the end.
    int? beforeId;
    int? afterId;
    if (newIndex >= reordered.length - 1) {
      afterId = reordered[newIndex - 1].id;
    } else {
      beforeId = reordered[newIndex + 1].id;
    }

    try {
      await _service!.moveItem(id: moved.id, beforeId: beforeId, afterId: afterId);
    } on ItemException catch (e) {
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.reorderFailed(e.message))),
        );
      }
      await _load();
    }
  }
}