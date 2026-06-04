import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';

import '../../data/meals_repository.dart';
import '../../data/models.dart';
import '../../l10n/generated/app_localizations.dart';
import 'widgets/allergen_chips.dart';
import 'widgets/barcode_scanner_sheet.dart';

class MealAddScreen extends ConsumerStatefulWidget {
  const MealAddScreen({super.key});
  @override
  ConsumerState<MealAddScreen> createState() => _MealAddScreenState();
}

class _MealAddScreenState extends ConsumerState<MealAddScreen> {
  final _formKey = GlobalKey<FormState>();
  final _title = TextEditingController();
  final _notes = TextEditingController();
  final _items = <_DraftItem>[_DraftItem()];
  DateTime _eatenAt = DateTime.now();
  bool _saving = false;

  @override
  void dispose() {
    _title.dispose();
    _notes.dispose();
    super.dispose();
  }

  Future<void> _pickFromCamera() async {
    final picker = ImagePicker();
    final file =
        await picker.pickImage(source: ImageSource.camera, imageQuality: 85);
    if (file == null) return;
    if (!mounted) return;
    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (_) => const Center(child: CircularProgressIndicator()),
    );
    try {
      final mealId =
          await ref.read(mealsRepositoryProvider).createFromImage(file.path);
      if (!mounted) return;
      Navigator.of(context).pop();
      ref.invalidate(mealsListProvider);
      context.go('/meals/$mealId');
    } catch (_) {
      if (!mounted) return;
      Navigator.of(context).pop();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).errorGeneric)),
      );
    }
  }

  Future<void> _scanBarcode() async {
    final ean = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      builder: (_) => const BarcodeScannerSheet(),
    );
    if (ean == null || !mounted) return;
    try {
      final mealId =
          await ref.read(mealsRepositoryProvider).createFromBarcode(ean);
      if (!mounted) return;
      ref.invalidate(mealsListProvider);
      context.go('/meals/$mealId');
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).errorGeneric)),
      );
    }
  }

  Future<void> _saveText() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);
    try {
      final items = _items
          .where((d) => d.name.text.trim().isNotEmpty)
          .map(
            (d) => MealItem(
              id: '',
              mealId: '',
              name: d.name.text.trim().toLowerCase(),
              displayName: d.name.text.trim(),
              tags: d.tags.toList(),
            ),
          )
          .toList();
      await ref.read(mealsRepositoryProvider).create(
            eatenAt: _eatenAt,
            source: 'text',
            title: _title.text.trim().isEmpty ? null : _title.text.trim(),
            notes: _notes.text.trim().isEmpty ? null : _notes.text.trim(),
            items: items,
          );
      if (!mounted) return;
      ref.invalidate(mealsListProvider);
      context.go('/meals');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(t.mealAddTitle)),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _pickFromCamera,
                    icon: const Icon(Icons.camera_alt),
                    label: Text(t.mealPhotoCapture),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _scanBarcode,
                    icon: const Icon(Icons.qr_code_scanner),
                    label: Text(t.mealBarcodeScan),
                  ),
                ),
              ],
            ),
            const Divider(height: 32),
            TextFormField(
              controller: _title,
              decoration: InputDecoration(labelText: t.mealFieldTitle),
            ),
            const SizedBox(height: 12),
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: Text(t.mealFieldTime),
              subtitle: Text('${_eatenAt.toLocal()}'),
              trailing: const Icon(Icons.edit),
              onTap: () async {
                final picked = await showDatePicker(
                  context: context,
                  initialDate: _eatenAt,
                  firstDate: DateTime(2020),
                  lastDate: DateTime(2100),
                );
                if (picked == null || !context.mounted) return;
                final tm = await showTimePicker(
                  context: context,
                  initialTime: TimeOfDay.fromDateTime(_eatenAt),
                );
                if (!context.mounted) return;
                setState(
                  () => _eatenAt = DateTime(
                    picked.year,
                    picked.month,
                    picked.day,
                    tm?.hour ?? 0,
                    tm?.minute ?? 0,
                  ),
                );
              },
            ),
            const SizedBox(height: 12),
            Text(
              t.mealFieldItems,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            ..._items.map(_buildItemRow),
            TextButton.icon(
              onPressed: () => setState(() => _items.add(_DraftItem())),
              icon: const Icon(Icons.add),
              label: Text(t.mealAddItem),
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _notes,
              decoration: InputDecoration(labelText: t.mealFieldNotes),
              maxLines: 3,
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _saving ? null : _saveText,
              child: Text(t.actionSave),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildItemRow(_DraftItem d) {
    final t = AppLocalizations.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextFormField(
              controller: d.name,
              decoration: const InputDecoration(labelText: 'Item'),
            ),
            const SizedBox(height: 8),
            AllergenChips(
              selected: d.tags,
              onChanged: (next) => setState(() {
                d.tags
                  ..clear()
                  ..addAll(next);
              }),
            ),
            Align(
              alignment: Alignment.centerRight,
              child: IconButton(
                icon: const Icon(Icons.delete_outline),
                onPressed: () => setState(() => _items.remove(d)),
                tooltip: t.actionDelete,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DraftItem {
  final TextEditingController name = TextEditingController();
  final Set<String> tags = <String>{};
}
