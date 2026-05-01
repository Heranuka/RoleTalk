class ChatMessage {
  ChatMessage({
    required this.isUser,
    required this.sentAt,
    this.text,
    this.voiceSeconds,
    this.voiceLocalPath,
    this.peerName,
    this.peerRole,
    this.isSystem = false,
  }) : assert(
          text != null || voiceSeconds != null,
          'Нужен текст или голос (voiceSeconds)',
        );

  final bool isUser;
  final bool isSystem;
  final DateTime sentAt;

  /// Текст (партнёр ИИ, ваш текст, чужой текст в группе).
  final String? text;

  /// Длительность голосового (сек), без «распознанного текста».
  final int? voiceSeconds;

  /// Локальный файл записи (воспроизведение по нажатию), если доступен.
  final String? voiceLocalPath;

  /// В группе: кто написал (не вы).
  final String? peerName;
  final String? peerRole;

  bool get isVoice => voiceSeconds != null;
}
