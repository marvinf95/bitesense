import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app/router.dart';
import 'app/theme.dart';
import 'core/locale_controller.dart';
import 'l10n/generated/app_localizations.dart';

void main() {
  runApp(const ProviderScope(child: BiteSenseApp()));
}

class BiteSenseApp extends ConsumerWidget {
  const BiteSenseApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    final locale = ref.watch(localeControllerProvider);
    return MaterialApp.router(
      title: 'BiteSense',
      debugShowCheckedModeBanner: false,
      theme: bitesenseTheme(Brightness.light),
      darkTheme: bitesenseTheme(Brightness.dark),
      routerConfig: router,
      locale: locale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
    );
  }
}
