import 'dart:convert';
import 'dart:math';

import 'package:crypto/crypto.dart';
import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/app_user.dart';

/// Локальная демо-авторизация. Пароли — SHA-256 в prefs (не для продакшена).
/// Google: реальный `google_sign_in`, при ошибке/отмене — мок-профиль.
class AuthService {
  AuthService._();
  static final AuthService instance = AuthService._();

  static const _sessionKey = 'auth_session_v2';
  static String _credKey(String email) => 'cred_v2_${email.toLowerCase().trim()}';

  AppUser? _cache;

  String _hash(String email, String password) {
    final bytes = utf8.encode('${email.toLowerCase().trim()}|$password|speak_sim');
    return sha256.convert(bytes).toString();
  }

  Future<bool> isLoggedIn() async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_sessionKey);
    if (raw == null) {
      _cache = null;
      return false;
    }
    try {
      _cache = AppUser.fromJson(jsonDecode(raw) as Map<String, dynamic>);
      return true;
    } catch (_) {
      return false;
    }
  }

  AppUser? get currentUser => _cache;

  Future<void> _saveSession(AppUser user) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_sessionKey, jsonEncode(user.toJson()));
    _cache = user;
  }

  Future<void> logout() async {
    final p = await SharedPreferences.getInstance();
    await p.remove(_sessionKey);
    _cache = null;
    try {
      await GoogleSignIn.instance.signOut();
    } catch (_) {}
  }

  Future<void> register({
    required String email,
    required String password,
    required String displayName,
  }) async {
    final e = email.trim();
    if (e.length < 5 || !e.contains('@')) {
      throw AuthException('Некорректная почта');
    }
    if (password.length < 6) {
      throw AuthException('Пароль от 6 символов');
    }
    final p = await SharedPreferences.getInstance();
    if (p.containsKey(_credKey(e))) {
      throw AuthException('Эта почта уже зарегистрирована');
    }
    final name = displayName.trim().isEmpty ? e.split('@').first : displayName.trim();
    await p.setString(
      _credKey(e),
      jsonEncode({
        'h': _hash(e, password),
        'name': name,
      }),
    );
    await _saveSession(AppUser(email: e, displayName: name, provider: 'email'));
  }

  Future<void> loginWithEmail({
    required String email,
    required String password,
  }) async {
    final e = email.trim();
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_credKey(e));
    if (raw == null) {
      throw AuthException('Нет пользователя с такой почтой');
    }
    final map = jsonDecode(raw) as Map<String, dynamic>;
    if (map['h'] != _hash(e, password)) {
      throw AuthException('Неверный пароль');
    }
    final name = map['name'] as String? ?? e.split('@').first;
    await _saveSession(AppUser(email: e, displayName: name, provider: 'email'));
  }

  /// `null` — ок. Строка — предупреждение (демо-фолбэк). [AuthException] при отмене.
  Future<String?> signInWithGoogle() async {
    try {
      await GoogleSignIn.instance.initialize();
      final account = await GoogleSignIn.instance.authenticate(scopeHint: ['email', 'profile']);
      final email = account.email;
      final name = account.displayName ?? email.split('@').first;
      await _saveSession(AppUser(email: email, displayName: name, provider: 'google'));
      return null;
    } on AuthException {
      rethrow;
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('GoogleSignIn fallback: $e\n$st');
      }
      final id = Random().nextInt(99999);
      await _saveSession(
        AppUser(
          email: 'demo.google.$id@gmail.com',
          displayName: 'Google (демо)',
          provider: 'google',
        ),
      );
      return 'Google без OAuth-конфига — вошли демо-профилём.';
    }
  }
}

class AuthException implements Exception {
  AuthException(this.message);

  final String message;
}
