import 'package:flutter/material.dart';

/// Комната в стиле языковых приложений (лобби без фиксированной «темы урока»).
class VoiceRoom {
  const VoiceRoom({
    required this.id,
    required this.title,
    required this.subtitle,
    required this.emoji,
    required this.onlineCount,
    required this.levelTag,
    required this.accent,
    required this.maxPlayers,
  });

  final String id;
  final String title;
  final String subtitle;
  final String emoji;
  final int onlineCount;
  final String levelTag;
  final Color accent;

  /// Сколько человек набирается в лобби (2–4).
  final int maxPlayers;
}
