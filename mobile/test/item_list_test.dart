import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/item_detail_page.dart';
import 'package:todo/widgets/item_list.dart';

/// Minimal in-memory stand-in for [ItemService] that records the [listItems]
/// calls made by [ItemList] and lets each test script the responses. Extending
/// the real service (rather than implementing an interface) keeps this test
/// resilient to internal refactors while avoiding any network initialisation:
/// the fake overrides every method the page calls and never touches the gRPC
/// channel lazily created by [ItemService._ensureInitialized].
class _FakeItemService extends ItemService {
  _FakeItemService({
    this.triaged = const [],
    this.completed = const [],
    this.moveItemError,
  });

  final List<Item> triaged;
  final List<Item> completed;

  /// When non-null, the next [moveItem] call throws this [ItemException].
  final ItemException? moveItemError;

  /// The views passed to each [listItems] call, in call order.
  final List<ItemView> viewsCalled = [];

  /// Each recorded [moveItem] call: the moved id and the anchor arguments.
  final List<({int id, int? beforeId, int? afterId})> moveCalls = [];

  @override
  Future<ListItemsResult> listItems({
    List<String>? labels,
    ItemView? view,
  }) async {
    viewsCalled.add(view ?? ItemView.ITEM_VIEW_UNSPECIFIED);
    switch (view) {
      case ItemView.ITEM_VIEW_DONE:
        return ListItemsResult(active: List<Item>.from(completed), completed: const []);
      default:
        return ListItemsResult(active: List<Item>.from(triaged), completed: const []);
    }
  }

  @override
  Future<Item> moveItem({
    required int id,
    int? beforeId,
    int? afterId,
    bool top = false,
    bool bottom = false,
    bool changeList = false,
    int? listId,
  }) async {
    moveCalls.add((id: id, beforeId: beforeId, afterId: afterId));
    if (moveItemError != null) {
      throw moveItemError!;
    }
    return Item(id: id);
  }

  @override
  Future<Item> getItem(int id) async {
    return Item(id: id, title: 'item $id');
  }

  @override
  Future<List<Comment>> listComments(int itemId) async {
    return const <Comment>[];
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({required _FakeItemService service}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: ItemList(service: service)),
  );
}

