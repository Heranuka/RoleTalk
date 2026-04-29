class HistoryEntry {
  HistoryEntry({
    required this.id,
    required this.at,
    required this.kind,
    required this.title,
    required this.subtitle,
    this.meta,
  });

  final String id;
  final DateTime at;
  /// `ai` | `multiplayer`
  final String kind;
  final String title;
  final String subtitle;
  final Map<String, dynamic>? meta;

  Map<String, dynamic> toJson() => {
        'id': id,
        'at': at.toIso8601String(),
        'kind': kind,
        'title': title,
        'subtitle': subtitle,
        'meta': meta,
      };

  static HistoryEntry fromJson(Map<String, dynamic> j) {
    return HistoryEntry(
      id: j['id'] as String,
      at: DateTime.parse(j['at'] as String),
      kind: j['kind'] as String,
      title: j['title'] as String,
      subtitle: j['subtitle'] as String,
      meta: j['meta'] != null ? Map<String, dynamic>.from(j['meta'] as Map) : null,
    );
  }
}
