import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/efforts_page.dart';
import 'package:todo/widgets/edit_item_page.dart';
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
      theme: ThemeData(colorScheme: ColorScheme.fromSeed(seedColor: Colors.cyan)),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: HomePage(title: 'Todo'),
    );
  }
}

class HomePage extends StatefulWidget {
  const HomePage({super.key, required this.title, this.service});

  final String title;

  /// Optional [ItemService] injected by tests. When null, [ItemList] and
  /// [EditItemPage] each build one lazily from the persisted backend
  /// configuration (production behaviour).
  final ItemService? service;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final GlobalKey<ItemListState> _itemListKey = GlobalKey<ItemListState>();

  /// Opens [EditItemPage] in create mode. On a successful create, switches the
  /// list to the untriaged view (where new items land) and reloads it.
  Future<void> _createItem(BuildContext context) async {
    final created = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => EditItemPage(service: widget.service),
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
            onSelected: (value) {
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
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const SettingsPage()),
                );
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