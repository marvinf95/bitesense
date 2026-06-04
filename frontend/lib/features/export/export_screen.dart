import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../core/api_client.dart';
import '../../l10n/generated/app_localizations.dart';

class ExportScreen extends ConsumerStatefulWidget {
  const ExportScreen({super.key});
  @override
  ConsumerState<ExportScreen> createState() => _ExportScreenState();
}

class _ExportScreenState extends ConsumerState<ExportScreen> {
  DateTime _from = DateTime.now().subtract(const Duration(days: 14));
  DateTime _to = DateTime.now();
  bool _busy = false;

  Future<DateTime?> _pick(DateTime initial) {
    return showDatePicker(context: context, initialDate: initial, firstDate: DateTime(2020), lastDate: DateTime(2100));
  }

  Future<void> _generate() async {
    setState(() => _busy = true);
    try {
      final dio = ref.read(apiClientProvider);
      final locale = Localizations.localeOf(context).languageCode;
      final resp = await dio.get(
        '/export/pdf',
        queryParameters: {
          'from': _from.toUtc().toIso8601String(),
          'to': _to.toUtc().toIso8601String(),
          'locale': locale,
        },
        options: const _PdfOptions(),
      );
      // On web/mobile the bytes are downloaded; a real implementation hands them off
      // to `printing` package for share/print. Keep MVP simple here.
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('PDF ${resp.data.length} bytes')),
      );
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = AppLocalizations.of(context);
    final df = DateFormat.yMMMd(Localizations.localeOf(context).toLanguageTag());
    return Scaffold(
      appBar: AppBar(title: Text(t.exportTitle)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ListTile(
            title: Text(t.exportRangeFrom),
            subtitle: Text(df.format(_from)),
            trailing: const Icon(Icons.edit),
            onTap: () async {
              final picked = await _pick(_from);
              if (picked != null) setState(() => _from = picked);
            },
          ),
          ListTile(
            title: Text(t.exportRangeTo),
            subtitle: Text(df.format(_to)),
            trailing: const Icon(Icons.edit),
            onTap: () async {
              final picked = await _pick(_to);
              if (picked != null) setState(() => _to = picked);
            },
          ),
          const SizedBox(height: 24),
          FilledButton.icon(
            onPressed: _busy ? null : _generate,
            icon: const Icon(Icons.picture_as_pdf),
            label: Text(t.exportGenerate),
          ),
        ],
      ),
    );
  }
}

class _PdfOptions {
  const _PdfOptions();
}
