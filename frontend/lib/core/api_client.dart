import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const _backendBaseUrl = String.fromEnvironment(
  'BITESENSE_API',
  defaultValue: 'http://localhost:8080',
);

final secureStorageProvider = Provider<FlutterSecureStorage>(
  (_) => const FlutterSecureStorage(),
);

final apiClientProvider = Provider<Dio>((ref) {
  final storage = ref.watch(secureStorageProvider);
  final dio = Dio(
    BaseOptions(
      baseUrl: '$_backendBaseUrl/api/v1',
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      headers: {'Accept': 'application/json'},
    ),
  );
  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await storage.read(key: 'access_token');
        if (token != null && token.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (e, handler) async {
        // Refresh-on-401: rotates refresh token and retries the request once.
        if (e.response?.statusCode == 401) {
          final refreshed = await _tryRefresh(dio, storage);
          if (refreshed) {
            final req = e.requestOptions;
            final token = await storage.read(key: 'access_token');
            req.headers['Authorization'] = 'Bearer $token';
            try {
              final response = await dio.fetch<dynamic>(req);
              return handler.resolve(response);
            } catch (_) {
              return handler.next(e);
            }
          }
        }
        handler.next(e);
      },
    ),
  );
  return dio;
});

Future<bool> _tryRefresh(Dio dio, FlutterSecureStorage storage) async {
  final refresh = await storage.read(key: 'refresh_token');
  if (refresh == null || refresh.isEmpty) return false;
  try {
    final resp =
        await Dio(BaseOptions(baseUrl: dio.options.baseUrl)).post<dynamic>(
      '/auth/refresh',
      data: {'refresh_token': refresh},
    );
    await storage.write(
      key: 'access_token',
      value: resp.data['access_token'] as String,
    );
    await storage.write(
      key: 'refresh_token',
      value: resp.data['refresh_token'] as String,
    );
    return true;
  } catch (_) {
    await storage.delete(key: 'access_token');
    await storage.delete(key: 'refresh_token');
    return false;
  }
}
