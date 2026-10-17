import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences_platform_interface/in_memory_shared_preferences_async.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/main.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/intent_service.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/edit_item_page.dart';

/// Minimal stand-in for [ItemService] shared by [ItemList] (via [HomePage]) and
/// [EditItemPage] in create mode. Records [listItems] views and [createItem]
/// calls. Extends the real service so the fake never touches the gRPC channel.
class _FakeItemService extends ItemService {
  _FakeItemService({this.triaged = const [], this.untriaged = const []});

  final List<Item> triaged;
  final List<Item> untriaged;

  final List<ItemView> viewsCalled = [];
  final List<({String title, String description})> createItemCalls = [];

  @override
  Future<ListItemsResult> listItems({
    List<String>? labels,
    ItemView? view,
  }) async {
    viewsCalled.add(view ?? ItemView.ITEM_VIEW_UNSPECIFIED);
    switch (view) {
      case ItemView.ITEM_VIEW_UNTRIAGED:
        return ListItemsResult(active: List<Item>.from(untriaged), completed: const []);
      default:
        return ListItemsResult(active: List<Item>.from(triaged), completed: const []);
    }
  }

  @override
  Future<Item> createItem({
    required String title,
    String description = '',
    List<String>? labels,
    String? effort,
  }) async {
    createItemCalls.add((title: title, description: description));
    return Item(id: 1, title: title, description: description);
  }

  @override
  Future<void> dispose() async {}
}

/// Stand-in for [IntentService] that returns [initial] as the launch intent
/// and lets a test deliver later intents through [deliver], without the
/// Android method channel.
class _FakeIntentService extends IntentService {
  _FakeIntentService({this.initial});

  final AppIntent? initial;
  void Function(AppIntent)? _onIntent;

  void deliver(AppIntent intent) => _onIntent?.call(intent);

  @override
  Future<AppIntent?> getInitialIntent() async => initial;

  @override
  void listen(void Function(AppIntent) onIntent) => _onIntent = onIntent;

  @override
  void dispose() => _onIntent = null;
}

Widget _harness({
  required _FakeItemService service,
  _FakeIntentService? intentService,
}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: HomePage(
      title: 'Todo',
      service: service,
      intentService: intentService ?? _FakeIntentService(),
    ),
  );
}

void main() {
  // HomePage's item list reads the persisted label-visibility setting in
  // initState, so install an in-memory SharedPreferencesAsync backend for
  // every test.
  setUp(() async {
    SharedPreferencesAsyncPlatform.instance =
        InMemorySharedPreferencesAsync.empty();
  });

  group('HomePage FAB', () {
    testWidgets('renders a floating action button with the add icon',
        (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'alpha')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.byType(FloatingActionButton), findsOneWidget);
      expect(find.byIcon(Icons.add), findsOneWidget);
    });

    testWidgets('tapping the FAB pushes EditItemPage in create mode',
        (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'alpha')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      // EditItemPage is on stage showing the create-mode AppBar title.
      expect(find.byType(EditItemPage), findsOneWidget);
      expect(find.widgetWithText(AppBar, 'Create item'), findsOneWidget);
    });

    testWidgets(
        'after a successful create, the list switches to the untriaged view',
        (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'old')],
        untriaged: [Item(id: 2, title: 'just created')],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Initially the triaged view is shown.
      expect(service.viewsCalled, [ItemView.ITEM_VIEW_TRIAGED]);
      expect(find.text('old'), findsOneWidget);

      // Open the create page and submit.
      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).first, 'just created');
      await tester.pump();
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      // createItem was called once with the entered title.
      expect(service.createItemCalls, hasLength(1));
      expect(service.createItemCalls.single.title, 'just created');

      // The list reloaded against the untriaged view, which now shows the
      // freshly created item.
      expect(
        service.viewsCalled,
        [ItemView.ITEM_VIEW_TRIAGED, ItemView.ITEM_VIEW_UNTRIAGED],
      );
      expect(find.text('just created'), findsOneWidget);
      expect(find.text('old'), findsNothing);
      // The collapsed chip reflects the untriaged view.
      final actionChip = tester.widget<ActionChip>(find.byType(ActionChip));
      final label = actionChip.label as Text;
      expect(label.data, 'Untriaged');
    });
  });

  group('HomePage intents', () {
    testWidgets('a create-item launch intent opens a pre-filled EditItemPage',
        (tester) async {
      final service = _FakeItemService();
      final intents = _FakeIntentService(
        initial: const CreateItemIntent(
          title: 'Buy milk',
          description: '2L',
        ),
      );

      await tester.pumpWidget(
        _harness(service: service, intentService: intents),
      );
      await tester.pumpAndSettle();

      expect(find.byType(EditItemPage), findsOneWidget);
      expect(find.text('Buy milk'), findsOneWidget);
      expect(find.text('2L'), findsOneWidget);
    });

    testWidgets('an open-view intent switches the list view', (tester) async {
      final service = _FakeItemService(
        untriaged: [Item(id: 2, title: 'inbox item')],
      );
      final intents = _FakeIntentService();

      await tester.pumpWidget(
        _harness(service: service, intentService: intents),
      );
      await tester.pumpAndSettle();

      intents.deliver(const OpenViewIntent(ItemView.ITEM_VIEW_UNTRIAGED));
      await tester.pumpAndSettle();

      expect(
        service.viewsCalled,
        [ItemView.ITEM_VIEW_TRIAGED, ItemView.ITEM_VIEW_UNTRIAGED],
      );
      expect(find.text('inbox item'), findsOneWidget);
    });
  });

  group('HomePage refresh button', () {
    testWidgets('renders a refresh IconButton on desktop', (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'alpha')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // A single refresh IconButton is present inside the AppBar.
      final refreshIcon = find.descendant(
        of: find.byType(AppBar),
        matching: find.byIcon(Icons.refresh),
      );
      expect(refreshIcon, findsOneWidget);
      expect(find.ancestor(of: refreshIcon, matching: find.byType(IconButton)),
          findsOneWidget);
    });

    testWidgets('tapping the refresh button reloads the list', (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'alpha')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Initial load fetches the default triaged view.
      expect(service.viewsCalled, [ItemView.ITEM_VIEW_TRIAGED]);

      await tester.tap(find.byIcon(Icons.refresh));
      await tester.pumpAndSettle();

      // retryLoading re-fetched the default triaged view.
      expect(
        service.viewsCalled,
        [ItemView.ITEM_VIEW_TRIAGED, ItemView.ITEM_VIEW_TRIAGED],
      );
    });
  });
}