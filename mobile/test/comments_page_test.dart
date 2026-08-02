import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/comments_page.dart';

/// Minimal in-memory stand-in for [ItemService] that records the calls made
/// by [CommentsPage] and lets each test script the responses. Mirrors the
/// approach used by labels_page_test.dart.
class _FakeItemService extends ItemService {
  _FakeItemService({
    List<Comment> comments = const [],
    this.listCommentsError,
  }) : _comments = comments;

  List<Comment> _comments;
  Object? listCommentsError;

  final List<({int itemId, String body})> _createCalls = [];
  final List<({int id, String body})> _updateCalls = [];
  final List<int> _deleteCalls = [];

  List<({int itemId, String body})> get createCalls =>
      List.unmodifiable(_createCalls);
  List<({int id, String body})> get updateCalls =>
      List.unmodifiable(_updateCalls);
  List<int> get deleteCalls => List.unmodifiable(_deleteCalls);

  @override
  Future<List<Comment>> listComments(int itemId) async {
    if (listCommentsError != null) {
      throw listCommentsError!;
    }
    return List<Comment>.from(_comments);
  }

  @override
  Future<Comment> createComment({
    required int itemId,
    required String body,
  }) async {
    _createCalls.add((itemId: itemId, body: body));
    final created = Comment(
      id: _comments.length + 1,
      body: body,
      author: 'testuser',
    );
    _comments = [..._comments, created];
    return created;
  }

  @override
  Future<Comment> updateComment({required int id, required String body}) async {
    _updateCalls.add((id: id, body: body));
    final index = _comments.indexWhere((c) => c.id == id);
    if (index == -1) {
      throw ItemException('comment not found');
    }
    final updated = Comment(
      id: id,
      body: body,
      author: _comments[index].author,
    );
    _comments = List<Comment>.from(_comments)..[index] = updated;
    return updated;
  }

  @override
  Future<void> deleteComment(int id) async {
    _deleteCalls.add(id);
    _comments = _comments.where((c) => c.id != id).toList();
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({required _FakeItemService service, int itemId = 7}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: CommentsPage(
      itemId: itemId,
      itemTitle: 'task',
      service: service,
    ),
  );
}

void main() {
  group('CommentsPage', () {
    testWidgets('shows loading indicator then populated list', (tester) async {
      final comments = [
        Comment(id: 1, body: 'first remark', author: 'testuser'),
        Comment(id: 2, body: 'second remark', author: 'testuser'),
      ];
      final service = _FakeItemService(comments: comments);

      await tester.pumpWidget(_harness(service: service));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpAndSettle();
      expect(find.byType(CircularProgressIndicator), findsNothing);
      expect(find.text('first remark'), findsOneWidget);
      expect(find.text('second remark'), findsOneWidget);
    });

    testWidgets('shows empty state when there are no comments', (tester) async {
      final service = _FakeItemService(comments: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();
      expect(find.text('No comments'), findsOneWidget);
    });

    testWidgets('shows error and retry button when listing fails',
        (tester) async {
      final service =
          _FakeItemService(listCommentsError: ItemException('boom'));
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.textContaining('boom'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      service.listCommentsError = null;
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      expect(find.text('No comments'), findsOneWidget);
    });

    testWidgets('create dialog creates a comment and refreshes the list',
        (tester) async {
      final service = _FakeItemService(comments: const []);
      await tester.pumpWidget(_harness(service: service, itemId: 7));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextField, 'Body'),
        'needs review',
      );
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Add comment'));
      await tester.pumpAndSettle();

      expect(service.createCalls, hasLength(1));
      expect(service.createCalls.single.itemId, equals(7));
      expect(service.createCalls.single.body, equals('needs review'));
      expect(find.text('needs review'), findsOneWidget);
      expect(find.text('Comment added'), findsOneWidget);
    });

    testWidgets('create dialog rejects empty body without calling the service',
        (tester) async {
      final service = _FakeItemService(comments: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Add comment'));
      await tester.pumpAndSettle();

      expect(service.createCalls, isEmpty);
      final bodyField = tester.widget<TextField>(
        find.widgetWithText(TextField, 'Body'),
      );
      expect(bodyField.decoration!.errorText, equals('Enter a comment'));
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('edit dialog pre-populates body and updates', (tester) async {
      final comments = [Comment(id: 5, body: 'old remark', author: 'testuser')];
      final service = _FakeItemService(comments: comments);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Edit comment'));
      await tester.pumpAndSettle();

      expect(
        tester.widget<TextField>(find.widgetWithText(TextField, 'Body')),
        isA<TextField>()
            .having((t) => t.controller!.text, 'value', equals('old remark')),
      );

      await tester.enterText(find.widgetWithText(TextField, 'Body'), 'new');
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Edit comment'));
      await tester.pumpAndSettle();

      expect(service.updateCalls, hasLength(1));
      expect(service.updateCalls.single.id, equals(5));
      expect(service.updateCalls.single.body, equals('new'));
      expect(find.text('new'), findsOneWidget);
      expect(find.text('Comment updated'), findsOneWidget);
    });

    testWidgets('delete confirmation deletes the comment on confirm',
        (tester) async {
      final comments = [Comment(id: 3, body: 'remark', author: 'testuser')];
      final service = _FakeItemService(comments: comments);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      expect(find.text('Delete this comment?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, equals(const <int>[3]));
      expect(find.text('remark'), findsNothing);
      expect(find.text('Comment deleted'), findsOneWidget);
    });

    testWidgets('delete confirmation does nothing on cancel', (tester) async {
      final comments = [Comment(id: 3, body: 'remark', author: 'testuser')];
      final service = _FakeItemService(comments: comments);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, isEmpty);
      expect(find.text('remark'), findsOneWidget);
    });
  });
}