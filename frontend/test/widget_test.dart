import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:bitesense/l10n/generated/app_localizations.dart';

void main() {
  testWidgets('supports both English and German', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('de'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home:
            Builder(builder: (ctx) => Text(AppLocalizations.of(ctx).navMeals)),
      ),
    );
    expect(find.text('Mahlzeiten'), findsOneWidget);

    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('en'),
        supportedLocales: AppLocalizations.supportedLocales,
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        home:
            Builder(builder: (ctx) => Text(AppLocalizations.of(ctx).navMeals)),
      ),
    );
    expect(find.text('Meals'), findsOneWidget);
  });
}
