import 'dart:math';

/// Лимиты комнат и роли игроков (без бэкенда).
class MultiplayerConfig {
  MultiplayerConfig._();

  /// Сколько человек в комнате по теме (2–4).
  static int maxPlayers(String topicId) {
    switch (topicId) {
      case 'cafe':
        return 2;
      case 'tavern':
      case 'raise':
        return 3;
      case 'tokyo':
      case 'airport':
      case 'custom':
        return 4;
      default:
        return 4;
    }
  }

  static const List<String> _defaultRoles = [
    'Участник A',
    'Участник B',
    'Участник C',
    'Участник D',
  ];

  static List<String> _pool(String topicId) {
    switch (topicId) {
      case 'cafe':
        return [
          'Бариста',
          'Посетитель',
          'Менеджер зала',
          'Уборщица',
          'Кассир',
        ];
      case 'tavern':
        return [
          'Трактирщик',
          'Странник',
          'Музыкант',
          'Охранник',
          'Торговец',
        ];
      case 'raise':
        return [
          'Руководитель',
          'Сотрудник',
          'HR из зала',
          'Наблюдатель совета',
        ];
      case 'tokyo':
        return [
          'Сотрудник станции',
          'Турист',
          'Сотрудник безопасности',
          'Продавец в киоске',
        ];
      case 'airport':
        return [
          'Офицер погранслужбы',
          'Путешественник',
          'Сотрудник авиакомпании',
          'Волонтёр',
        ];
      case 'custom':
        return _pool('cafe');
      default:
        return List<String>.from(_defaultRoles);
    }
  }

  /// Каждому игроку — своя роль (перемешиваем пул, при нехватке дублируем по кругу).
  static Map<String, String> assignRoles(List<String> playerNames, String topicId, Random rng) {
    final names = List<String>.from(playerNames)..shuffle(rng);
    final pool = List<String>.from(_pool(topicId))..shuffle(rng);
    final out = <String, String>{};
    for (var i = 0; i < names.length; i++) {
      out[names[i]] = pool[i % pool.length];
    }
    return out;
  }
}
