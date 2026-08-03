import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_markdown_plus/flutter_markdown_plus.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/edit_item_page.dart';
import 'package:todo/widgets/item_detail_page.dart';
import 'package:todo/widgets/select_linked_items_page.dart';

/// Minimal stand-in for [ItemService] that records calls and lets each test
/// script the responses. Extends the real service so the fake never touches
/// the gRPC channel.
class _FakeItemService extends ItemService {
  _FakeItemService({Item? item, this.error})
      : _item = item,
        comments = const [],
        allLabels = const [],
        allEfforts = const [],
        allItems = const [];

  Item? _item;
  ItemException? error;

  /// All known labels returned by [listLabels]. Tests may mutate this to
  /// script the add-label picker.
  List<Label> allLabels;

  /// All known efforts returned by [listEfforts]. Tests may mutate this to
  /// script the edit-effort picker.
  List<Effort> allEfforts;

  final List<int> getItemCalls = [];
  final List<int> listCommentsCalls = [];
  final List<int> listLabelsCalls = [];
  final List<({int itemId, String body})> createCommentCalls = [];

  /// Comments returned by [listComments]. The fake also appends newly created
  /// comments here so the page's reload reflects them.
  List<Comment> comments;

  @override
  Future<Item> getItem(int id) async {
    getItemCalls.add(id);
    if (error != null) {
      throw error!;
    }
    return _item ?? Item(id: id, title: 'item $id');
  }

  @override
  Future<List<Label>> listLabels() async {
    listLabelsCalls.add(0);
    if (_listLabelsError != null) {
      throw _listLabelsError!;
    }
    return List<Label>.from(allLabels);
  }

  ItemException? _listLabelsError;

  final List<({int id, String title, String description})> updateItemCalls = [];
  ItemException? updateItemError;

  @override
  Future<Item> updateItem({
    required int id,
    required String title,
    required String description,
  }) async {
    updateItemCalls.add((id: id, title: title, description: description));
    if (updateItemError != null) {
      throw updateItemError!;
    }
    // Reflect the change so a reload shows the new title/description.
    if (_item != null) {
      _item = _item!.deepCopy()
        ..title = title
        ..description = description;
    }
    return _item ?? Item(id: id, title: title, description: description);
  }

  @override
  Future<List<Comment>> listComments(int itemId) async {
    listCommentsCalls.add(itemId);
    if (_commentsError != null) {
      throw _commentsError!;
    }
    return List<Comment>.from(comments);
  }

  ItemException? _commentsError;

  final List<({int id, List<String> add, List<String> remove})>
      updateItemLabelsCalls = [];
  ItemException? updateItemLabelsError;

  @override
  Future<Item> updateItemLabels({
    required int id,
    List<String>? add,
    List<String>? remove,
  }) async {
    updateItemLabelsCalls.add(
      (id: id, add: add ?? const [], remove: remove ?? const []),
    );
    if (updateItemLabelsError != null) {
      throw updateItemLabelsError!;
    }
    // Reflect the change in the held item so a reload shows the new state.
    if (_item != null) {
      final updated = _item!.deepCopy();
      if (remove != null) {
        updated.labels.removeWhere((l) => remove.contains(l.name));
      }
      if (add != null) {
        for (final name in add) {
          final match = allLabels.firstWhere(
            (l) => l.name == name,
            orElse: () => Label(id: -1, name: name),
          );
          if (updated.labels.every((l) => l.name != name)) {
            updated.labels.add(match);
          }
        }
      }
      _item = updated;
    }
    return _item ?? Item(id: id);
  }

  final List<({int id, String effort})> setItemEffortCalls = [];
  ItemException? setItemEffortError;
  ItemException? _listEffortsError;

  @override
  Future<List<Effort>> listEfforts() async {
    if (_listEffortsError != null) {
      throw _listEffortsError!;
    }
    return List<Effort>.from(allEfforts);
  }

