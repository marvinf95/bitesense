import 'package:flutter/widgets.dart';

import '../../l10n/generated/app_localizations.dart';

String symptomLabel(BuildContext context, String type) {
  final t = AppLocalizations.of(context);
  return switch (type) {
    'heartburn' => t.symHeartburn,
    'bloating' => t.symBloating,
    'diarrhea' => t.symDiarrhea,
    'constipation' => t.symConstipation,
    'headache' => t.symHeadache,
    'fatigue' => t.symFatigue,
    'brain_fog' => t.symBrainFog,
    'skin' => t.symSkin,
    'joint_pain' => t.symJointPain,
    'mood' => t.symMood,
    'nausea' => t.symNausea,
    'other' => t.symOther,
    _ => type,
  };
}
