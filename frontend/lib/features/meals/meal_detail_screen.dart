import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../data/meals_repository.dart';
import '../../data/models.dart';
import '../../l10n/generated/app_localizations.dart';

final _mealDetailProvider = FutureProvider.family<Meal, String>(
  (ref, id) => ref.read(mealsRepositoryProvider).get(id),
);

class MealDetailScreen extends ConsumerWidget {
  const MealDetailScreen({required this.id, super.key});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = AppLocalizations.of(context);
    final meal = ref.watch(_mealDetailProvider(id));
    final dateFmt =
        DateFormat.yMMMEd(Localizations.localeOf(context).toLanguageTag())
            .add_Hm();

    return Scaffold(
      appBar: AppBar(
        title: Text(t.mealEditTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: () async {
              await ref.read(mealsRepositoryProvider).delete(id);
              if (context.mounted) {
                ref.invalidate(mealsListProvider);
                context.go('/meals');
              }
            },
          ),
        ],
      ),
      body: meal.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text(t.errorGeneric)),
        data: (m) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(m.title ?? '—', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            Text(dateFmt.format(m.eatenAt.toLocal())),
            const SizedBox(height: 16),
            Text(
              t.mealFieldItems,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            for (final it in m.items)
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(it.displayName),
                subtitle: it.tags.isEmpty
                    ? null
                    : Wrap(
                        spacing: 4,
                        children: [
                          for (final t in it.tags) Chip(label: Text(t)),
                        ],
                      ),
                trailing: it.quantityG == null
                    ? null
                    : Text('${it.quantityG!.toStringAsFixed(0)} g'),
              ),
            if (m.notes != null) ...[
              const SizedBox(height: 16),
              Text(
                t.mealFieldNotes,
                style: Theme.of(context).textTheme.titleMedium,
              ),
              Text(m.notes!),
            ],
          ],
        ),
      ),
    );
  }
}
