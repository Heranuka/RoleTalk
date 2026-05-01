import 'package:dio/dio.dart';
import 'api_client.dart';

class PracticeService {
  PracticeService._() : _apiClient = ApiClient(Dio());
  static final PracticeService instance = PracticeService._();

  final ApiClient _apiClient;

  Future<Map<String, dynamic>> startSession(String topicId) async {
    final response = await _apiClient.instance.post('/sessions', data: {
      'topic_id': topicId,
    });
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> sendVoiceTurn({
    required String sessionId,
    required String audioPath,
    required String language,
  }) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(audioPath, filename: 'user_turn.m4a'),
    });

    final response = await _apiClient.instance.post(
      '/sessions/$sessionId/voice',
      data: formData,
      options: Options(headers: {
        'X-Practice-Language': language,
      }),
    );
    return response.data as Map<String, dynamic>; // Should contain AI Text + AI Audio URL
  }

  Future<List<dynamic>> getSessionHistory(String sessionId) async {
    final response = await _apiClient.instance.get('/sessions/$sessionId/history');
    return response.data as List<dynamic>;
  }

  Future<void> completeSession(String sessionId) async {
    await _apiClient.instance.post('/sessions/$sessionId/complete');
  }
}