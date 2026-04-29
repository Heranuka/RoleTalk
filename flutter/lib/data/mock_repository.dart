import 'package:flutter/material.dart';

import '../models/mock_user.dart';
import '../models/topic_vote.dart';
import '../models/voice_room.dart';

class MockRepository {
  MockRepository._();

  static final List<MockUser> users = [
    MockUser(id: 'u1', name: 'Анна К.', initials: 'АК', accentColor: 0xFFE8B4BC),
    MockUser(id: 'u2', name: 'Маркус Л.', initials: 'МЛ', accentColor: 0xFFB8C5F0),
    MockUser(id: 'u3', name: 'Елена В.', initials: 'ЕВ', accentColor: 0xFFC5E8D5),
    MockUser(id: 'u4', name: 'James O.', initials: 'JO', accentColor: 0xFFFFD8A8),
    MockUser(id: 'u5', name: 'София Р.', initials: 'СР', accentColor: 0xFFD4C4F0),
    MockUser(id: 'u6', name: 'Oliver T.', initials: 'OT', accentColor: 0xFFB3E5FC),
    MockUser(id: 'u7', name: 'Илья Н.', initials: 'ИН', accentColor: 0xFFFFCCBC),
    MockUser(id: 'u8', name: 'Mia Chen', initials: 'MC', accentColor: 0xFFC8E6C9),
  ];

  static final List<TopicVote> topics = [
    TopicVote(
      id: 'raise',
      title: 'Переговоры о повышении',
      emoji: '💼',
      level: 'Хард',
      duration: '8–12 мин',
      skill: 'Переговоры',
      goal: 'Договориться о пересмотре зарплаты.',
      votes: 1842,
      voterIds: ['u1', 'u2', 'u4', 'u6', 'u8'],
      publicContext: 'Встреча 1:1. Вы подводите итоги полугодия.',
      myRole: 'Менеджер',
      partnerRole: 'Ваш Босс (Сергей)',
      aiRoleName: 'Сергей (Босс)',
      aiEmoji: '🧑‍💼',
    ),
    TopicVote(
      id: 'broken_car',
      title: 'Покупка б/у авто',
      emoji: '🚗',
      level: 'Медиум',
      duration: '5–8 мин',
      skill: 'Навыки торга',
      goal: 'Сбить цену до \$5000.',
      votes: 1203,
      voterIds: ['u3', 'u5', 'u7', 'u1'],
      publicContext: 'Вы приехали смотреть машину по объявлению.',
      myRole: 'Покупатель',
      partnerRole: 'Продавец (Максим)',
      aiRoleName: 'Максим (Продавец)',
      aiEmoji: '🧔',
    ),
    TopicVote(
      id: 'taxi',
      title: 'Таксист и пассажир',
      emoji: '🚕',
      level: 'Изи',
      duration: '5–10 мин',
      skill: 'Стрессоустойчивость',
      goal: 'Урегулировать конфликт с таксистом.',
      votes: 2156,
      voterIds: ['u2', 'u4', 'u5', 'u6', 'u7', 'u8'],
      publicContext: 'Вы сели в такси, но водитель едет не туда, а счетчик крутится.',
      myRole: 'Пассажир',
      partnerRole: 'Таксист',
      aiRoleName: 'Таксист',
      aiEmoji: '🧑‍✈️',
    ),
    TopicVote(
      id: 'breakup',
      title: 'Расставание',
      emoji: '💔',
      level: 'Хард',
      duration: '6–9 мин',
      skill: 'Эмпатия',
      goal: 'Расстаться максимально экологично.',
      votes: 978,
      voterIds: ['u1', 'u3', 'u8'],
      publicContext: 'Вы позвали партнера в кафе, чтобы сказать, что все кончено.',
      myRole: 'Партнер',
      partnerRole: 'Ваш Партнер',
      aiRoleName: 'Партнер',
      aiEmoji: '😒',
    ),
  ];

  static final List<VoiceRoom> voiceRooms = [
    const VoiceRoom(
      id: 'vr_mock',
      title: 'RoleTalk Lobby',
      subtitle: 'Открытая сессия',
      emoji: '🎭',
      onlineCount: 12,
      levelTag: 'любой',
      accent: Color(0xFF6366F1),
      maxPlayers: 2,
    ),
  ];

  static List<TopicVote> topicsByVotes() {
    final copy = List<TopicVote>.from(topics);
    copy.sort((a, b) => b.votes.compareTo(a.votes));
    return copy;
  }

  static int effectiveVotes(TopicVote t, Map<String, int> userPicks) {
    return t.votes + (userPicks[t.id] ?? 0);
  }

  static List<TopicVote> topicsSortedByEffectiveVotes(Map<String, int> userPicks) {
    final copy = List<TopicVote>.from(topics);
    copy.sort((a, b) {
      final va = effectiveVotes(a, userPicks);
      final vb = effectiveVotes(b, userPicks);
      return vb.compareTo(va);
    });
    return copy;
  }
}
