import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/models.dart';
import '../../data/symptoms_repository.dart';
import '../../l10n/generated/app_localizations.dart';
import 'symptom_label.dart';

class SymptomAddScreen extends ConsumerStatefulWidget {
  const SymptomAddScreen({super.key});
  @override
  ConsumerState<SymptomAddScreen> createState() => _SymptomAddScreenState();
}

class _SymptomAddScreenState extends ConsumerState<SymptomAddScreen> {
  String _type = 'bloating';
  double _severity = 5;
  DateTime _occurredAt = DateTime.now();
  int? _durationMin;
  int? _bristol;
  final _notes = TextEditingController();
  bool _saving = false;

  @override
  void dispose() {
    _notes.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      await ref.read(symptomsRepositoryProvider).create(
            occurredAt: _occurredAt,
            type: _type,
            severity: _severity.round(),
            durationMin: _durationMin,
            bristolStool: _bristol,
            notes: _notes.text.trim().isEmpty ? null : _notes.text.trim(),
          );
      if (!mounted) return;
      ref.invalidate(symptomsListProvider);
      context.go('/symptoms');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    final isStool = _type == 'diarrhea' || _type == 'constipation';
    return Scaffold(
      appBar: AppBar(title: Text(t.symptomAddTitle)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          DropdownButtonFormField<String>(
            value: _type,
            decoration: InputDecoration(labelText: t.symptomFieldType),
            items: [
              for (final s in symptomTypes)
                DropdownMenuItem(
                  value: s,
                  child: Text(symptomLabel(context, s)),
                ),
            ],
            onChanged: (v) => setState(() => _type = v ?? 'other'),
          ),
          const SizedBox(height: 16),
          Text('${t.symptomFieldSeverity}: ${_severity.round()}'),
          Slider(
            value: _severity,
            min: 1,
            max: 10,
            divisions: 9,
            label: '${_severity.round()}',
            onChanged: (v) => setState(() => _severity = v),
          ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: Text(t.symptomFieldTime),
            subtitle: Text('${_occurredAt.toLocal()}'),
            trailing: const Icon(Icons.edit),
            onTap: () async {
              final picked = await showDatePicker(
                context: context,
                initialDate: _occurredAt,
                firstDate: DateTime(2020),
                lastDate: DateTime(2100),
              );
              if (picked == null || !context.mounted) return;
              final tm = await showTimePicker(
                context: context,
                initialTime: TimeOfDay.fromDateTime(_occurredAt),
              );
              if (!context.mounted) return;
              setState(
                () => _occurredAt = DateTime(
                  picked.year,
                  picked.month,
                  picked.day,
                  tm?.hour ?? 0,
                  tm?.minute ?? 0,
                ),
              );
            },
          ),
          TextFormField(
            decoration: InputDecoration(labelText: t.symptomFieldDuration),
            keyboardType: TextInputType.number,
            onChanged: (v) => _durationMin = int.tryParse(v),
          ),
          if (isStool) ...[
            const SizedBox(height: 8),
            Text(t.symptomFieldBristol),
            Wrap(
              spacing: 6,
              children: [
                for (var i = 1; i <= 7; i++)
                  ChoiceChip(
                    label: Text('$i'),
                    selected: _bristol == i,
                    onSelected: (_) => setState(() => _bristol = i),
                  ),
              ],
            ),
          ],
          const SizedBox(height: 12),
          TextField(
            controller: _notes,
            decoration: InputDecoration(labelText: t.symptomFieldNotes),
            maxLines: 3,
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: _saving ? null : _save,
            child: Text(t.actionSave),
          ),
        ],
      ),
    );
  }
}
