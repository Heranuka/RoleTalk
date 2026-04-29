import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

import '../models/history_entry.dart';

const _key = 'session_history_v1';
const _max = 50;

class HistoryStore {
  HistoryStore._();
  static final HistoryStore instance = HistoryStore._();
  static const _uuid = Uuid();

  Future<List<HistoryEntry>> load() async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_key);
    if (raw == null || raw.isEmpty) return [];
    try {
      final list = jsonDecode(raw) as List<dynamic>;
      return list.map((e) => HistoryEntry.fromJson(Map<String, dynamic>.from(e as Map))).toList();
    } catch (_) {
      return [];
    }
  }

  Future<void> add(HistoryEntry e) async {
    final all = await load();
    all.insert(0, e);
    while (all.length > _max) {
      all.removeLast();
    }
    final p = await SharedPreferences.getInstance();
    await p.setString(_key, jsonEncode(all.map((x) => x.toJson()).toList()));
  }

  Future<void> addAiSession({required String topicTitle, String detail = 'Сессия с ИИ'}) async {
    await add(
      HistoryEntry(
        id: _uuid.v4(),
        at: DateTime.now(),
        kind: 'ai',
        title: topicTitle,
        subtitle: detail,
      ),
    );
  }

  Future<void> addMultiplayerSession({
    required String topicTitle,
    String subtitle = 'Роли · свободный диалог',
    int players = 0,
  }) async {
    await add(
      HistoryEntry(
        id: _uuid.v4(),
        at: DateTime.now(),
        kind: 'multiplayer',
        title: topicTitle,
        subtitle: subtitle,
        meta: {'players': players},
      ),
    );
  }

  Future<String> exportJson() async {
    final all = await load();
    return const JsonEncoder.withIndent('  ').convert(all.map((e) => e.toJson()).toList());
  }
}
