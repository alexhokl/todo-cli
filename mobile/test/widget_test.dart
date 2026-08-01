import 'package:flutter_test/flutter_test.dart';

import 'package:todo/main.dart';

void main() {
  testWidgets('app builds and shows title', (WidgetTester tester) async {
    await tester.pumpWidget(const TodoApp());
    expect(find.text('Todo'), findsWidgets);
  });
}