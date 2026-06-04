import 'dart:ui';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'api_client.dart';

class LocaleController extends StateNotifier<Locale?> {
  LocaleController(this._storage) : super(null) {
    _restore();
  }
  final FlutterSecureStorage _storage;

  Future<void> _restore() async {
    final code = await _storage.read(key: 'locale');
    if (code != null && (code == 'en' || code == 'de')) {
      state = Locale(code);
    }
  }

  Future<void> set(Locale? locale) async {
    state = locale;
    if (locale != null) {
      await _storage.write(key: 'locale', value: locale.languageCode);
    } else {
      await _storage.delete(key: 'locale');
    }
  }
}

final localeControllerProvider =
    StateNotifierProvider<LocaleController, Locale?>(
  (ref) => LocaleController(ref.read(secureStorageProvider)),
);
