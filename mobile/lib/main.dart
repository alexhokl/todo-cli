import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_helper/ui.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/intent_service.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/efforts_page.dart';
import 'package:todo/widgets/edit_item_page.dart';
import 'package:todo/widgets/item_detail_page.dart';
import 'package:todo/widgets/item_list.dart';
import 'package:todo/widgets/labels_page.dart';
import 'package:todo/widgets/settings_page.dart';

void main() {
  runApp(const TodoApp());
}

class TodoApp extends StatelessWidget {
  const TodoApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Todo',
      theme: themesFromSeed(Colors.cyan).light,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: HomePage(title: 'Todo'),
    );
  }
}

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    required this.title,
    this.service,
    this.intentService,
  });

  final String title;

  /// Optional [ItemService] injected by tests. When null, [ItemList] and
  /// [EditItemPage] each build one lazily from the persisted backend
  /// configuration (production behaviour).
  final ItemService? service;

  /// Optional [IntentService] injected by tests. When null, a real one
  /// receives Android shortcuts, deep links and shared text.
  final IntentService? intentService;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final GlobalKey<ItemListState> _itemListKey = GlobalKey<ItemListState>();
  late final IntentService _intentService;

  @override
  void initState() {
    super.initState();
    _intentService = widget.intentService ?? IntentService();
    // Handle shortcuts, deep links and shared text from other apps
    _intentService.listen(_handleIntent);
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final intent = await _intentService.getInitialIntent();
      if (intent != null && mounted) {
        _handleIntent(intent);
      }
    });
  }

  @override
  void dispose() {
    _intentService.dispose();
    super.dispose();
  }

  void _handleIntent(AppIntent intent) {
    switch (intent) {
      case CreateItemIntent():
        _createItem(context, prefill: intent);
      case OpenViewIntent(:final view):
        _itemListKey.currentState?.selectView(view);
      case OpenItemIntent(:final id):
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => ItemDetailPage(
              itemId: id,
              service: widget.service,
              onItemChanged: () => _itemListKey.currentState?.retryLoading(),
            ),
          ),
        );
    }
  }

  /// Opens [EditItemPage] in create mode, optionally pre-filled from
  /// [prefill]. On a successful create, switches the list to the untriaged
  /// view (where new items land) and reloads it.
  Future<void> _createItem(
    BuildContext context, {
    CreateItemIntent? prefill,
  }) async {
    final created = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => EditItemPage(
          service: widget.service,
          initialTitle: prefill?.title ?? '',
          initialDescription: prefill?.description ?? '',
          initialLabels: prefill?.labels ?? const [],
        ),
      ),
    );
    if (created == true) {
      _itemListKey.currentState?.selectView(ItemView.ITEM_VIEW_UNTRIAGED);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(widget.title),
        actions: [
          if (!kIsWeb &&
              (Platform.isWindows || Platform.isMacOS || Platform.isLinux))
            IconButton(
              icon: const Icon(Icons.refresh),
              tooltip: l10n.refresh,
              onPressed: () => _itemListKey.currentState?.retryLoading(),
            ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.more_vert),
            onSelected: (value) async {
              if (value == 'labels') {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const LabelsPage()),
                );
              } else if (value == 'efforts') {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const EffortsPage()),
                );
              } else if (value == 'settings') {
                await Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const SettingsPage()),
                );
                // Re-read the persisted label-visibility setting so a toggle
                // made on the settings page takes effect on the already-mounted
                // list without a manual refresh.
                _itemListKey.currentState?.refreshShowLabels();
              }
            },
            itemBuilder: (context) => [
              PopupMenuItem(
                value: 'labels',
                child: Text(AppLocalizations.of(context)!.labels),
              ),
              PopupMenuItem(
                value: 'efforts',
                child: Text(AppLocalizations.of(context)!.efforts),
              ),
              PopupMenuItem(
                value: 'settings',
                child: Text(AppLocalizations.of(context)!.settings),
              ),
            ],
          ),
        ],
      ),
      body: ItemList(key: _itemListKey, service: widget.service),
      floatingActionButton: FloatingActionButton(
        tooltip: l10n.addItem,
        onPressed: () => _createItem(context),
        child: const Icon(Icons.add),
      ),
    );
  }
}