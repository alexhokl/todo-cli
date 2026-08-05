import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/edit_item_page.dart';

/// Minimal stand-in for [ItemService] that records [updateItem] and
/// [createItem] calls and lets each test script the responses. Extends the
/// real service so the fake never touches the gRPC channel.
class _FakeItemService extends ItemService {
  _FakeItemService({
    this.updateItemError,
    this.createItemError,
    List<Label> labels = const [],
    this.listLabelsError,
    List<Effort> efforts = const [],
    this.listEffortsError,
  })  : _labels = labels,
        _efforts = efforts;

  final List<({int id, String title, String description})> updateItemCalls = [];
  final ItemException? updateItemError;

  final List<
      ({
        String title,
        String description,
        List<String> labels,
        String effort
      })> createItemCalls = [];
  final ItemException? createItemError;

  final List<Label> _labels;
  final Object? listLabelsError;

  final List<Effort> _efforts;
  final Object? listEffortsError;

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
    return Item(id: id, title: title, description: description);
  }

  @override
  Future<Item> createItem({
    required String title,
    String description = '',
    List<String>? labels,
    String? effort,
  }) async {
    createItemCalls.add((
      title: title,
      description: description,
      labels: labels ?? const <String>[],
      effort: effort ?? '',
    ));
    if (createItemError != null) {
      throw createItemError!;
    }
    return Item(id: 1, title: title, description: description);
  }

  @override
  Future<List<Label>> listLabels() async {
    if (listLabelsError != null) {
      throw listLabelsError!;
    }
    return List<Label>.from(_labels);
  }

  @override
  Future<List<Effort>> listEfforts() async {
    if (listEffortsError != null) {
      throw listEffortsError!;
    }
    return List<Effort>.from(_efforts);
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({
  required _FakeItemService service,
  int? itemId,
  String initialTitle = '',
  String initialDescription = '',
}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: EditItemPage(
      itemId: itemId,
      initialTitle: initialTitle,
      initialDescription: initialDescription,
      service: service,
    ),
  );
}

