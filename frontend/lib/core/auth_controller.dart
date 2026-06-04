import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'api_client.dart';

@immutable
class AuthState {
  const AuthState({required this.isAuthenticated, this.userId});
  final bool isAuthenticated;
  final String? userId;
}

class AuthController extends StateNotifier<AuthState> {
  AuthController(this._ref) : super(const AuthState(isAuthenticated: false)) {
    _restore();
  }

  final Ref _ref;
  FlutterSecureStorage get _storage => _ref.read(secureStorageProvider);

  Future<void> _restore() async {
    final access = await _storage.read(key: 'access_token');
    final uid = await _storage.read(key: 'user_id');
    if (access != null && access.isNotEmpty) {
      state = AuthState(isAuthenticated: true, userId: uid);
    }
  }

  Future<void> login(String email, String password) async {
    final dio = _ref.read(apiClientProvider);
    final resp = await dio.post<Map<String, dynamic>>(
      '/auth/login',
      data: {'email': email, 'password': password},
    );
    await _persist(resp.data!);
  }

  Future<void> register(String email, String password, String locale) async {
    final dio = _ref.read(apiClientProvider);
    final resp = await dio.post<Map<String, dynamic>>(
      '/auth/register',
      data: {
        'email': email,
        'password': password,
        'locale': locale,
      },
    );
    await _persist(resp.data!);
  }

  Future<void> logout() async {
    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
    await _storage.delete(key: 'user_id');
    state = const AuthState(isAuthenticated: false);
  }

  Future<void> _persist(Map<String, dynamic> data) async {
    await _storage.write(key: 'access_token', value: data['access_token'] as String);
    await _storage.write(key: 'refresh_token', value: data['refresh_token'] as String);
    await _storage.write(key: 'user_id', value: data['user_id'] as String);
    state = AuthState(isAuthenticated: true, userId: data['user_id'] as String);
  }
}

final authControllerProvider = StateNotifierProvider<AuthController, AuthState>(
  AuthController.new,
);

/// Listenable bridge for GoRouter's `refreshListenable`. Flips a counter every
/// time the auth state transitions between authenticated and unauthenticated,
/// which is the only signal GoRouter needs to re-run its redirect logic.
class AuthListenable extends ValueNotifier<int> {
  AuthListenable(Ref ref) : super(0) {
    ref.listen<AuthState>(authControllerProvider, (prev, next) {
      if (prev?.isAuthenticated != next.isAuthenticated) value++;
    });
  }
}

final authListenableProvider = Provider<AuthListenable>(AuthListenable.new);
