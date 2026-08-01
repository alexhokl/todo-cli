import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/settings_page.dart';

/// Read-only list of todo items fetched from the gRPC server.
class ItemList extends StatefulWidget {
  const ItemList({super.key});

  @override
  State<ItemList> createState() => ItemListState();
}

class ItemListState extends State<ItemList> {
  List<Item>? _active;
  List<Item>? _completed;
  String? _error;
  bool _isLoading = true;
  ItemService? _service;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final config = await BackendConfig.load();
      _service = ItemService(host: config.host, port: config.port);
      final result = await _service!.listItems();
      setState(() {
        _active = result.active;
        _completed = result.completed;
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

  Future<void> retryLoading() async {
    await _service?.dispose();
    _service = null;
    _active = null;
    _completed = null;
    await _load();
  }

  @override
  void dispose() {
    _service?.dispose();
    super.dispose();
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
    final active = _active ?? const <Item>[];
    final completed = _completed ?? const <Item>[];
    if (active.isEmpty && completed.isEmpty) {
      return Center(child: Text(AppLocalizations.of(context)!.noItems));
    }
    return RefreshIndicator(
      onRefresh: retryLoading,
      child: ListView(
        children: [
          if (active.isNotEmpty) ...[
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                AppLocalizations.of(context)!.active,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ),
            for (final item in active)
              ListTile(
                leading: const Icon(Icons.circle_outlined),
                title: Text(item.title),
                subtitle:
                    item.description.isEmpty ? null : Text(item.description),
              ),
          ],
          if (completed.isNotEmpty) ...[
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Text(
                AppLocalizations.of(context)!.completed,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ),
            for (final item in completed)
              ListTile(
                leading: const Icon(Icons.check_circle_outline),
                title: Text(
                  item.title,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        decoration: TextDecoration.lineThrough,
                        color: Theme.of(context).disabledColor,
                      ),
                ),
              ),
          ],
        ],
      ),
    );
  }
}