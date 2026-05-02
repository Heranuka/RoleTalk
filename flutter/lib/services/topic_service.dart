import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart'; // Для debugPrint
import 'api_client.dart';

class TopicService {
  TopicService._() : _apiClient = ApiClient(Dio());
  static final TopicService instance = TopicService._();

  final ApiClient _apiClient;

  Future<List<dynamic>> getOfficialTopics() async {
    try {
      final response = await _apiClient.instance.get('/topics/official');
      return response.data as List<dynamic>;
    } catch (e) {
      debugPrint("API ERROR (getOfficialTopics): $e");
      return [];
    }
  }

  Future<List<dynamic>> getCommunityTopics() async {
    try {
      final response = await _apiClient.instance.get('/topics/community');
      return response.data as List<dynamic>;
    } catch (e) {
      debugPrint("API ERROR (getCommunityTopics): $e");
      return [];
    }
  }

  Future<Map<String, dynamic>> createTopic({
    required String title,
    required String description,
    required String goal,
    String myRole = "Student",
    String partnerRole = "AI Teacher",
    String partnerEmoji = "🤖",
    String emoji = "🎭",
    String difficultyLevel = "B1",
  }) async {
    try {
      final response = await _apiClient.instance.post('/topics', data: {
        'title': title,
        'description': description,
        'emoji': emoji,
        'difficulty_level': difficultyLevel,
        'my_role': myRole,
        'partner_role': partnerRole,
        'partner_emoji': partnerEmoji,
        'goal': goal,
      });
      return response.data as Map<String, dynamic>;
    } on DioException catch (e) {
      debugPrint("API ERROR: ${e.response?.data}");
      rethrow;
    }
  }

  Future<void> likeTopic(String topicId) async {
    try {
      await _apiClient.instance.post('/topics/$topicId/like');
    } catch (e) {
      debugPrint("API ERROR (likeTopic): $e");
    }
  }

  Future<void> unlikeTopic(String topicId) async {
    try {
      await _apiClient.instance.delete('/topics/$topicId/like');
    } catch (e) {
      debugPrint("API ERROR (unlikeTopic): $e");
    }
  }
}