import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/comments_page.dart';
import 'package:todo/widgets/settings_page.dart';

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

  /// Current search query, trimmed. Empty means no filtering.
  String _query = '';
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocus = FocusNode();

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
      final result = await _service!.listItems(view: _view);
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

  Future<ItemService> _buildService() async {
    final config = await BackendConfig.load();
    return ItemService(host: config.host, port: config.port);
  }

  Future<void> retryLoading() async {
    await _service?.dispose();
    _service = widget.service;
    _items = null;
    await _load();
  }

  void _selectView(ItemView view) {
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

  void _openComments(BuildContext context, Item item) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => CommentsPage(itemId: item.id, itemTitle: item.title),
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
      ItemView.ITEM_VIEW_TRIAGED,
      ItemView.ITEM_VIEW_UNTRIAGED,
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
                      onSelected: (_) => _selectView(v),
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
    return Column(
      children: [
        _buildChipBar(context),
        _buildSearchBar(context),
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
                  child: ListView.builder(
                    itemCount: items.length,
                    itemBuilder: (context, index) {
                      final item = items[index];
                      final done = item.done;
                      return ListTile(
                        leading: Icon(
                          done
                              ? Icons.check_circle_outline
                              : Icons.circle_outlined,
                        ),
                        title: Text(
                          item.title,
                          style: done
                              ? Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    decoration: TextDecoration.lineThrough,
                                    color: Theme.of(context).disabledColor,
                                  )
                              : null,
                        ),
                        trailing: IconButton(
                          icon: const Icon(Icons.comment_outlined),
                          tooltip: AppLocalizations.of(context)!.comments,
                          onPressed: () => _openComments(context, item),
                        ),
                      );
                    },
                  ),
                ),
        ),
      ],
    );
  }
}