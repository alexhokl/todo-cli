import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/select_linked_items_page.dart';

/// Minimal stand-in for [ItemService] that records [listItems] and
/// [updateItemLinks] calls and lets each test script the responses. Extends
/// the real service so the fake never touches the gRPC channel.
class _FakeItemService extends ItemService {
  _FakeItemService({this.active = const [], this.completed = const []});

  List<Item> active;
  List<Item> completed;

  final List<({int id, List<int> add, List<int> remove})> updateItemLinksCalls =
      [];
  ItemException? updateItemLinksError;
  ItemException? listItemsError;

  @override
  Future<ListItemsResult> listItems({
    List<String>? labels,
    ItemView? view,
  }) async {
    if (listItemsError != null) {
      throw listItemsError!;
    }
    return ListItemsResult(
      active: List<Item>.from(active),
      completed: List<Item>.from(completed),
    );
  }

  @override
  Future<Item> updateItemLinks({
    required int id,
    List<int>? add,
    List<int>? remove,
  }) async {
    updateItemLinksCalls.add(
      (id: id, add: add ?? const [], remove: remove ?? const []),
    );
    if (updateItemLinksError != null) {
      throw updateItemLinksError!;
    }
    return Item(id: id);
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({
  required _FakeItemService service,
  required int itemId,
  List<Item> alreadyLinked = const [],
}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: SelectLinkedItemsPage(
      itemId: itemId,
      alreadyLinked: alreadyLinked,
      service: service,
    ),
  );
}

void main() {
  group('SelectLinkedItemsPage', () {
    testWidgets('renders candidates excluding the current and already-linked items',
        (tester) async {
      final service = _FakeItemService(
        active: [
          Item(id: 1, title: 'current'),
          Item(id: 2, title: 'alpha'),
          Item(id: 3, title: 'beta'),
        ],
        completed: [Item(id: 4, title: 'done item')],
      );

      await tester.pumpWidget(_harness(
        service: service,
        itemId: 1,
        alreadyLinked: [Item(id: 3, title: 'beta')],
      ));
      await tester.pumpAndSettle();

      // The current item (id 1) and the already-linked item (id 3) are
      // excluded; the completed item (id 4) is included.
      expect(find.text('alpha'), findsOneWidget);
      expect(find.text('done item'), findsOneWidget);
      expect(find.text('current'), findsNothing);
      expect(find.text('beta'), findsNothing);
      expect(find.byType(CheckboxListTile), findsNWidgets(2));
    });

    testWidgets('tapping a checkbox toggles the selection', (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'alpha'), Item(id: 3, title: 'beta')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Initially no checkbox is checked.
      Checkbox checkbox(int id) =>
          tester.widget<Checkbox>(find.descendant(
            of: find.widgetWithText(CheckboxListTile, id == 2 ? 'alpha' : 'beta'),
            matching: find.byType(Checkbox),
          ));
      expect(checkbox(2).value, false);
      expect(checkbox(3).value, false);

      // Tap the 'alpha' tile to select it.
      await tester.tap(find.text('alpha'));
      await tester.pump();
      expect(checkbox(2).value, true);
      expect(checkbox(3).value, false);

      // Tap it again to deselect.
      await tester.tap(find.text('alpha'));
      await tester.pump();
      expect(checkbox(2).value, false);
    });

    testWidgets('save with selections calls updateItemLinks and pops',
        (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'alpha'), Item(id: 3, title: 'beta')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.text('alpha'));
      await tester.tap(find.text('beta'));
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.updateItemLinksCalls, hasLength(1));
      expect(service.updateItemLinksCalls.single.id, 1);
      expect(service.updateItemLinksCalls.single.add, containsAll([2, 3]));
      expect(service.updateItemLinksCalls.single.remove, isEmpty);
      // The page popped after the save succeeded.
      expect(find.byType(SelectLinkedItemsPage), findsNothing);
    });

    testWidgets('save with no selections pops without calling updateItemLinks',
        (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'alpha')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      // No server call was made.
      expect(service.updateItemLinksCalls, isEmpty);
      // The page popped without error.
      expect(find.byType(SelectLinkedItemsPage), findsNothing);
    });

    testWidgets('a failed save shows a SnackBar and stays open', (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'alpha')],
      )..updateItemLinksError = ItemException('server says no');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.text('alpha'));
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to add linked items'), findsOneWidget);
      // The page stays open so the user can retry.
      expect(find.byType(SelectLinkedItemsPage), findsOneWidget);
    });

    testWidgets('a listItems failure shows an error with retry', (tester) async {
      final service = _FakeItemService()
        ..listItemsError = ItemException('items unavailable');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      expect(find.textContaining('items unavailable'), findsOneWidget);
      expect(find.byIcon(Icons.refresh), findsOneWidget);

      // Recover and retry.
      service.listItemsError = null;
      service.active = [Item(id: 2, title: 'alpha')];
      await tester.tap(find.byIcon(Icons.refresh));
      await tester.pumpAndSettle();

      expect(find.text('alpha'), findsOneWidget);
    });

    testWidgets('shows the noItems empty state when there are no candidates',
        (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 1, title: 'current')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The only item is the current item, so there are no candidates.
      expect(find.text('No items'), findsOneWidget);
      expect(find.byType(CheckboxListTile), findsNothing);
    });
  });

  group('SelectLinkedItemsPage search', () {
    testWidgets('renders a search box with the localised hint', (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'alpha')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      expect(find.byType(TextField), findsOneWidget);
      expect(find.text('Search items'), findsOneWidget);
    });

    testWidgets('filters candidates by title as the query is typed',
        (tester) async {
      final service = _FakeItemService(
        active: [
          Item(id: 2, title: 'ship release'),
          Item(id: 3, title: 'write docs'),
        ],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Both candidates are visible before searching.
      expect(find.text('ship release'), findsOneWidget);
      expect(find.text('write docs'), findsOneWidget);

      await tester.enterText(find.byType(TextField), 'ship');
      await tester.pump();

      // Only the matching title remains.
      expect(find.text('ship release'), findsOneWidget);
      expect(find.text('write docs'), findsNothing);
    });

    testWidgets('matching is case-insensitive', (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'Ship It')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'SHIP');
      await tester.pump();

      expect(find.text('Ship It'), findsOneWidget);
    });

    testWidgets('shows the no-matching-items empty state when the query yields nothing',
        (tester) async {
      final service = _FakeItemService(
        active: [Item(id: 2, title: 'ship it')],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'zzz');
      await tester.pump();

      expect(find.text('No matching items'), findsOneWidget);
      expect(find.text('ship it'), findsNothing);
      expect(find.byType(CheckboxListTile), findsNothing);
    });

    testWidgets('clearing the query restores all candidates', (tester) async {
      final service = _FakeItemService(
        active: [
          Item(id: 2, title: 'ship it'),
          Item(id: 3, title: 'write docs'),
        ],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'ship');
      await tester.pump();
      expect(find.text('write docs'), findsNothing);

      // Tap the clear suffix icon to reset the query.
      await tester.tap(find.byIcon(Icons.clear));
      await tester.pump();

      expect(find.text('ship it'), findsOneWidget);
      expect(find.text('write docs'), findsOneWidget);
    });

    testWidgets('selections persist across search filtering', (tester) async {
      final service = _FakeItemService(
        active: [
          Item(id: 2, title: 'alpha'),
          Item(id: 3, title: 'beta'),
        ],
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Select 'alpha'.
      await tester.tap(find.text('alpha'));
      await tester.pump();
      final alphaCheckbox = tester.widget<Checkbox>(find.descendant(
        of: find.widgetWithText(CheckboxListTile, 'alpha'),
        matching: find.byType(Checkbox),
      ));
      expect(alphaCheckbox.value, true);

      // Search for 'beta' so 'alpha' is filtered out of the visible list.
      await tester.enterText(find.byType(TextField), 'beta');
      await tester.pump();
      expect(find.text('alpha'), findsNothing);
      expect(find.widgetWithText(CheckboxListTile, 'beta'), findsOneWidget);

      // Clear the search; 'alpha' reappears and is still selected.
      await tester.tap(find.byIcon(Icons.clear));
      await tester.pump();
      final alphaCheckboxAfter = tester.widget<Checkbox>(find.descendant(
        of: find.widgetWithText(CheckboxListTile, 'alpha'),
        matching: find.byType(Checkbox),
      ));
      expect(alphaCheckboxAfter.value, true);
    });
  });
}