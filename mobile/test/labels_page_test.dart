import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/proto/item.pb.dart';
import 'package:todo/services/item_service.dart';
import 'package:todo/widgets/labels_page.dart';

/// Minimal in-memory stand-in for [ItemService] that records the calls made
/// by [LabelsPage] and lets each test script the responses. Extending the real
/// service (rather than implementing an interface) keeps this test resilient
/// to internal refactors while avoiding any network initialisation: the
/// fake overrides every method the page calls and never touches the gRPC
/// channel lazily created by [ItemService._ensureInitialized].
class _FakeItemService extends ItemService {
  _FakeItemService({
    List<Label> labels = const [],
    this.createLabelResult,
    this.deleteLabelResult,
    this.listLabelsError,
  }) : _labels = labels;

  List<Label> _labels;
  Future<Label> Function(String name, {String? colour})? createLabelResult;
  Future<void> Function(int id)? deleteLabelResult;
  Object? listLabelsError;

  int get createCallCount => _createCalls.length;
  final List<({String name, String? colour})> _createCalls = [];
  final List<({int id, String? name, String? colour})> _renameCalls = [];
  final List<int> _deleteCalls = [];

  List<({String name, String? colour})> get createCalls =>
      List.unmodifiable(_createCalls);
  List<({int id, String? name, String? colour})> get renameCalls =>
      List.unmodifiable(_renameCalls);
  List<int> get deleteCalls => List.unmodifiable(_deleteCalls);

  @override
  Future<List<Label>> listLabels() async {
    if (listLabelsError != null) {
      throw listLabelsError!;
    }
    return List<Label>.from(_labels);
  }

  @override
  Future<Label> createLabel(String name, {String? colour}) async {
    _createCalls.add((name: name, colour: colour));
    if (createLabelResult != null) {
      final created = await createLabelResult!(name, colour: colour);
      _labels = [..._labels, created]..sort((a, b) => a.name.compareTo(b.name));
      return created;
    }
    final created = Label(id: _labels.length + 1, name: name, colour: colour ?? '#FFFF00');
    _labels = [..._labels, created]..sort((a, b) => a.name.compareTo(b.name));
    return created;
  }

  @override
  Future<Label> renameLabel(int id, {String? name, String? colour}) async {
    _renameCalls.add((id: id, name: name, colour: colour));
    final index = _labels.indexWhere((l) => l.id == id);
    if (index == -1) {
      throw ItemException('label not found');
    }
    final existing = _labels[index];
    final renamed = Label(
      id: id,
      name: name ?? existing.name,
      colour: colour ?? existing.colour,
    );
    _labels = List<Label>.from(_labels)..[index] = renamed;
    return renamed;
  }

  @override
  Future<void> deleteLabel(int id) async {
    _deleteCalls.add(id);
    if (deleteLabelResult != null) {
      await deleteLabelResult!(id);
      return;
    }
    _labels = _labels.where((l) => l.id != id).toList();
  }

  @override
  Future<void> dispose() async {}
}

Widget _harness({required _FakeItemService service}) {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: LabelsPage(service: service),
  );
}

