import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/app_user.dart';
import 'api_client.dart';

class AuthService {
  AuthService._() {
    _apiClient = ApiClient(Dio());
  }

  static final AuthService instance = AuthService._();
  late final ApiClient _apiClient;

  static const _sessionKey = 'auth_session_v2';
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';

  // ВАЖНО: Убедись, что этот ID совпадает с твоим Web Client ID из Google Console
  static const _googleServerClientId = 'YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com';

  final GoogleSignIn _googleSignIn = GoogleSignIn.instance;
  bool _googleSignInInitialized = false;

  AppUser? _cache;

  Future<void> _initGoogleSignIn() async {
    if (_googleSignInInitialized) return;
    // На мобилках инициализация может отличаться, но для примера оставляем так
    _googleSignInInitialized = true;
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
      _cache = null;
      return false;
    }
  }

  AppUser? get currentUser => _cache;

  Future<void> _saveSession(
      AppUser user,
      String accessToken,
      String refreshToken,
      ) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_sessionKey, jsonEncode(user.toJson()));
    await p.setString(_accessTokenKey, accessToken);
    await p.setString(_refreshTokenKey, refreshToken);
    _cache = user;
    debugPrint("Session saved. User isVerified: ${user.isVerified}");
  }

  Future<void> logout() async {
    debugPrint("AuthService: Starting logout procedure...");
    try {
      final p = await SharedPreferences.getInstance();
      final refreshToken = p.getString(_refreshTokenKey);

      if (refreshToken != null) {
        // Оповещаем сервер (204 No Content)
        await _apiClient.instance.post('/auth/logout', data: {
          'refresh_token': refreshToken,
        });
        debugPrint("AuthService: Backend logout successful");
      }
    } catch (e) {
      debugPrint("AuthService: Network logout failed (ignored): $e");
    } finally {
      // Очищаем локальные данные в любом случае
      final p = await SharedPreferences.getInstance();
      await p.remove(_sessionKey);
      await p.remove(_accessTokenKey);
      await p.remove(_refreshTokenKey);
      _cache = null;

      try {
        await _googleSignIn.signOut();
      } catch (_) {}

      debugPrint("AuthService: Local data cleared");
    }
  }

  Future<void> register({
    required String email,
    required String password,
    required String displayName,
  }) async {
    try {
      await _apiClient.instance.post('/auth/register', data: {
        'email': email,
        'password': password,
        'display_name': displayName,
      });
      debugPrint("AuthService: Registration success, performing auto-login...");
      await loginWithEmail(email: email, password: password);
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Registration failed';
      throw AuthException(msg.toString());
    }
  }

  Future<void> loginWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      // 1. Авторизация
      final response = await _apiClient.instance.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

      final data = response.data as Map<String, dynamic>;
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;

      // 2. Сразу загружаем профиль, чтобы знать статус верификации
      final profileResponse = await _apiClient.instance.get(
        '/users/me',
        options: Options(
          headers: {'Authorization': 'Bearer $accessToken'},
        ),
      );

      final profileData = profileResponse.data as Map<String, dynamic>;

      final user = AppUser(
        email: profileData['email'],
        displayName: profileData['display_name'],
        provider: 'email',
        // Убедись, что Go возвращает именно этот ключ: is_email_verified
        isVerified: profileData['is_email_verified'] ?? false,
      );

      await _saveSession(user, accessToken, refreshToken);
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Login failed';
      throw AuthException(msg.toString());
    }
  }

  /// Обновляет текущего пользователя данными с сервера.
  /// Используется для проверки статуса подтверждения почты.
  Future<void> refreshProfile() async {
    try {
      debugPrint("AuthService: Refreshing profile from /users/me...");
      final response = await _apiClient.instance.get('/users/me');
      final profileData = response.data as Map<String, dynamic>;

      // Обновляем кэш новыми данными
      _cache = AppUser(
        email: profileData['email'],
        displayName: profileData['display_name'],
        provider: 'email',
        isVerified: profileData['is_email_verified'] ?? false,
      );

      // Сохраняем обновленного юзера в память
      final p = await SharedPreferences.getInstance();
      await p.setString(_sessionKey, jsonEncode(_cache!.toJson()));

      debugPrint("AuthService: Profile updated. New verified status: ${_cache!.isVerified}");
    } on DioException catch (e) {
      debugPrint("AuthService: Failed to refresh profile: ${e.message}");
      // Если 401, значит токен протух - разлогиниваем
      if (e.response?.statusCode == 401) {
        // Здесь можно вызвать локальный logout без запроса к серверу
      }
    } catch (e) {
      debugPrint("AuthService: Unexpected error during refresh: $e");
    }
  }

  Future<void> resendVerification(String email) async {
    try {
      await _apiClient.instance.post('/auth/resend-verification', data: {
        'email': email,
      });
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Failed to send email';
      throw AuthException(msg.toString());
    }
  }

  Future<void> requestPasswordReset(String email) async {
    try {
      await _apiClient.instance.post('/auth/forgot-password', data: {
        'email': email,
      });
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Request failed';
      throw AuthException(msg.toString());
    }
  }

  Future<void> resetPassword(String token, String newPassword) async {
    try {
      await _apiClient.instance.post('/auth/reset-password', data: {
        'token': token,
        'new_password': newPassword,
      });
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Password reset failed';
      throw AuthException(msg.toString());
    }
  }

  Future<void> verifyEmail(String token) async {
    try {
      await _apiClient.instance.post('/auth/verify-email', data: {
        'token': token,
      });
      await refreshProfile(); // Сразу обновляем статус после ручного ввода
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Verification failed';
      throw AuthException(msg.toString());
    }
  }

  Future<String?> signInWithGoogle() async {
    try {
      final googleSignIn = GoogleSignIn.instance;
      await googleSignIn.initialize();

      if (!googleSignIn.supportsAuthenticate()) {
        return "Google Sign-In is not supported on this device";
      }

      final GoogleSignInAccount googleUser = await googleSignIn.authenticate();

      final GoogleSignInAuthentication auth = await googleUser.authentication;
      final String? idToken = auth.idToken;

      if (idToken == null) throw AuthException('No ID Token');

      final response = await _apiClient.instance.post(
        '/auth/google/callback',
        data: {'id_token': idToken},
      );


      final data = response.data as Map<String, dynamic>;
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;

      final user = AppUser(
        email: googleUser.email,
        displayName: googleUser.displayName ?? googleUser.email.split('@').first,
        provider: 'google',
        isVerified: true,
      );

      await _saveSession(user, accessToken, refreshToken);
      return null;
    } catch (e) {
      debugPrint('Google Sign In Error: $e');
      return 'Google OAuth failed: $e';
    }
  }
}

class AuthException implements Exception {
  AuthException(this.message);
  final String message;
  @override
  String toString() => message;
}