import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:shared_preferences_platform_interface/in_memory_shared_preferences_async.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';
import 'package:todo/l10n/app_localizations.dart';
import 'package:todo/widgets/settings_page.dart';

Widget _harness() {
  return MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: const SettingsPage(),
  );
}

void main() {
  // Install a fresh in-memory SharedPreferencesAsync backend per test so each
  // starts from a clean slate (every key reads as null, which the SettingsPage
  // maps to its declared defaults).
  setUp(() async {
    SharedPreferencesAsyncPlatform.instance =
        InMemorySharedPreferencesAsync.empty();
  });

  group('SettingsPage show labels toggle', () {
    testWidgets('defaults to on and shows the toggle', (tester) async {
      await tester.pumpWidget(_harness());
      await tester.pumpAndSettle();

      // The SwitchListTile for the label-visibility setting is present and on.
      final switchFinder = find.byType(SwitchListTile);
      expect(switchFinder, findsOneWidget);
      expect(
        find.text('Show labels in item list'),
        findsOneWidget,
      );
      expect(
        tester.widget<SwitchListTile>(switchFinder).value,
        isTrue,
      );
    });

    testWidgets('toggling off persists and reflects the change',
        (tester) async {
      await tester.pumpWidget(_harness());
      await tester.pumpAndSettle();

      // Tap the toggle to turn it off. Tapping the row wraps the switch, so
      // pumpAndSettle handles the animated state transition.
      await tester.tap(find.byType(SwitchListTile));
      await tester.pumpAndSettle();

      expect(
        tester.widget<SwitchListTile>(find.byType(SwitchListTile)).value,
        isFalse,
      );

      // The persisted value now reads back as false.
      final prefs = SharedPreferencesAsync();
      expect(
        await prefs.getBool(SettingsPage.showLabelsKey),
        isFalse,
      );
    });

    test('loadShowLabels returns the default when unset', () async {
      // A fresh in-memory store yields the declared default.
      expect(
        await SettingsPage.loadShowLabels(),
        equals(SettingsPage.defaultShowLabels),
      );
    });

    test('loadShowLabels returns the persisted value when set', () async {
      final prefs = SharedPreferencesAsync();
      await prefs.setBool(SettingsPage.showLabelsKey, false);

      expect(await SettingsPage.loadShowLabels(), isFalse);
    });
  });
}