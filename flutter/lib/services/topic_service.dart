import 'package:dio/dio.dart';
import 'api_client.dart';

class TopicService {
  TopicService._() : _apiClient = ApiClient(Dio());
  static final TopicService instance = TopicService._();

  final ApiClient _apiClient;

  Future<List<dynamic>> getOfficialTopics() async {
    final response = await _apiClient.instance.get('/topics/official');
    return response.data as List<dynamic>;
  }

  Future<List<dynamic>> getCommunityTopics() async {
    final response = await _apiClient.instance.get('/topics/community');
    return response.data as List<dynamic>;
  }

  Future<Map<String, dynamic>> createTopic({
    required String title,
    required String description,
    required String prompt,
  }) async {
    final response = await _apiClient.instance.post('/topics', data: {
      'title': title,
      'description': description,
      'prompt': prompt,
    });
    return response.data as Map<String, dynamic>;
  }

  Future<void> likeTopic(String topicId) async {
    await _apiClient.instance.post('/topics/$topicId/like');
  }

  Future<void> unlikeTopic(String topicId) async {
    await _apiClient.instance.delete('/topics/$topicId/like');
  }
}