void main() {
  group('LabelsPage', () {
    testWidgets('shows loading indicator then populated list', (tester) async {
      final labels = [
        Label(id: 1, name: 'urgent', colour: '#FF0000'),
        Label(id: 2, name: 'work', colour: '#0000FF'),
      ];
      final service = _FakeItemService(labels: labels);

      await tester.pumpWidget(_harness(service: service));
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      await tester.pumpAndSettle();
      expect(find.byType(CircularProgressIndicator), findsNothing);
      expect(find.text('urgent'), findsOneWidget);
      expect(find.text('work'), findsOneWidget);
    });

    testWidgets('shows empty state when there are no labels', (tester) async {
      final service = _FakeItemService(labels: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();
      expect(find.text('No labels'), findsOneWidget);
    });

    testWidgets('shows error and retry button when listing fails',
        (tester) async {
      final service = _FakeItemService(listLabelsError: ItemException('boom'));
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      expect(find.textContaining('boom'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      // Recover: clear the error and retry.
      service.listLabelsError = null;
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      expect(find.text('No labels'), findsOneWidget);
    });

    testWidgets('create dialog creates a label and refreshes the list',
        (tester) async {
      final service = _FakeItemService(labels: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextField, 'Name'),
        'urgent',
      );
      await tester.pumpAndSettle();
      // Default colour is already populated (#FFFF00); submit as-is.
      await tester.tap(find.widgetWithText(FilledButton, 'Add label'));
      await tester.pumpAndSettle();

      expect(service.createCalls, hasLength(1));
      expect(service.createCalls.single.name, equals('urgent'));
      expect(service.createCalls.single.colour, equals('#FFFF00'));
      expect(find.text('urgent'), findsOneWidget);
      expect(find.text('Label created'), findsOneWidget);
    });

    testWidgets('create dialog rejects invalid colour without calling the service',
        (tester) async {
      final service = _FakeItemService(labels: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(find.widgetWithText(TextField, 'Name'), 'urgent');
      await tester.enterText(
        find.widgetWithText(TextField, 'Colour'),
        'red',
      );
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Add label'));
      await tester.pumpAndSettle();

      expect(service.createCallCount, equals(0));
      expect(find.text('Colour must be in #RRGGBB format'), findsOneWidget);
      // The dialog stays open for the user to retry.
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('create dialog rejects empty name without calling the service',
        (tester) async {
      final service = _FakeItemService(labels: const []);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Add label'));
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
        labels: const [],
        createLabelResult: (_, {colour}) async =>
            throw ItemException('name already taken'),
      );
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      await tester.enterText(find.widgetWithText(TextField, 'Name'), 'urgent');
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Add label'));
      await tester.pumpAndSettle();

      expect(service.createCallCount, equals(1));
      expect(find.textContaining('name already taken'), findsOneWidget);
      // The dialog stays open so the user can retry.
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('edit dialog pre-populates name and colour and renames',
        (tester) async {
      final labels = [Label(id: 5, name: 'urgent', colour: '#FF0000')];
      final service = _FakeItemService(labels: labels);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Edit label'));
      await tester.pumpAndSettle();

      // Pre-populated values.
      expect(
        tester.widget<TextField>(find.widgetWithText(TextField, 'Name')),
        isA<TextField>()
            .having((t) => t.controller!.text, 'value', equals('urgent')),
      );
      expect(
        tester.widget<TextField>(find.widgetWithText(TextField, 'Colour')),
        isA<TextField>()
            .having((t) => t.controller!.text, 'value', equals('#FF0000')),
      );

      await tester.enterText(find.widgetWithText(TextField, 'Name'), 'work');
      await tester.enterText(
        find.widgetWithText(TextField, 'Colour'),
        '#00FF00',
      );
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(FilledButton, 'Edit label'));
      await tester.pumpAndSettle();

      expect(service.renameCalls, hasLength(1));
      expect(service.renameCalls.single.id, equals(5));
      expect(service.renameCalls.single.name, equals('work'));
      expect(service.renameCalls.single.colour, equals('#00FF00'));
      expect(find.text('work'), findsOneWidget);
      expect(find.text('Label updated'), findsOneWidget);
    });

    testWidgets('delete confirmation deletes the label on confirm',
        (tester) async {
      final labels = [Label(id: 3, name: 'urgent', colour: '#FF0000')];
      final service = _FakeItemService(labels: labels);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      expect(find.text('Delete label "urgent"?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, equals(const <int>[3]));
      expect(find.text('urgent'), findsNothing);
      expect(find.text('Label deleted'), findsOneWidget);
    });

    testWidgets('delete confirmation does nothing on cancel', (tester) async {
      final labels = [Label(id: 3, name: 'urgent', colour: '#FF0000')];
      final service = _FakeItemService(labels: labels);
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, isEmpty);
      expect(find.text('urgent'), findsOneWidget);
    });

    testWidgets('delete surfaces a server error via SnackBar', (tester) async {
      final labels = [Label(id: 3, name: 'urgent', colour: '#FF0000')];
      final service = _FakeItemService(
        labels: labels,
        deleteLabelResult: (_) async =>
            throw ItemException('label is in use'),
      );
      await tester.pumpWidget(_harness(service: service));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Delete'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();

      expect(service.deleteCalls, equals(const <int>[3]));
      expect(find.textContaining('label is in use'), findsOneWidget);
      // The label is still present because the delete failed.
      expect(find.text('urgent'), findsOneWidget);
    });
  });
}