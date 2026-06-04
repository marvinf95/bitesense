import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/api_client.dart';
import 'models.dart';

class AnalyticsRepository {
  AnalyticsRepository(this._dio);
  final Dio _dio;

  Future<List<CorrelationSuspect>> topSuspects() async {
    final resp = await _dio.get<dynamic>('/analytics/correlations');
    final list = (resp.data['suspects'] as List).cast<Map<String, dynamic>>();
    return list.map(CorrelationSuspect.fromJson).toList();
  }
}

final analyticsRepositoryProvider = Provider<AnalyticsRepository>(
  (ref) => AnalyticsRepository(ref.read(apiClientProvider)),
);

final correlationsProvider = FutureProvider<List<CorrelationSuspect>>(
  (ref) => ref.read(analyticsRepositoryProvider).topSuspects(),
);
