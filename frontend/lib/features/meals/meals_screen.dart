import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../data/meals_repository.dart';
import '../../l10n/generated/app_localizations.dart';

class MealsScreen extends ConsumerWidget {
  const MealsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final meals = ref.watch(mealsListProvider);
    final dateFmt =
        DateFormat.yMMMEd(Localizations.localeOf(context).toLanguageTag());
    final timeFmt =
        DateFormat.Hm(Localizations.localeOf(context).toLanguageTag());

    return Scaffold(
      appBar: AppBar(title: Text(t.navMeals)),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.go('/meals/add'),
        icon: const Icon(Icons.add),
        label: Text(t.actionAdd),
      ),
      body: meals.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(t.errorGeneric)),
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Text(t.mealsEmpty, textAlign: TextAlign.center),
              ),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(mealsListProvider),
            child: ListView.separated(
              itemCount: list.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (_, i) {
                final m = list[i];
                final subtitle = m.items.map((it) => it.displayName).join(', ');
                return ListTile(
                  leading: const CircleAvatar(child: Icon(Icons.restaurant)),
                  title: Text(
                    m.title ??
                        '${dateFmt.format(m.eatenAt)} · ${timeFmt.format(m.eatenAt)}',
                  ),
                  subtitle: Text(
                    subtitle.isEmpty ? '—' : subtitle,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  trailing: Text(timeFmt.format(m.eatenAt)),
                  onTap: () => context.go('/meals/${m.id}'),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