  @override
  Future<Item> setEffort({required int id, required String effort}) async {
    setItemEffortCalls.add((id: id, effort: effort));
    if (setItemEffortError != null) {
      throw setItemEffortError!;
    }
    // Reflect the change so a reload shows the new effort state.
    if (_item != null) {
      final updated = _item!.deepCopy();
      if (effort.isEmpty) {
        updated.clearEffort();
      } else {
        final match = allEfforts.firstWhere(
          (e) => e.name == effort,
          orElse: () => Effort(name: effort),
        );
        updated.effort = match;
      }
      _item = updated;
    }
    return _item ?? Item(id: id);
  }

  final List<({int id, List<int> add, List<int> remove})> updateItemLinksCalls =
      [];
  ItemException? updateItemLinksError;

  /// All known items returned by [listItems]. Tests may mutate this to
  /// script the linked-items selection page.
  List<Item> allItems;

  @override
  Future<ListItemsResult> listItems({
    List<String>? labels,
    ItemView? view,
  }) async {
    return ListItemsResult(
      active: List<Item>.from(allItems),
      completed: const [],
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
    // Reflect the change so a reload shows the new linked items.
    if (_item != null && add != null) {
      final updated = _item!.deepCopy();
      for (final linkedId in add) {
        final match = allItems.firstWhere(
          (i) => i.id == linkedId,
          orElse: () => Item(id: linkedId, title: 'item $linkedId'),
        );
        if (updated.linkedItems.every((i) => i.id != linkedId)) {
          updated.linkedItems.add(match);
        }
      }
      _item = updated;
    }
    return _item ?? Item(id: id);
  }

  @override
  Future<Comment> createComment({required int itemId, required String body}) async {
    createCommentCalls.add((itemId: itemId, body: body));
    final nextId = (comments.fold<int>(0, (m, c) => c.id > m ? c.id : m)) + 1;
    // Use a timestamp strictly greater than any existing comment's so the
    // newest-first sort places the freshly added comment at the top.
    final maxSeconds = comments.fold<Int64>(
      Int64.ZERO,
      (m, c) => c.createdAt.seconds > m ? c.createdAt.seconds : m,
    );
    final created = Comment(
      id: nextId,
      body: body,
      createdAt: Timestamp(seconds: maxSeconds + Int64.ONE),
      author: 'me',
    );
    comments = [...comments, created];
    return created;
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({required _FakeItemService service, required int itemId}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: ItemDetailPage(itemId: itemId, service: service)),
  );
}

void main() {
  group('ItemDetailPage', () {
    testWidgets('shows a spinner while loading then the item title and body',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 5, title: 'Ship it', description: 'Cut a release'),
      );
      await tester.pumpWidget(_harness(service: service, itemId: 5));
      // The spinner shows before the first load resolves.
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpAndSettle();

      expect(find.byType(CircularProgressIndicator), findsNothing);
      expect(find.text('Ship it'), findsWidgets);
      expect(find.text('Cut a release'), findsOneWidget);
      expect(service.getItemCalls, [5]);
    });

