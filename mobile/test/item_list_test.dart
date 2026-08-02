import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/item_list.dart';

/// Minimal in-memory stand-in for [ItemService] that records the [listItems]
/// calls made by [ItemList] and lets each test script the responses. Extending
/// the real service (rather than implementing an interface) keeps this test
/// resilient to internal refactors while avoiding any network initialisation:
/// the fake overrides every method the page calls and never touches the gRPC
/// channel lazily created by [ItemService._ensureInitialized].
class _FakeItemService extends ItemService {
  _FakeItemService({this.triaged = const [], this.completed = const []});

  final List<Item> triaged;
  final List<Item> completed;

  /// The views passed to each [listItems] call, in call order.
  final List<ItemView> viewsCalled = [];

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

    testWidgets('renders completed items with line-through styling', (tester) async {
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

      // The done item is rendered with a check icon and line-through text.
      expect(find.byIcon(Icons.check_circle_outline), findsOneWidget);
      final titleText = tester.widget<Text>(find.text('old release'));
      expect(titleText.style?.decoration, TextDecoration.lineThrough);
    });
  });
}