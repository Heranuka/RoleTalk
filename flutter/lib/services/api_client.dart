import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

class ApiClient {
  final Dio _dio;
  // Localhost for iOS/Desktop. For Android Emulator use 10.0.2.2
  static const String baseUrl = "http://localhost:8080/api/v1";

  ApiClient(this._dio) {
    _dio.options.baseUrl = baseUrl;
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        // Traceability: Every request gets a unique ID if not present
        options.headers['X-Request-ID'] = const Uuid().v4();

        // Auth: Attach JWT from SharedPreferences
        final prefs = await SharedPreferences.getInstance();
        final token = prefs.getString('access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        return handler.next(options);
      },
      onError: (DioException e, handler) {
        // Centralized Error Mapping
        if (e.response?.statusCode == 401) {
          // Future: Trigger refresh token logic or logout
        }
        return handler.next(e);
      },
    ));
  }

  Dio get instance => _dio;
}