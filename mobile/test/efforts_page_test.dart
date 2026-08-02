import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/efforts_page.dart';

/// Minimal in-memory stand-in for [ItemService] that records the calls made
/// by [EffortsPage] and lets each test script the responses. Extending the real
/// service (rather than implementing an interface) keeps this test resilient
/// to internal refactors while avoiding any network initialisation: the
/// fake overrides every method the page calls and never touches the gRPC
/// channel lazily created by [ItemService._ensureInitialized].
class _FakeItemService extends ItemService {
  _FakeItemService({
    List<Effort> efforts = const [],
    this.createEffortResult,
    this.deleteEffortResult,
    this.listEffortsError,
  }) : _efforts = efforts;

  List<Effort> _efforts;
  Future<Effort> Function(String name)? createEffortResult;
  Future<void> Function(int id)? deleteEffortResult;
  Object? listEffortsError;

  int get createCallCount => _createCalls.length;
  final List<String> _createCalls = [];
  final List<({int id, String name})> _renameCalls = [];
  final List<int> _deleteCalls = [];

  List<String> get createCalls => List.unmodifiable(_createCalls);
  List<({int id, String name})> get renameCalls =>
      List.unmodifiable(_renameCalls);
  List<int> get deleteCalls => List.unmodifiable(_deleteCalls);

  @override
  Future<List<Effort>> listEfforts() async {
    if (listEffortsError != null) {
      throw listEffortsError!;
    }
    return List<Effort>.from(_efforts);
  }

  @override
  Future<Effort> createEffort(String name) async {
    _createCalls.add(name);
    if (createEffortResult != null) {
      final created = await createEffortResult!(name);
      _efforts = [..._efforts, created]..sort((a, b) => a.name.compareTo(b.name));
      return created;
    }
    final created = Effort(id: _efforts.length + 1, name: name);
    _efforts = [..._efforts, created]..sort((a, b) => a.name.compareTo(b.name));
    return created;
  }

  @override
  Future<Effort> renameEffort(int id, {required String name}) async {
    _renameCalls.add((id: id, name: name));
    final index = _efforts.indexWhere((e) => e.id == id);
    if (index == -1) {
      throw ItemException('effort not found');
    }
    final renamed = Effort(id: id, name: name);
    _efforts = List<Effort>.from(_efforts)..[index] = renamed;
    return renamed;
  }

  @override
  Future<void> deleteEffort(int id) async {
    _deleteCalls.add(id);
    if (deleteEffortResult != null) {
      await deleteEffortResult!(id);
      return;
    }
    _efforts = _efforts.where((e) => e.id != id).toList();
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({required _FakeItemService service}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: EffortsPage(service: service),
  );
}

void main() {
  group('EffortsPage', () {
    testWidgets('shows loading indicator then populated list', (tester) async {
      final efforts = [
        Effort(id: 1, name: 'high'),
        Effort(id: 2, name: 'low'),
      ];
      final service = _FakeItemService(efforts: efforts);

      await tester.pumpWidget(_harness(service: service));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpAndSettle();
      expect(find.byType(CircularProgressIndicator), findsNothing);
      expect(find.text('high'), findsOneWidget);
      expect(find.text('low'), findsOneWidget);
    });

    testWidgets('shows empty state when there are no efforts', (tester) async {
      final service = _FakeItemService(efforts: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();
      expect(find.text('No efforts'), findsOneWidget);
    });

    testWidgets('shows error and retry button when listing fails',
        (tester) async {
      final service = _FakeItemService(listEffortsError: ItemException('boom'));
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.textContaining('boom'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      // Recover: clear the error and retry.
      service.listEffortsError = null;
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      expect(find.text('No efforts'), findsOneWidget);
    });

    testWidgets('create dialog creates an effort and refreshes the list',
        (tester) async {
      final service = _FakeItemService(efforts: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextField, 'Name'),
        'high',
      );
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Add effort'));
      await tester.pumpAndSettle();

      expect(service.createCalls, hasLength(1));
      expect(service.createCalls.single, equals('high'));
      expect(find.text('high'), findsOneWidget);
      expect(find.text('Effort created'), findsOneWidget);
    });

    testWidgets('create dialog rejects empty name without calling the service',
        (tester) async {
      final service = _FakeItemService(efforts: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Add effort'));
      await tester.pumpAndSettle();

      expect(service.createCallCount, equals(0));
      final nameField = tester.widget<TextField>(
        find.widgetWithText(TextField, 'Name'),
      );
      expect(nameField.decoration!.errorText, equals('Enter a name'));
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('create surfaces a server error via SnackBar and keeps the list',
        (tester) async {
      final service = _FakeItemService(
        efforts: const [],
        createEffortResult: (_) async => throw ItemException('name already taken'),
      );
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(find.widgetWithText(TextField, 'Name'), 'high');
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Add effort'));
      await tester.pumpAndSettle();

      expect(service.createCallCount, equals(1));
      expect(find.textContaining('name already taken'), findsOneWidget);
      // The dialog stays open so the user can retry.
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('edit dialog pre-populates name and renames', (tester) async {
      final efforts = [Effort(id: 5, name: 'high')];
      final service = _FakeItemService(efforts: efforts);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Edit effort'));
      await tester.pumpAndSettle();

      // Pre-populated value.
      expect(
        tester.widget<TextField>(find.widgetWithText(TextField, 'Name')),
        isA<TextField>()
            .having((t) => t.controller!.text, 'value', equals('high')),
      );

      await tester.enterText(find.widgetWithText(TextField, 'Name'), 'medium');
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Edit effort'));
      await tester.pumpAndSettle();

      expect(service.renameCalls, hasLength(1));
      expect(service.renameCalls.single.id, equals(5));
      expect(service.renameCalls.single.name, equals('medium'));
      expect(find.text('medium'), findsOneWidget);
      expect(find.text('Effort updated'), findsOneWidget);
    });

    testWidgets('delete confirmation deletes the effort on confirm',
        (tester) async {
      final efforts = [Effort(id: 3, name: 'high')];
      final service = _FakeItemService(efforts: efforts);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      expect(find.text('Delete effort "high"?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, equals(const <int>[3]));
      expect(find.text('high'), findsNothing);
      expect(find.text('Effort deleted'), findsOneWidget);
    });

    testWidgets('delete confirmation does nothing on cancel', (tester) async {
      final efforts = [Effort(id: 3, name: 'high')];
      final service = _FakeItemService(efforts: efforts);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, isEmpty);
      expect(find.text('high'), findsOneWidget);
    });

    testWidgets('delete surfaces a server error via SnackBar', (tester) async {
      final efforts = [Effort(id: 3, name: 'high')];
      final service = _FakeItemService(
        efforts: efforts,
        deleteEffortResult: (_) async => throw ItemException('effort is in use'),
      );
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, equals(const <int>[3]));
      expect(find.textContaining('effort is in use'), findsOneWidget);
      // The effort is still present because the delete failed.
      expect(find.text('high'), findsOneWidget);
    });
  });
}