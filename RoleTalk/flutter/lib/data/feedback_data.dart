class FeedbackData {
  const FeedbackData({
    required this.roleplayScore,
    required this.logicScore,
    required this.empathyScore,
    required this.highlights,
    required this.improvements,
  });

  final int roleplayScore;
  final int logicScore;
  final int empathyScore;
  final List<String> highlights;
  final List<String> improvements;
}

class FeedbackRepository {
  FeedbackRepository._();

  static FeedbackData forTopic(String id) {
    return _map[id] ?? _map['raise']!;
  }

  static final Map<String, FeedbackData> _map = {
    'raise': const FeedbackData(
      roleplayScore: 82,
      logicScore: 90,
      empathyScore: 65,
      highlights: [
        'Отлично сыграно на уверенности: вы четко назвали свои достижения.',
        'Логичная аргументация об оффере конкурентов без прямой угрозы уходом.',
      ],
      improvements: [
        'Не хватило эмпатии к боссу — стоило учесть урезанные бюджеты.',
        'В середине разговора вы начали оправдываться.',
      ],
    ),
    'broken_car': const FeedbackData(
      roleplayScore: 95,
      logicScore: 88,
      empathyScore: 50,
      highlights: [
        'Очень жестко и по фактам: указали на капот и сбили цену.',
        'Идеальная игра роли проницательного покупателя.',
      ],
      improvements: [
        'Можно было быть чуть менее агрессивным, чтобы не сорвать сделку.',
      ],
    ),
    'taxi': const FeedbackData(
      roleplayScore: 78,
      logicScore: 85,
      empathyScore: 80,
      highlights: [
        'Вы отлично сохранили самообладание (эмпатия 80%).',
        'Не стали кричать и спокойно потребовали вернуть маршрут.',
      ],
      improvements: [
        'Вы слишком легко уступили, когда таксист начал давить на жалость.',
        'Забыли упомянуть, что дико опаздываете на самолет (ваш секретный мотив).',
      ],
    ),
    'breakup': const FeedbackData(
      roleplayScore: 85,
      logicScore: 70,
      empathyScore: 90,
      highlights: [
        'Идеальное использование "Я-сообщений".',
        'Очень мягкий и экологичный выход из отношений без лишней боли.',
      ],
      improvements: [
        'Ваш секретный мотив остался нереализованным, но возможно, в данной ситуации это и к лучшему.',
        'Была логическая нестыковка, когда вы начали путаться в причинах.',
      ],
    ),
    'custom': const FeedbackData(
      roleplayScore: 88,
      logicScore: 85,
      empathyScore: 80,
      highlights: [
        'Хорошая импровизация и адаптация под тему.',
      ],
      improvements: [
        'Иногда выходили из роли.',
      ],
    ),
  };
}
