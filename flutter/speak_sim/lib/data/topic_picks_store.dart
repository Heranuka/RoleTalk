import 'package:shared_preferences/shared_preferences.dart';

import 'mock_repository.dart';

/// Сколько раз пользователь открыл тему (нажал на карточку) — считается «голосом» за неё.
class TopicPicksStore {
  TopicPicksStore._();

  static const _keyPrefix = 'topic_open_';

  static Future<Map<String, int>> loadPicks() async {
    final prefs = await SharedPreferences.getInstance();
    final out = <String, int>{};
    for (final t in MockRepository.topics) {
      out[t.id] = prefs.getInt('$_keyPrefix${t.id}') ?? 0;
    }
    return out;
  }

  static Future<int> incrementPick(String topicId) async {
    final prefs = await SharedPreferences.getInstance();
    final key = '$_keyPrefix$topicId';
    final next = (prefs.getInt(key) ?? 0) + 1;
    await prefs.setInt(key, next);
    return next;
  }
}
