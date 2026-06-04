import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/api_client.dart';
import 'models.dart';

class SymptomsRepository {
  SymptomsRepository(this._dio);
  final Dio _dio;

  Future<List<Symptom>> list({
    DateTime? from,
    DateTime? to,
    String? type,
  }) async {
    final resp = await _dio.get<dynamic>(
      '/symptoms',
      queryParameters: {
        if (from != null) 'from': from.toUtc().toIso8601String(),
        if (to != null) 'to': to.toUtc().toIso8601String(),
        if (type != null) 'type': type,
      },
    );
    final list = (resp.data['symptoms'] as List).cast<Map<String, dynamic>>();
    return list.map(Symptom.fromJson).toList();
  }

  Future<Symptom> create({
    required DateTime occurredAt,
    required String type,
    required int severity,
    int? durationMin,
    int? bristolStool,
    String? notes,
  }) async {
    final resp = await _dio.post<dynamic>(
      '/symptoms',
      data: {
        'occurred_at': occurredAt.toUtc().toIso8601String(),
        'type': type,
        'severity': severity,
        if (durationMin != null) 'duration_min': durationMin,
        if (bristolStool != null) 'bristol_stool': bristolStool,
        if (notes != null) 'notes': notes,
      },
    );
    return Symptom.fromJson(resp.data as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete<dynamic>('/symptoms/$id');
  }
}

final symptomsRepositoryProvider = Provider<SymptomsRepository>(
  (ref) => SymptomsRepository(ref.read(apiClientProvider)),
);

final symptomsListProvider = FutureProvider<List<Symptom>>(
  (ref) => ref.read(symptomsRepositoryProvider).list(
        from: DateTime.now().subtract(const Duration(days: 30)),
        to: DateTime.now().add(const Duration(days: 1)),
      ),
);
