import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/api_client.dart';
import '../../core/auth_controller.dart';
import '../../core/locale_controller.dart';
import '../../l10n/generated/app_localizations.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final locale = ref.watch(localeControllerProvider);

    return Scaffold(
      appBar: AppBar(title: Text(t.settingsTitle)),
      body: ListView(
        children: [
          ListTile(title: Text(t.settingsLanguage)),
          RadioListTile<Locale?>(
            title: const Text('Auto'),
            value: null,
            groupValue: locale,
            onChanged: (v) =>
                ref.read(localeControllerProvider.notifier).set(v),
          ),
          RadioListTile<Locale?>(
            title: const Text('English'),
            value: const Locale('en'),
            groupValue: locale,
            onChanged: (v) =>
                ref.read(localeControllerProvider.notifier).set(v),
          ),
          RadioListTile<Locale?>(
            title: const Text('Deutsch'),
            value: const Locale('de'),
            groupValue: locale,
            onChanged: (v) =>
                ref.read(localeControllerProvider.notifier).set(v),
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.picture_as_pdf),
            title: Text(t.actionExport),
            onTap: () => context.go('/export'),
          ),
          ListTile(
            leading: const Icon(Icons.logout),
            title: Text(t.settingsLogout),
            onTap: () async {
              await ref.read(authControllerProvider.notifier).logout();
              if (context.mounted) context.go('/login');
            },
          ),
          ListTile(
            leading: const Icon(Icons.delete_forever, color: Colors.red),
            title: Text(
              t.settingsDeleteAccount,
              style: const TextStyle(color: Colors.red),
            ),
            onTap: () => _confirmDelete(context, ref),
          ),
        ],
      ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref) async {
    final t = AppLocalizations.of(context);
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        content: Text(t.settingsDeleteConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(t.actionCancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(t.actionDelete),
          ),
        ],
      ),
    );
    if (ok != true) return;
    final dio = ref.read(apiClientProvider);
    await dio.delete<dynamic>('/auth/account');
    await ref.read(authControllerProvider.notifier).logout();
    if (context.mounted) context.go('/login');
  }
}