    testWidgets('shows a muted hint when the description is empty',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'No desc'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      expect(find.text('No description'), findsOneWidget);
    });

    testWidgets('renders labels as chips, effort, blockers, and linked items',
        (tester) async {
      final item = Item(
        id: 1,
        title: 'Alpha',
        description: 'desc',
        labels: [
          Label(id: 1, name: 'work', colour: '#FF0000'),
          Label(id: 2, name: 'urgent'),
        ],
        effort: Effort(id: 1, name: 'high'),
        blockers: [Blocker(id: 1, description: 'waiting on review')],
        linkedItems: [Item(id: 2, title: 'Beta')],
      );
      final service = _FakeItemService(item: item);
      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Two label InputChips inside the Wrap, each with a close delete icon.
      expect(find.byType(InputChip), findsNWidgets(2));
      expect(find.widgetWithText(InputChip, 'work'), findsOneWidget);
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);
      expect(find.byIcon(Icons.close), findsNWidgets(2));
      expect(find.text('high'), findsOneWidget);
      expect(find.text('waiting on review'), findsOneWidget);
      expect(find.text('Beta'), findsOneWidget);
    });

    testWidgets('shows muted hints when optional collections are empty',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Bare'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      expect(find.text('No effort'), findsOneWidget);
      expect(find.text('No labels'), findsOneWidget);
      expect(find.text('No blockers'), findsOneWidget);
      expect(find.text('No linked items'), findsOneWidget);
    });

    testWidgets('shows the due date when present', (tester) async {
      final item = Item(
        id: 1,
        title: 'Scheduled',
        dueDate: Timestamp(seconds: Int64(0)),
        priority: 42,
      );
      final service = _FakeItemService(item: item);
      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Epoch is rendered as a localised date string starting with 1970.
      expect(find.textContaining('1970'), findsOneWidget);
      // Priority is no longer rendered on the details page.
      expect(find.text('42'), findsNothing);
    });

    testWidgets('shows noComments hint when there are no comments',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Quiet'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      expect(find.text('No comments'), findsOneWidget);
    });

    testWidgets('renders comments in descending create-date order',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Talk'))
        ..comments = [
          Comment(
            id: 1,
            body: 'oldest',
            createdAt: Timestamp(seconds: Int64(100)),
            author: 'a',
          ),
          Comment(
            id: 2,
            body: 'newest',
            createdAt: Timestamp(seconds: Int64(300)),
            author: 'b',
          ),
          Comment(
            id: 3,
            body: 'middle',
            createdAt: Timestamp(seconds: Int64(200)),
            author: 'c',
          ),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The comments section shows all three bodies.
      expect(find.text('oldest'), findsOneWidget);
      expect(find.text('middle'), findsOneWidget);
      expect(find.text('newest'), findsOneWidget);

      // Newest appears above middle, which appears above oldest. Use the
      // vertical position of the Card widgets to assert the order.
      final newestRect = tester.getRect(find.widgetWithText(Card, 'newest'));
      final middleRect = tester.getRect(find.widgetWithText(Card, 'middle'));
      final oldestRect = tester.getRect(find.widgetWithText(Card, 'oldest'));
      expect(newestRect.top, lessThan(middleRect.top));
      expect(middleRect.top, lessThan(oldestRect.top));
    });

    testWidgets('typing and sending adds a comment at the top',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Talk'))
        ..comments = [
          Comment(
            id: 1,
            body: 'old',
            createdAt: Timestamp(seconds: Int64(100)),
            author: 'a',
          ),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField), 'fresh');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      // createComment was called with the typed body.
      expect(service.createCommentCalls, [(itemId: 1, body: 'fresh')]);
      // The text field was cleared.
      final field = tester.widget<TextField>(find.byType(TextField));
      expect(field.controller!.text, isEmpty);
      // The new comment appears at the top (its card is above the old one).
      expect(find.text('fresh'), findsOneWidget);
      final freshRect = tester.getRect(find.widgetWithText(Card, 'fresh'));
      final oldRect = tester.getRect(find.widgetWithText(Card, 'old'));
      expect(freshRect.top, lessThan(oldRect.top));
      // listComments was called twice: initial load + reload after add.
      expect(service.listCommentsCalls, [1, 1]);
    });

    testWidgets('submitting an empty body shows an error and does not add',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Talk'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Submit with an empty field. The comments section may be off-screen
      // because the page content is taller than the viewport, so scroll it
      // into view first, then use testTextInput to dispatch the done action
      // (the field's onSubmitted triggers the validation).
      await tester.drag(find.byType(ListView), const Offset(0, -500));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField), '');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pump();

      // The validation error is shown (the field's errorText, not the hint).
      final field = tester.widget<TextField>(find.byType(TextField));
      expect(field.decoration!.errorText, 'Enter a comment');
      expect(service.createCommentCalls, isEmpty);
    });

    testWidgets('listComments failure shows an inline error and retry',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Talk'))
        .._commentsError = ItemException('comments down');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The page itself rendered (item title is visible).
      expect(find.text('Talk'), findsWidgets);
      // The inline comments error is shown.
      expect(find.text('comments down'), findsOneWidget);

      // Recover and retry. The comments retry button may be off-screen
      // because the page content is taller than the viewport (the Edit-effort
      // button added height to the page), so drag the ListView to bring it
      // into view, then tap the enclosing FilledButton (tapping the icon
      // directly does not always propagate when partially obscured).
      service._commentsError = null;
      await tester.drag(find.byType(ListView), const Offset(0, -500));
      await tester.pumpAndSettle();
      await tester.tap(find.ancestor(
        of: find.byIcon(Icons.refresh),
        matching: find.byType(FilledButton),
      ));
      await tester.pumpAndSettle();

      expect(find.text('comments down'), findsNothing);
      expect(find.text('No comments'), findsOneWidget);
      expect(service.listCommentsCalls, [1, 1]);
    });

    testWidgets('error state shows the message and Retry re-fetches',
        (tester) async {
      final service = _FakeItemService(error: ItemException('boom'));

      await tester.pumpWidget(_harness(service: service, itemId: 9));
      await tester.pumpAndSettle();

      expect(find.text('boom'), findsOneWidget);
      expect(find.byType(FilledButton), findsOneWidget);

      // Recover: the next getItem call succeeds.
      service.error = null;
      service._item = Item(id: 9, title: 'Recovered');

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(find.text('boom'), findsNothing);
      expect(find.text('Recovered'), findsWidgets);
      expect(service.getItemCalls, [9, 9]);
    });
  });

  group('ItemDetailPage markdown rendering', () {
    testWidgets('renders the description as GitHub-Flavored Markdown',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Doc', description: '## Heading\n\n**bold**'),
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The description is rendered through a MarkdownBody widget, not a
      // plain Text. The bold text and heading both appear as rendered text.
      expect(find.byType(MarkdownBody), findsWidgets);
      expect(find.text('bold'), findsOneWidget);
      expect(find.text('Heading'), findsOneWidget);
    });

    testWidgets('renders each comment body as GitHub-Flavored Markdown',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Talk'))
        ..comments = [
          Comment(
            id: 1,
            body: '- item one\n- item two',
            createdAt: Timestamp(seconds: Int64(100)),
            author: 'a',
          ),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Both list items render as text inside a MarkdownBody within the
      // comment card.
      expect(find.byType(MarkdownBody), findsWidgets);
      expect(find.text('item one'), findsOneWidget);
      expect(find.text('item two'), findsOneWidget);
    });

    testWidgets('shows a copy-to-clipboard button on each markdown block',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Doc', description: 'desc'))
        ..comments = [
          Comment(
            id: 1,
            body: 'remark',
            createdAt: Timestamp(seconds: Int64(100)),
            author: 'a',
          ),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // One copy button for the description and one for the comment.
      expect(find.byIcon(Icons.copy), findsNWidgets(2));

      // Tapping the description's copy button shows the confirmation SnackBar.
      await tester.tap(find.byIcon(Icons.copy).first);
      await tester.pump();
      expect(find.text('Copied to clipboard'), findsOneWidget);
    });

    testWidgets('tapping a markdown link does not crash the page',
        (tester) async {
      // url_launcher has no platform backing in widget tests, so launching
      // either no-ops or surfaces a failure SnackBar; the page must remain
      // intact in either case.
      final service = _FakeItemService(
        item: Item(
          id: 1,
          title: 'Link',
          description: '[Flutter](https://flutter.dev)',
        ),
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Tap the rendered link text. In MarkdownBody the link is a Text
      // widget; find it by the link's display text. The tap should not throw.
      expect(find.text('Flutter'), findsOneWidget);
      await tester.tap(find.text('Flutter'));
      await tester.pumpAndSettle();

      // The page is still intact: the app bar title is visible.
      expect(find.text('Link'), findsWidgets);
    });
  });

  group('ItemDetailPage label removal', () {
    testWidgets('tapping the chip delete icon optimistically removes the label',
        (tester) async {
      final item = Item(
        id: 1,
        title: 'Tagged',
        labels: [
          Label(id: 1, name: 'work', colour: '#FF0000'),
          Label(id: 2, name: 'urgent'),
        ],
      );
      final service = _FakeItemService(item: item);

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Both labels are present initially.
      expect(find.widgetWithText(InputChip, 'work'), findsOneWidget);
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);

      // Tap the delete icon on the 'work' chip.
      final workChip = find.widgetWithText(InputChip, 'work');
      await tester.tap(find.descendant(of: workChip, matching: find.byIcon(Icons.close)));
      await tester.pumpAndSettle();

      // The 'work' chip disappeared immediately (optimistic), 'urgent' remains.
      expect(find.widgetWithText(InputChip, 'work'), findsNothing);
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);
      // updateItemLabels was called to detach 'work' by name.
      expect(service.updateItemLabelsCalls, hasLength(1));
      expect(service.updateItemLabelsCalls.single.id, 1);
      expect(service.updateItemLabelsCalls.single.add, isEmpty);
      expect(service.updateItemLabelsCalls.single.remove, ['work']);
      // No failure SnackBar was shown.
      expect(find.byType(SnackBar), findsNothing);
    });

    testWidgets('a failed removal shows a SnackBar and reverts the chip',
        (tester) async {
      final item = Item(
        id: 1,
        title: 'Tagged',
        labels: [Label(id: 1, name: 'work', colour: '#FF0000')],
      );
      final service = _FakeItemService(item: item)
        ..updateItemLabelsError = ItemException('server says no');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      final chip = find.widgetWithText(InputChip, 'work');
      await tester.tap(find.descendant(of: chip, matching: find.byIcon(Icons.close)));
      await tester.pumpAndSettle();

      // The failure SnackBar is shown with the localised message.
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to remove label'), findsOneWidget);
      // The chip reappeared after the reload reverted the optimistic removal.
      expect(find.widgetWithText(InputChip, 'work'), findsOneWidget);
      // updateItemLabels was attempted.
      expect(service.updateItemLabelsCalls, hasLength(1));
      // getItem was called again to revert (initial load + reload on failure).
      expect(service.getItemCalls, [1, 1]);
    });
  });

  group('ItemDetailPage add label', () {
    testWidgets('shows an Add label button in the labels section',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Tagged'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      final button = find.widgetWithText(TextButton, 'Add label');
      expect(button, findsOneWidget);
    });

    testWidgets('tapping the button opens a dialog of unattached labels',
        (tester) async {
      final service = _FakeItemService(
        item: Item(
          id: 1,
          title: 'Tagged',
          labels: [Label(id: 1, name: 'work', colour: '#FF0000')],
        ),
      )..allLabels = [
          Label(id: 1, name: 'work', colour: '#FF0000'),
          Label(id: 2, name: 'urgent'),
          Label(id: 3, name: 'docs'),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      // The dialog lists the two unattached labels; the attached one is
      // absent from the dialog's options (it's still shown as a chip on the
      // page behind the dialog).
      final dialog = find.byType(SimpleDialog);
      expect(dialog, findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('urgent')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('docs')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('work')),
          findsNothing);
    });

    testWidgets('selecting a label attaches it and shows a confirmation',
        (tester) async {
      final service = _FakeItemService(
        item: Item(
          id: 1,
          title: 'Tagged',
          labels: [Label(id: 1, name: 'work', colour: '#FF0000')],
        ),
      )..allLabels = [
          Label(id: 1, name: 'work', colour: '#FF0000'),
          Label(id: 2, name: 'urgent'),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('urgent'));
      await tester.pumpAndSettle();

      // updateItemLabels was called to attach 'urgent' by name.
      expect(service.updateItemLabelsCalls, hasLength(1));
      expect(service.updateItemLabelsCalls.single.id, 1);
      expect(service.updateItemLabelsCalls.single.add, ['urgent']);
      expect(service.updateItemLabelsCalls.single.remove, isEmpty);
      // The confirmation SnackBar is shown.
      expect(find.text('Label added'), findsOneWidget);
      // The new chip appears after the optimistic add + reload.
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);
    });

    testWidgets('a failed add shows a SnackBar and reverts the chip',
        (tester) async {
      final service = _FakeItemService(
        item: Item(
          id: 1,
          title: 'Tagged',
          labels: [Label(id: 1, name: 'work', colour: '#FF0000')],
        ),
      )
        ..allLabels = [
            Label(id: 1, name: 'work', colour: '#FF0000'),
            Label(id: 2, name: 'urgent'),
          ]
        ..updateItemLabelsError = ItemException('server says no');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('urgent'));
      await tester.pumpAndSettle();

      // The failure SnackBar is shown.
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to add label'), findsOneWidget);
      // The optimistic 'urgent' chip was reverted by the reload.
      expect(find.widgetWithText(InputChip, 'urgent'), findsNothing);
      expect(service.updateItemLabelsCalls, hasLength(1));
      // getItem re-fetched to revert: initial load + reload on failure.
      expect(service.getItemCalls, [1, 1]);
    });

    testWidgets('shows noMoreLabels when all known labels are attached',
        (tester) async {
      final service = _FakeItemService(
        item: Item(
          id: 1,
          title: 'Tagged',
          labels: [Label(id: 1, name: 'work', colour: '#FF0000')],
        ),
      )..allLabels = [Label(id: 1, name: 'work', colour: '#FF0000')];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      // The dialog shows the no-more-labels message and no options.
      expect(find.byType(SimpleDialog), findsOneWidget);
      expect(find.text('No more labels to add'), findsOneWidget);
    });
  });

  group('ItemDetailPage effort', () {
    testWidgets('renders an Edit effort button next to the effort section',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', effort: Effort(id: 1, name: 'high')),
      )..allEfforts = [Effort(id: 1, name: 'high'), Effort(id: 2, name: 'low')];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The current effort name is rendered.
      expect(find.text('high'), findsOneWidget);
      // The Edit effort button is present.
      expect(find.widgetWithText(TextButton, 'Edit effort'), findsOneWidget);
    });

    testWidgets('tapping the button opens a dialog listing all efforts plus No effort',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', effort: Effort(id: 1, name: 'high')),
      )..allEfforts = [Effort(id: 1, name: 'high'), Effort(id: 2, name: 'low')];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      // The dialog lists every effort (including the currently-set one) plus
      // the leading "No effort" option.
      final dialog = find.byType(SimpleDialog);
      expect(dialog, findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('No effort')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('high')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('low')),
          findsOneWidget);
    });

    testWidgets('selecting an effort calls setEffort, shows a SnackBar, and updates the section',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha'),
      )..allEfforts = [Effort(id: 1, name: 'high'), Effort(id: 2, name: 'low')];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Initially no effort is set.
      expect(find.text('No effort'), findsOneWidget);

      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      // Tap the "low" option in the dialog.
      await tester.tap(find.text('low'));
      await tester.pumpAndSettle();

      // setEffort was called with the effort name.
      expect(service.setItemEffortCalls, hasLength(1));
      expect(service.setItemEffortCalls.single.id, 1);
      expect(service.setItemEffortCalls.single.effort, 'low');
      // The confirmation SnackBar is shown.
      expect(find.text('Effort updated'), findsOneWidget);
      // The section now shows the new effort name.
      expect(find.text('low'), findsOneWidget);
    });

    testWidgets('selecting No effort clears the effort', (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', effort: Effort(id: 1, name: 'high')),
      )..allEfforts = [Effort(id: 1, name: 'high')];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The effort name is shown initially.
      expect(find.text('high'), findsOneWidget);

      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      // Tap the "No effort" option (first in the dialog). There are two
      // widgets with the text "No effort" (the muted hint on the page behind
      // the dialog and the dialog option); target the one inside the dialog.
      final dialog = find.byType(SimpleDialog);
      await tester.tap(find.descendant(of: dialog, matching: find.text('No effort')));
      await tester.pumpAndSettle();

      // setEffort was called with an empty string to clear.
      expect(service.setItemEffortCalls, hasLength(1));
      expect(service.setItemEffortCalls.single.effort, isEmpty);
      // The section now shows the muted "No effort" hint.
      expect(find.text('Effort updated'), findsOneWidget);
    });

    testWidgets('a failed set shows a SnackBar and reverts the effort',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', effort: Effort(id: 1, name: 'high')),
      )
        ..allEfforts = [Effort(id: 1, name: 'high'), Effort(id: 2, name: 'low')]
        ..setItemEffortError = ItemException('server says no');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('low'));
      await tester.pumpAndSettle();

      // The failure SnackBar is shown.
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to set effort'), findsOneWidget);
      // The optimistic change was reverted by the reload: the section shows
      // the original effort name again.
      expect(find.text('high'), findsOneWidget);
      expect(service.setItemEffortCalls, hasLength(1));
      // getItem re-fetched to revert: initial load + reload on failure.
      expect(service.getItemCalls, [1, 1]);
    });

    testWidgets('a failed effort-catalogue load surfaces a SnackBar when tapped',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Alpha'))
        .._listEffortsError = ItemException('efforts unavailable');

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      // The failure is surfaced as a SnackBar rather than opening the dialog.
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to set effort'), findsOneWidget);
      expect(find.byType(SimpleDialog), findsNothing);
    });
  });

  group('ItemDetailPage linked items', () {
    testWidgets('renders an Add linked items button below the linked items section',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', linkedItems: [Item(id: 2, title: 'Beta')]),
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // The linked item is rendered and the Add linked items button is present.
      expect(find.text('Beta'), findsOneWidget);
      expect(find.widgetWithText(TextButton, 'Add linked items'), findsOneWidget);
    });

    testWidgets('tapping the button pushes SelectLinkedItemsPage', (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha', linkedItems: [Item(id: 2, title: 'Beta')]),
      )..allItems = [
          Item(id: 2, title: 'Beta'),
          Item(id: 3, title: 'Gamma'),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Add linked items'));
      await tester.pumpAndSettle();

      // The selection page is on stage. The current item (id 1) and the
      // already-linked item (id 2) are excluded; Gamma is the only candidate.
      expect(find.byType(SelectLinkedItemsPage), findsOneWidget);
      expect(find.text('Gamma'), findsOneWidget);
      expect(find.text('Beta'), findsNothing);
    });

    testWidgets('returning true reloads the detail page with the new links',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Alpha'),
      )..allItems = [
          Item(id: 2, title: 'Beta'),
          Item(id: 3, title: 'Gamma'),
        ];

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Initially no linked items.
      expect(find.text('No linked items'), findsOneWidget);

      // Open the selection page.
      await tester.tap(find.widgetWithText(TextButton, 'Add linked items'));
      await tester.pumpAndSettle();

      // Select Gamma and save.
      await tester.tap(find.text('Gamma'));
      await tester.pump();
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      // updateItemLinks was called to add Gamma (id 3).
      expect(service.updateItemLinksCalls, hasLength(1));
      expect(service.updateItemLinksCalls.single.id, 1);
      expect(service.updateItemLinksCalls.single.add, [3]);
      // The selection page popped and the detail page reloaded, now showing
      // the newly linked item.
      expect(find.byType(SelectLinkedItemsPage), findsNothing);
      expect(find.text('Gamma'), findsOneWidget);
      expect(find.text('No linked items'), findsNothing);
    });
  });

  group('ItemDetailPage edit item', () {
    testWidgets('shows a FloatingActionButton with the edit tooltip',
        (tester) async {
      final service = _FakeItemService(item: Item(id: 1, title: 'Editable'));

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      final fab = find.byType(FloatingActionButton);
      expect(fab, findsOneWidget);
      // The FAB carries the edit icon. (An Edit-effort TextButton on the
      // page also uses Icons.edit, so scope the assertion to the FAB.)
      expect(find.descendant(of: fab, matching: find.byIcon(Icons.edit)),
          findsOneWidget);
    });

    testWidgets('tapping the FAB pushes EditItemPage pre-populated',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Old title', description: 'Old description'),
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      // The edit page is on stage with the current values pre-populated.
      expect(find.byType(EditItemPage), findsOneWidget);
      expect(find.text('Old title'), findsOneWidget);
      expect(find.text('Old description'), findsOneWidget);
    });

    testWidgets('saving on the edit page reloads the detail page with the new title',
        (tester) async {
      final service = _FakeItemService(
        item: Item(id: 1, title: 'Old title', description: 'Old description'),
      );

      await tester.pumpWidget(_harness(service: service, itemId: 1));
      await tester.pumpAndSettle();

      // Open the edit page.
      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      // Replace the title and save.
      await tester.enterText(find.byType(TextField).first, 'New title');
      await tester.pump();
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      // The edit page popped and the detail page reloaded with the new title
      // in its AppBar.
      expect(find.byType(EditItemPage), findsNothing);
      expect(find.widgetWithText(AppBar, 'New title'), findsOneWidget);
      expect(service.updateItemCalls, hasLength(1));
      expect(service.updateItemCalls.single.title, 'New title');
    });
  });
}