void main() {
  group('EditItemPage', () {
    testWidgets('pre-populates the title and description fields',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(
        service: service,
        itemId: 1,
        initialTitle: 'current title',
        initialDescription: 'current description',
      ));

      expect(find.text('current title'), findsOneWidget);
      expect(find.text('current description'), findsOneWidget);
    });

    testWidgets('submitting with a valid title calls updateItem and pops',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(
        service: service,
        itemId: 7,
        initialTitle: 'old',
        initialDescription: 'old desc',
      ));

      await tester.enterText(
        find.widgetWithText(TextField, 'old').first,
        'new title',
      );
      await tester.enterText(
        find.widgetWithText(TextField, 'old desc').first,
        'new description',
      );
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.updateItemCalls, hasLength(1));
      expect(service.updateItemCalls.single.id, 7);
      expect(service.updateItemCalls.single.title, 'new title');
      expect(service.updateItemCalls.single.description, 'new description');
      // The page popped after the save succeeded.
      expect(find.byType(EditItemPage), findsNothing);
    });

    testWidgets('submitting an empty title shows an error and does not save',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(
        service: service,
        itemId: 1,
        initialTitle: 'old',
      ));

      await tester.enterText(find.byType(TextField).first, '   ');
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.text('Title is required'), findsOneWidget);
      expect(service.updateItemCalls, isEmpty);
      // The page is still on stage.
      expect(find.byType(EditItemPage), findsOneWidget);
    });

    testWidgets('a failed update shows a SnackBar and stays open',
        (tester) async {
      final service =
          _FakeItemService(updateItemError: ItemException('server says no'));

      await tester.pumpWidget(_harness(
        service: service,
        itemId: 1,
        initialTitle: 'old',
      ));

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to update item'), findsOneWidget);
      // The page stays open so the user can retry.
      expect(find.byType(EditItemPage), findsOneWidget);
    });
  });

  group('EditItemPage (create mode)', () {
    testWidgets('shows the create title and empty fields', (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(service: service));

      // AppBar shows the create-mode title.
      expect(find.widgetWithText(AppBar, 'Create item'), findsOneWidget);
      // Both fields start empty.
      final fields = tester.widgetList<TextField>(find.byType(TextField));
      expect(fields.first.controller!.text, isEmpty);
      expect(fields.last.controller!.text, isEmpty);
    });

    testWidgets('submitting with a valid title calls createItem and pops',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(service: service));

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.enterText(find.byType(TextField).last, 'a description');
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.createItemCalls, hasLength(1));
      expect(service.createItemCalls.single.title, 'new item');
      expect(service.createItemCalls.single.description, 'a description');
      expect(service.createItemCalls.single.labels, isEmpty);
      expect(service.createItemCalls.single.effort, isEmpty);
      // updateItem is never called in create mode.
      expect(service.updateItemCalls, isEmpty);
      // The page popped after the create succeeded.
      expect(find.byType(EditItemPage), findsNothing);
    });

    testWidgets('submitting an empty title shows an error and does not create',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(service: service));

      await tester.enterText(find.byType(TextField).first, '   ');
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.text('Title is required'), findsOneWidget);
      expect(service.createItemCalls, isEmpty);
      // The page is still on stage.
      expect(find.byType(EditItemPage), findsOneWidget);
    });

    testWidgets('a failed create shows a SnackBar and stays open',
        (tester) async {
      final service =
          _FakeItemService(createItemError: ItemException('server says no'));

      await tester.pumpWidget(_harness(service: service));

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.pump();

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to create item'), findsOneWidget);
      // The page stays open so the user can retry.
      expect(find.byType(EditItemPage), findsOneWidget);
    });
  });

  group('EditItemPage (create mode labels)', () {
    testWidgets('shows the labels section with an Add label button',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.text('Labels'), findsOneWidget);
      expect(find.widgetWithText(TextButton, 'Add label'), findsOneWidget);
      // No labels selected yet -> the empty hint is shown.
      expect(find.text('No labels'), findsOneWidget);
    });

    testWidgets(
        'tapping Add label opens a dialog of known labels and selecting one '
        'adds a chip and submits the label name', (tester) async {
      final service = _FakeItemService(labels: [
        Label(id: 1, name: 'work', colour: '#FF0000'),
        Label(id: 2, name: 'urgent'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      final dialog = find.byType(SimpleDialog);
      expect(dialog, findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('work')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('urgent')),
          findsOneWidget);

      await tester.tap(find.text('urgent'));
      await tester.pumpAndSettle();

      // A chip for the selected label appears.
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.createItemCalls, hasLength(1));
      expect(service.createItemCalls.single.labels, ['urgent']);
    });

    testWidgets('already-selected labels are excluded from the picker',
        (tester) async {
      final service = _FakeItemService(labels: [
        Label(id: 1, name: 'work', colour: '#FF0000'),
        Label(id: 2, name: 'urgent'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('work'));
      await tester.pumpAndSettle();

      // Open the picker again: 'work' is excluded, only 'urgent' remains.
      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      final dialog = find.byType(SimpleDialog);
      expect(find.descendant(of: dialog, matching: find.text('work')),
          findsNothing);
      expect(find.descendant(of: dialog, matching: find.text('urgent')),
          findsOneWidget);
    });

    testWidgets('deleting a chip removes the label from the submission',
        (tester) async {
      final service = _FakeItemService(labels: [
        Label(id: 1, name: 'work', colour: '#FF0000'),
        Label(id: 2, name: 'urgent'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('work'));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('urgent'));
      await tester.pumpAndSettle();

      // Two chips selected.
      expect(find.widgetWithText(InputChip, 'work'), findsOneWidget);
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);

      // Remove 'work' via the chip's delete icon.
      await tester.ensureVisible(find.widgetWithText(InputChip, 'work'));
      final workChip = find.widgetWithText(InputChip, 'work');
      await tester.tap(
        find.descendant(of: workChip, matching: find.byIcon(Icons.close)),
      );
      await tester.pumpAndSettle();

      expect(find.widgetWithText(InputChip, 'work'), findsNothing);
      expect(find.widgetWithText(InputChip, 'urgent'), findsOneWidget);

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.createItemCalls.single.labels, ['urgent']);
    });

    testWidgets('a failed listLabels surfaces an error when Add label tapped',
        (tester) async {
      final service =
          _FakeItemService(listLabelsError: ItemException('boom'));

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      // The dialog is aborted and a SnackBar surfaces the error.
      expect(find.byType(SimpleDialog), findsNothing);
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to add label'), findsOneWidget);
    });

    testWidgets('no more labels shows the noMoreLabels message', (tester) async {
      final service = _FakeItemService(labels: [
        Label(id: 1, name: 'work'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Pick the only label, then reopen the picker.
      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('work'));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Add label'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Add label'));
      await tester.pumpAndSettle();

      expect(find.byType(SimpleDialog), findsOneWidget);
      expect(find.text('No more labels to add'), findsOneWidget);
    });
  });

  group('EditItemPage (create mode effort)', () {
    testWidgets('shows the effort section with an Edit effort button',
        (tester) async {
      final service = _FakeItemService();

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // The effort section label and the muted "No effort" hint are shown
      // initially (no effort selected).
      expect(find.text('Effort'), findsOneWidget);
      expect(find.text('No effort'), findsOneWidget);
      expect(find.widgetWithText(TextButton, 'Edit effort'), findsOneWidget);
    });

    testWidgets(
        'tapping Edit effort opens a dialog of known efforts and selecting '
        'one updates the section and submits the effort name', (tester) async {
      final service = _FakeItemService(efforts: [
        Effort(id: 1, name: 'high'),
        Effort(id: 2, name: 'low'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Edit effort'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      final dialog = find.byType(SimpleDialog);
      expect(dialog, findsOneWidget);
      // The dialog lists "No effort" plus every known effort.
      expect(find.descendant(of: dialog, matching: find.text('No effort')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('high')),
          findsOneWidget);
      expect(find.descendant(of: dialog, matching: find.text('low')),
          findsOneWidget);

      await tester.tap(find.text('low'));
      await tester.pumpAndSettle();

      // The section now shows the selected effort name (the muted "No
      // effort" hint is gone).
      expect(find.text('low'), findsOneWidget);

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.createItemCalls, hasLength(1));
      expect(service.createItemCalls.single.effort, 'low');
    });

    testWidgets('selecting No effort in the dialog submits an empty effort',
        (tester) async {
      final service = _FakeItemService(efforts: [
        Effort(id: 1, name: 'high'),
      ]);

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      // Pick 'high' first so the section shows a non-empty effort.
      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Edit effort'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('high'));
      await tester.pumpAndSettle();
      expect(find.text('high'), findsOneWidget);

      // Reopen and choose "No effort" to clear it.
      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Edit effort'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();
      final dialog = find.byType(SimpleDialog);
      await tester.tap(
        find.descendant(of: dialog, matching: find.text('No effort')),
      );
      await tester.pumpAndSettle();

      // The section reverted to the muted "No effort" hint.
      expect(find.text('high'), findsNothing);

      await tester.enterText(find.byType(TextField).first, 'new item');
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(service.createItemCalls, hasLength(1));
      expect(service.createItemCalls.single.effort, isEmpty);
    });

    testWidgets('a failed listEfforts surfaces an error when Edit effort tapped',
        (tester) async {
      final service =
          _FakeItemService(listEffortsError: ItemException('boom'));

      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.ensureVisible(
        find.widgetWithText(TextButton, 'Edit effort'),
      );
      await tester.tap(find.widgetWithText(TextButton, 'Edit effort'));
      await tester.pumpAndSettle();

      // The dialog is aborted and a SnackBar surfaces the error.
      expect(find.byType(SimpleDialog), findsNothing);
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.textContaining('Failed to set effort'), findsOneWidget);
    });
  });
}