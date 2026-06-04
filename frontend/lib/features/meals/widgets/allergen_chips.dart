import 'package:flutter/material.dart';

import '../../../data/models.dart';
import '../../../l10n/generated/app_localizations.dart';

class AllergenChips extends StatelessWidget {
  const AllergenChips({
    required this.selected,
    required this.onChanged,
    super.key,
  });
  final Set<String> selected;
  final ValueChanged<Set<String>> onChanged;

  String _labelFor(BuildContext c, String tag) {
    final t = AppLocalizations.of(c);
    return switch (tag) {
      'gluten' => t.tagGluten,
      'lactose' => t.tagLactose,
      'histamine' => t.tagHistamine,
      'fodmap_high' => t.tagFodmapHigh,
      'nuts' => t.tagNuts,
      'egg' => t.tagEgg,
      'soy' => t.tagSoy,
      'fructose' => t.tagFructose,
      'fish' => t.tagFish,
      'shellfish' => t.tagShellfish,
      'sulphites' => t.tagSulphites,
      'sesame' => t.tagSesame,
      'mustard' => t.tagMustard,
      'celery' => t.tagCelery,
      _ => tag,
    };
  }

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 6,
      runSpacing: 6,
      children: [
        for (final tag in allergenTags)
          FilterChip(
            label: Text(_labelFor(context, tag)),
            selected: selected.contains(tag),
            onSelected: (yes) {
              final next = Set<String>.from(selected);
              if (yes) {
                next.add(tag);
              } else {
                next.remove(tag);
              }
              onChanged(next);
            },
          ),
      ],
    );
  }
}
