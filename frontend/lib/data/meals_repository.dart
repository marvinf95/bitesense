import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/api_client.dart';
import 'models.dart';

class MealsRepository {
  MealsRepository(this._dio);
  final Dio _dio;

  Future<List<Meal>> list({DateTime? from, DateTime? to}) async {
    final resp = await _dio.get('/meals', queryParameters: {
      if (from != null) 'from': from.toUtc().toIso8601String(),
      if (to != null) 'to': to.toUtc().toIso8601String(),
    });
    final list = (resp.data['meals'] as List).cast<Map<String, dynamic>>();
    return list.map(Meal.fromJson).toList();
  }

  Future<Meal> get(String id) async {
    final resp = await _dio.get('/meals/$id');
    return Meal.fromJson(resp.data as Map<String, dynamic>);
  }

  Future<Meal> create({
    required DateTime eatenAt,
    required String source,
    String? title,
    String? notes,
    required List<MealItem> items,
  }) async {
    final resp = await _dio.post('/meals', data: {
      'eaten_at': eatenAt.toUtc().toIso8601String(),
      'source': source,
      if (title != null) 'title': title,
      if (notes != null) 'notes': notes,
      'items': items.map((e) => e.toJson()..remove('id')..remove('meal_id')).toList(),
    });
    return Meal.fromJson(resp.data as Map<String, dynamic>);
  }

  Future<void> delete(String id) async {
    await _dio.delete('/meals/$id');
  }

  Future<String> createFromBarcode(String ean) async {
    final resp = await _dio.post('/meals/from-barcode/$ean');
    return resp.data['meal_id'] as String;
  }

  Future<String> createFromImage(String localFilePath) async {
    final form = FormData.fromMap({
      'photo': await MultipartFile.fromFile(localFilePath),
    });
    final resp = await _dio.post('/meals/from-image', data: form);
    return resp.data['meal_id'] as String;
  }
}

final mealsRepositoryProvider = Provider<MealsRepository>(
  (ref) => MealsRepository(ref.read(apiClientProvider)),
);

final mealsListProvider = FutureProvider<List<Meal>>(
  (ref) => ref.read(mealsRepositoryProvider).list(
        from: DateTime.now().subtract(const Duration(days: 30)),
        to: DateTime.now().add(const Duration(days: 1)),
      ),
);
