import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/analytics_repository.dart';
import '../../l10n/generated/app_localizations.dart';
import '../symptoms/symptom_label.dart';

class AnalyticsScreen extends ConsumerWidget {
  const AnalyticsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final data = ref.watch(correlationsProvider);
    return Scaffold(
      appBar: AppBar(title: Text(t.analyticsTitle)),
      body: data.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(t.errorGeneric)),
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Text(t.analyticsEmpty, textAlign: TextAlign.center),
              ),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(correlationsProvider),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Card(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest,
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Text(
                      t.analyticsDisclaimer,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                for (final s in list)
                  Card(
                    child: ListTile(
                      title: Text(
                        '${s.food} ↔ ${symptomLabel(context, s.symptomType)}',
                      ),
                      subtitle: Text(
                        '${_tierLabel(t, s.tier)} · n=${s.n} · RR=${s.riskRatio.toStringAsFixed(1)} · Δ${s.avgHoursLag.toStringAsFixed(1)}h',
                      ),
                      trailing:
                          Chip(label: Text(s.avgSeverity.toStringAsFixed(1))),
                    ),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }

  String _tierLabel(AppLocalizations t, String tier) => switch (tier) {
        'STRONG_SUSPECT' => t.analyticsTierStrong,
        'SUSPECT' => t.analyticsTierSuspect,
        _ => t.analyticsTierWeak,
      };
}
