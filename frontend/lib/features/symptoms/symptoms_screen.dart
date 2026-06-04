import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../data/symptoms_repository.dart';
import '../../l10n/generated/app_localizations.dart';
import 'symptom_label.dart';

class SymptomsScreen extends ConsumerWidget {
  const SymptomsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final symptoms = ref.watch(symptomsListProvider);
    final fmt = DateFormat.yMMMEd(Localizations.localeOf(context).toLanguageTag()).add_Hm();

    return Scaffold(
      appBar: AppBar(title: Text(t.navSymptoms)),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.go('/symptoms/add'),
        icon: const Icon(Icons.add),
        label: Text(t.actionAdd),
      ),
      body: symptoms.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(t.errorGeneric)),
        data: (list) {
          if (list.isEmpty) {
            return Center(child: Padding(padding: const EdgeInsets.all(24), child: Text(t.symptomsEmpty, textAlign: TextAlign.center)));
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(symptomsListProvider),
            child: ListView.separated(
              itemCount: list.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (_, i) {
                final s = list[i];
                return ListTile(
                  leading: CircleAvatar(child: Text('${s.severity}')),
                  title: Text(symptomLabel(context, s.type)),
                  subtitle: Text(fmt.format(s.occurredAt.toLocal())),
                  trailing: s.bristolStool == null ? null : Chip(label: Text('Bristol ${s.bristolStool}')),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
