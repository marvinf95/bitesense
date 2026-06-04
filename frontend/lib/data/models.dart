// Domain types used across UI/repositories. Manually serialised — no codegen so the
// project builds cleanly without `build_runner` having run yet.

class MealItem {
  MealItem({
    required this.id,
    required this.mealId,
    required this.name,
    required this.displayName,
    this.quantityG,
    this.offId,
    this.confidence,
    this.tags = const [],
  });

  final String id;
  final String mealId;
  final String name;
  final String displayName;
  final double? quantityG;
  final String? offId;
  final double? confidence;
  final List<String> tags;

  factory MealItem.fromJson(Map<String, dynamic> j) => MealItem(
        id: j['id'] as String,
        mealId: j['meal_id'] as String,
        name: j['name'] as String,
        displayName: j['display_name'] as String,
        quantityG: (j['quantity_g'] as num?)?.toDouble(),
        offId: j['off_id'] as String?,
        confidence: (j['confidence'] as num?)?.toDouble(),
        tags: ((j['tags'] as List?) ?? const []).cast<String>(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'meal_id': mealId,
        'name': name,
        'display_name': displayName,
        if (quantityG != null) 'quantity_g': quantityG,
        if (offId != null) 'off_id': offId,
        if (confidence != null) 'confidence': confidence,
        'tags': tags,
      };
}

class Meal {
  Meal({
    required this.id,
    required this.eatenAt,
    required this.source,
    this.title,
    this.notes,
    this.photoPath,
    this.items = const [],
  });

  final String id;
  final DateTime eatenAt;
  final String source; // text|image|barcode|favorite
  final String? title;
  final String? notes;
  final String? photoPath;
  final List<MealItem> items;

  factory Meal.fromJson(Map<String, dynamic> j) => Meal(
        id: j['id'] as String,
        eatenAt: DateTime.parse(j['eaten_at'] as String),
        source: j['source'] as String,
        title: j['title'] as String?,
        notes: j['notes'] as String?,
        photoPath: j['photo_path'] as String?,
        items: ((j['items'] as List?) ?? const [])
            .map((e) => MealItem.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}

class Symptom {
  Symptom({
    required this.id,
    required this.occurredAt,
    required this.type,
    required this.severity,
    this.durationMin,
    this.bristolStool,
    this.notes,
  });

  final String id;
  final DateTime occurredAt;
  final String type;
  final int severity;
  final int? durationMin;
  final int? bristolStool;
  final String? notes;

  factory Symptom.fromJson(Map<String, dynamic> j) => Symptom(
        id: j['id'] as String,
        occurredAt: DateTime.parse(j['occurred_at'] as String),
        type: j['type'] as String,
        severity: j['severity'] as int,
        durationMin: j['duration_min'] as int?,
        bristolStool: j['bristol_stool'] as int?,
        notes: j['notes'] as String?,
      );
}

class CorrelationSuspect {
  CorrelationSuspect({
    required this.food,
    required this.symptomType,
    required this.riskRatio,
    required this.pValue,
    required this.n,
    required this.avgHoursLag,
    required this.avgSeverity,
    required this.tier,
  });

  final String food;
  final String symptomType;
  final double riskRatio;
  final double pValue;
  final int n;
  final double avgHoursLag;
  final double avgSeverity;
  final String tier; // STRONG_SUSPECT|SUSPECT|WEAK_SIGNAL

  factory CorrelationSuspect.fromJson(Map<String, dynamic> j) => CorrelationSuspect(
        food: j['food'] as String,
        symptomType: j['symptom_type'] as String,
        riskRatio: (j['risk_ratio'] as num).toDouble(),
        pValue: (j['p_value'] as num).toDouble(),
        n: j['n'] as int,
        avgHoursLag: (j['avg_hours_lag'] as num).toDouble(),
        avgSeverity: (j['avg_severity'] as num).toDouble(),
        tier: j['tier'] as String,
      );
}

/// Canonical allergen tags shipped with the UI (must match backend `AllTags`).
const allergenTags = <String>[
  'gluten',
  'lactose',
  'histamine',
  'fodmap_high',
  'nuts',
  'egg',
  'soy',
  'fructose',
  'fish',
  'shellfish',
  'sulphites',
  'sesame',
  'mustard',
  'celery',
];

const symptomTypes = <String>[
  'heartburn',
  'bloating',
  'diarrhea',
  'constipation',
  'headache',
  'fatigue',
  'brain_fog',
  'skin',
  'joint_pain',
  'mood',
  'nausea',
  'other',
];
