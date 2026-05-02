import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'api_client.dart';

class PracticeService {
  PracticeService._() : _apiClient = ApiClient(Dio());
  static final PracticeService instance = PracticeService._();

  final ApiClient _apiClient;

  Future<Map<String, dynamic>> startSession(String topicId) async {
    print("!!! SENDING TOPIC_ID: '$topicId'");
    try {
      debugPrint("API: Starting session for topic: $topicId");
      final response = await _apiClient.instance.post('/sessions', data: {
        'topic_id': topicId,
      });
      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        debugPrint("API ERROR: Ты не авторизован (401). Проверь JWT Secret на бэкенде!");
      }
      rethrow;
    }
  }

  Future<Map<String, dynamic>> sendVoiceTurn({
    required String sessionId,
    required String audioPath,
    required String language,
  }) async {
    try {
      if (sessionId.isEmpty) throw Exception("Session ID is empty");

      // Создаем FormData. Ключ 'file' должен совпадать с Go: r.FormFile("file")
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          audioPath,
          filename: 'user_audio.m4a',
        ),
      });

      debugPrint("API: Sending voice to /sessions/$sessionId/voice");

      final response = await _apiClient.instance.post(
        '/sessions/$sessionId/voice',
        data: formData,
        options: Options(
          headers: {
            'X-Practice-Language': language,
            // Токен добавится автоматически интерцептором ApiClient
          },
          // Важно для больших аудиофайлов
          sendTimeout: const Duration(seconds: 30),
          receiveTimeout: const Duration(seconds: 30),
        ),
      );

      return response.data as Map<String, dynamic>;
    } catch (e) {
      debugPrint("API ERROR (sendVoiceTurn): $e");
      rethrow;
    }
  }




  Future<List<dynamic>> getSessionHistory(String sessionId) async {
    try {
      final response = await _apiClient.instance.get('/sessions/$sessionId/history');
      return response.data as List<dynamic>;
    } catch (e) {
      debugPrint("API ERROR (getHistory): $e");
      rethrow;
    }
  }

  Future<void> completeSession(String sessionId) async {
    try {
      await _apiClient.instance.post('/sessions/$sessionId/complete');
    } catch (e) {
      debugPrint("API ERROR (completeSession): $e");
      rethrow;
    }
  }
}