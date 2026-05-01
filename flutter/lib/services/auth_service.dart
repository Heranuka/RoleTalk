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

  static const _googleServerClientId =
      'YOUR_WEB_CLIENT_ID.apps.googleusercontent.com';

  final GoogleSignIn _googleSignIn = GoogleSignIn.instance;
  bool _googleSignInInitialized = false;

  AppUser? _cache;

  Future<void> _initGoogleSignIn() async {
    if (_googleSignInInitialized) return;
    await _googleSignIn.initialize(serverClientId: _googleServerClientId);
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
  }

  Future<void> logout() async {
    final p = await SharedPreferences.getInstance();
    await p.remove(_sessionKey);
    await p.remove(_accessTokenKey);
    await p.remove(_refreshTokenKey);
    _cache = null;
    try {
      await _googleSignIn.signOut();
    } catch (_) {}
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
      await loginWithEmail(email: email, password: password);
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Ошибка регистрации';
      throw AuthException(msg.toString());
    }
  }

  Future<void> loginWithEmail({
    required String email,
    required String password,
  }) async {
    try {
      final response = await _apiClient.instance.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

      final data = response.data as Map<String, dynamic>;
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;

      final profileResponse = await _apiClient.instance.get(
        '/users/me',
        options: Options(
          headers: {'Authorization': 'Bearer $accessToken'},
        ),
      );

      final profileData = profileResponse.data as Map<String, dynamic>;

      final user = AppUser(
        email: profileData['email'] as String,
        displayName: profileData['display_name'] as String,
        provider: 'email',
      );

      await _saveSession(user, accessToken, refreshToken);
    } on DioException catch (e) {
      final msg = e.response?.data['message'] ?? 'Ошибка входа';
      throw AuthException(msg.toString());
    }
  }

  Future<String?> signInWithGoogle() async {
    try {
      await _initGoogleSignIn();

      final googleUser = await _googleSignIn.authenticate();

      final auth = await googleUser.authentication;
      final idToken = auth.idToken;

      if (idToken == null) {
        throw AuthException('Could not get ID token from Google');
      }

      if (idToken == null) {
        throw AuthException('Could not get ID token from Google');
      }

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
      );

      await _saveSession(user, accessToken, refreshToken);
      return null;
    } catch (e) {
      if (kDebugMode) print('Google Sign In Error: $e');
      return 'Google OAuth failed. Make sure your client_id is configured in the backend.';
    }
  }
}

class AuthException implements Exception {
  AuthException(this.message);
  final String message;
  @override
  String toString() => message;
}