void main() {
  group('ItemList chip bar', () {
    testWidgets('shows the collapsed triaged chip by default and lists triaged items', (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'ship it')],
        completed: [Item(id: 2, title: 'old release', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // The collapsed chip shows the default view's label.
      expect(find.text('Triaged'), findsOneWidget);
      // The triaged item is rendered; the completed one is not (only one bucket is fetched).
      expect(find.text('ship it'), findsOneWidget);
      expect(find.text('old release'), findsNothing);
      // The listItems call was made with the default triaged view.
      expect(service.viewsCalled, [ItemView.ITEM_VIEW_TRIAGED]);
      // Only the collapsed ActionChip is present; no FilterChips yet.
      expect(find.byType(ActionChip), findsOneWidget);
      expect(find.byType(FilterChip), findsNothing);
    });

    testWidgets('expands into four FilterChips when the collapsed chip is tapped', (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'ship it')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();

      // All four bucket chips are now visible.
      expect(find.byType(FilterChip), findsNWidgets(4));
      expect(find.text('Triaged'), findsOneWidget);
      expect(find.text('Untriaged'), findsOneWidget);
      expect(find.text('Time-sensitive'), findsOneWidget);
      expect(find.text('Completed'), findsOneWidget);
      // The collapsed ActionChip is gone.
      expect(find.byType(ActionChip), findsNothing);
    });

    testWidgets('renders Untriaged before Triaged in the expanded chip bar',
        (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'ship it')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();

      // The FilterChips render in the order of the views list, so the
      // Untriaged chip appears before the Triaged chip.
      final chips = tester.widgetList<FilterChip>(find.byType(FilterChip)).toList();
      expect(chips.length, 4);
      final labels = chips.map((c) => (c.label as Text).data).toList();
      expect(labels.indexOf('Untriaged'), lessThan(labels.indexOf('Triaged')));
    });

    testWidgets('selecting the Completed chip switches the view and collapses the bar', (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'ship it')],
        completed: [Item(id: 2, title: 'old release', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Expand the chip bar.
      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();

      // Tap the Completed FilterChip. There are two widgets with the text
      // "Completed" (the chip label and the listItems default), so find the
      // FilterChip specifically.
      await tester.tap(find.widgetWithText(FilterChip, 'Completed'));
      await tester.pumpAndSettle();

      // The view switched to DONE and the completed item now renders.
      expect(service.viewsCalled, [ItemView.ITEM_VIEW_TRIAGED, ItemView.ITEM_VIEW_DONE]);
      expect(find.text('old release'), findsOneWidget);
      expect(find.text('ship it'), findsNothing);

      // The bar collapsed back to a single ActionChip showing the new view.
      expect(find.byType(ActionChip), findsOneWidget);
      expect(find.byType(FilterChip), findsNothing);
      // The collapsed chip now reflects the selected view. The ActionChip's
      // label text is "Completed".
      final actionChip = tester.widget<ActionChip>(find.byType(ActionChip));
      final label = actionChip.label as Text;
      expect(label.data, 'Completed');
    });

    testWidgets('shows the empty state when the selected bucket has no items', (tester) async {
      final service = _FakeItemService(triaged: const []);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.text('No items'), findsOneWidget);
      // The chip bar is still visible above the empty state.
      expect(find.byType(ActionChip), findsOneWidget);
    });

    testWidgets('renders completed items with a check icon', (tester) async {
      final service = _FakeItemService(
        completed: [Item(id: 2, title: 'old release', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Switch to the Completed view.
      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilterChip, 'Completed'));
      await tester.pumpAndSettle();

      // The done item is rendered with a check icon.
      expect(find.byIcon(Icons.check_circle_outline), findsOneWidget);
      expect(find.text('old release'), findsOneWidget);
    });
  });

  group('ItemList status icons', () {
    testWidgets('renders the triaged status icon for an item with priority',
        (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'triaged', priority: 1.0)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.low_priority), findsOneWidget);
      expect(find.text('triaged'), findsOneWidget);
    });

    testWidgets('renders the untriaged status icon for an item without priority',
        (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 2, title: 'untriaged')],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.radio_button_unchecked), findsOneWidget);
      expect(find.text('untriaged'), findsOneWidget);
    });

    testWidgets('renders the done status icon for a completed item',
        (tester) async {
      final service = _FakeItemService(
        completed: [Item(id: 3, title: 'done', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Switch to the Completed view.
      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilterChip, 'Completed'));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.check_circle_outline), findsOneWidget);
      expect(find.text('done'), findsOneWidget);
    });
  });

  group('ItemList search box', () {
    testWidgets('renders a search box directly under the chip bar', (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'ship it')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // The search field is present with the localised hint.
      expect(find.byType(TextField), findsOneWidget);
      expect(find.text('Search items'), findsOneWidget);
    });

    testWidgets('filters items by title as the query is typed', (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 1, title: 'ship release'),
        Item(id: 2, title: 'write docs'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Both items are visible before searching.
      expect(find.text('ship release'), findsOneWidget);
      expect(find.text('write docs'), findsOneWidget);

      await tester.enterText(find.byType(TextField), 'ship');
      await tester.pump();

      // Only the matching title remains.
      expect(find.text('ship release'), findsOneWidget);
      expect(find.text('write docs'), findsNothing);
    });

    testWidgets('matches against the description as well as the title', (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 1, title: 'alpha', description: 'fix the login bug'),
        Item(id: 2, title: 'beta', description: 'polish the UI'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'login');
      await tester.pump();

      expect(find.text('alpha'), findsOneWidget);
      expect(find.text('beta'), findsNothing);
    });

    testWidgets('matching is case-insensitive', (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 1, title: 'Ship It'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'SHIP');
      await tester.pump();

      expect(find.text('Ship It'), findsOneWidget);
    });

    testWidgets('shows the no-matching-items empty state when the query yields nothing', (tester) async {
      final service = _FakeItemService(triaged: [Item(id: 1, title: 'ship it')]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'zzz');
      await tester.pump();

      expect(find.text('No matching items'), findsOneWidget);
      expect(find.text('ship it'), findsNothing);
    });

    testWidgets('clearing the query restores all items', (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 1, title: 'ship it'),
        Item(id: 2, title: 'write docs'),
      ]);

      await tester.pumpWidget(_harness(service: service));
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

    testWidgets('switching the view clears the query', (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'ship it')],
        completed: [Item(id: 2, title: 'old release', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'ship');
      await tester.pump();
      expect(find.text('ship it'), findsOneWidget);

      // Switch to the Completed view.
      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilterChip, 'Completed'));
      await tester.pumpAndSettle();

      // The search field is empty again and the completed item is visible.
      final field = tester.widget<TextField>(find.byType(TextField));
      expect(field.controller!.text, isEmpty);
      expect(find.text('old release'), findsOneWidget);
      expect(find.text('ship it'), findsNothing);
    });
  });

  group('ItemList drag-and-drop reordering', () {
    testWidgets(
        'shows a drag handle on triaged items but not on other views',
        (tester) async {
      final service = _FakeItemService(
        triaged: [Item(id: 1, title: 'alpha')],
        completed: [Item(id: 2, title: 'old release', done: true)],
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Triaged view renders the drag handle.
      expect(find.byIcon(Icons.drag_indicator), findsOneWidget);

      // Switch to the Completed view, which does not support reordering.
      await tester.tap(find.byType(ActionChip));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilterChip, 'Completed'));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.drag_indicator), findsNothing);
    });

    testWidgets('reordering down calls moveItem with beforeId anchor',
        (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 10, title: 'alpha'),
        Item(id: 20, title: 'beta'),
        Item(id: 30, title: 'gamma'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Drag the first tile below the third tile. The drag handle is at the
      // left of each row; target the handle of item 0 and drop below item 2.
      await tester.drag(
        find.byIcon(Icons.drag_indicator).first,
        const Offset(0, 250),
      );
      await tester.pumpAndSettle();

      // The optimistic reorder moved alpha to the end, so the server should
      // have been asked to place item 10 after item 30 (no beforeId).
      expect(service.moveCalls, [
        (id: 10, beforeId: null, afterId: 30),
      ]);
    });

    testWidgets('reordering up calls moveItem with beforeId anchor',
        (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 10, title: 'alpha'),
        Item(id: 20, title: 'beta'),
        Item(id: 30, title: 'gamma'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Drag the last tile above the first tile.
      await tester.drag(
        find.byIcon(Icons.drag_indicator).last,
        const Offset(0, -250),
      );
      await tester.pumpAndSettle();

      // gamma moved to the top, so the server should have been asked to place
      // item 30 before item 10 (no afterId).
      expect(service.moveCalls, [
        (id: 30, beforeId: 10, afterId: null),
      ]);
    });

    testWidgets(
        'a failed moveItem shows a SnackBar and reloads the list',
        (tester) async {
      final service = _FakeItemService(
        triaged: [
          Item(id: 10, title: 'alpha'),
          Item(id: 20, title: 'beta'),
        ],
        moveItemError: ItemException('server unavailable'),
      );

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.drag(
        find.byIcon(Icons.drag_indicator).first,
        const Offset(0, 150),
      );
      await tester.pumpAndSettle();

      // The localised SnackBar is shown.
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to reorder item'), findsOneWidget);
      // moveItem was attempted and the list was reloaded from the service,
      // resetting the optimistic order.
      expect(service.moveCalls, hasLength(1));
      expect(service.viewsCalled, [ItemView.ITEM_VIEW_TRIAGED, ItemView.ITEM_VIEW_TRIAGED]);
      expect(find.text('alpha'), findsOneWidget);
      expect(find.text('beta'), findsOneWidget);
    });

    testWidgets('reordering is disabled while a search is active',
        (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 10, title: 'alpha'),
        Item(id: 20, title: 'beta'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'alpha');
      await tester.pump();

      // Only the filtered item is shown and no drag handle is rendered.
      expect(find.byIcon(Icons.drag_indicator), findsNothing);
    });
  });

  group('ItemList navigation to detail page', () {
    testWidgets('tapping a row pushes ItemDetailPage', (tester) async {
      final service = _FakeItemService(triaged: [
        Item(id: 10, title: 'alpha'),
        Item(id: 20, title: 'beta'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // No detail page is on stage yet.
      expect(find.byType(ItemDetailPage), findsNothing);

      // Tap the first row.
      await tester.tap(find.text('alpha'));
      await tester.pumpAndSettle();

      // The ItemDetailPage is pushed. Its app bar shows the freshly fetched
      // item title from getItem ('item 10'), not the stale list title.
      expect(find.byType(ItemDetailPage), findsOneWidget);
      expect(find.widgetWithText(AppBar, 'item 10'), findsOneWidget);
    });
  });
}