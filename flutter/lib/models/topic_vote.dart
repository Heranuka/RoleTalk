import 'mock_user.dart';

/// Тема сценария (сцена для ролевого общения).
class TopicVote {
  const TopicVote({
    required this.id,
    required this.title,
    required this.emoji,
    required this.level,
    required this.duration,
    required this.skill,
    required this.goal,
    required this.votes,
    required this.voterIds,
    required this.publicContext,
    required this.myRole,
    required this.partnerRole,
    required this.aiRoleName,
    required this.aiEmoji,
    // --- НОВЫЕ ПОЛЯ ДЛЯ РЕЙТИНГА И UGC ---
    this.rating = 0.0,
    this.isOfficial = false,
    this.authorName = 'System',
    this.playsCount = 0,
  });

  final String id;
  final String title;
  final String emoji;
  final String level;
  final String duration;
  final String skill;
  final String goal;
  
  final int votes;
  final List<String> voterIds;

  final String publicContext;
  final String myRole;
  final String partnerRole;

  final String aiRoleName; 
  final String aiEmoji;

  // Новые параметры
  final double rating;       // Например: 4.8
  final bool isOfficial;     // true для ваших тем, false для юзерских
  final String authorName;   // Имя создателя
  final int playsCount;      // Сколько раз сцена была проиграна

  List<MockUser> voters(List<MockUser> all) {
    final map = {for (final u in all) u.id: u};
    return voterIds.map((id) => map[id]!).toList();
  }

  /// Для создания темы пользователем (UGC)
  factory TopicVote.userCreated({
    required String title,
    required String description,
    String author = 'User',
    String emoji = '✍️',
  }) {
    return TopicVote(
      id: 'user_${DateTime.now().millisecondsSinceEpoch}',
      title: title,
      emoji: emoji,
      level: 'любой',
      duration: '5-10 мин',
      skill: 'improv',
      goal: 'Выполните цель, заданную автором.',
      votes: 0,
      voterIds: const [],
      publicContext: description,
      myRole: 'Игрок 1',
      partnerRole: 'Игрок 2',
      aiRoleName: 'Собеседник',
      aiEmoji: '🤖',
      rating: 0.0,
      isOfficial: false,
      authorName: author,
      playsCount: 0,
    );
  }

  /// Пользовательская тема через AI-ввод (тот самый Search Bar)
  factory TopicVote.custom(String rawTitle) {
    final title = rawTitle.trim().isEmpty ? 'Своя тема' : rawTitle.trim();
    return TopicVote(
      id: 'custom_${DateTime.now().millisecondsSinceEpoch}',
      title: title,
      emoji: '✨',
      level: 'любой',
      duration: 'свободно',
      skill: 'диалог',
      goal: 'Вы задали контекст; ИИ создаст сцену.',
      votes: 0,
      voterIds: const [],
      publicContext: 'Сцена генерируется ИИ налету по вашему запросу: "$title".',
      myRole: 'Вы',
      partnerRole: 'Собеседник',
      aiRoleName: 'Собеседник',
      aiEmoji: '🤖',
      rating: 5.0, // AI темы всегда в топе
      isOfficial: true,
      authorName: 'AI Director',
      playsCount: 0,
    );
  }
}