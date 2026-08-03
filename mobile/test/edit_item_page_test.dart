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
  _FakeItemService({this.updateItemError, this.createItemError});

  final List<({int id, String title, String description})> updateItemCalls = [];
  final ItemException? updateItemError;

  final List<({String title, String description})> createItemCalls = [];
  final ItemException? createItemError;

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
  }) async {
    createItemCalls.add((title: title, description: description));
    if (createItemError != null) {
      throw createItemError!;
    }
    return Item(id: 1, title: title, description: description);
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
}