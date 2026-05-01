import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

const _mutedKey = 'muted_players_v1';
const _reportsKey = 'reports_v1';

class MuteReportStore {
  MuteReportStore._();
  static final MuteReportStore instance = MuteReportStore._();
  static const _uuid = Uuid();

  Future<Set<String>> mutedNames() async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_mutedKey);
    if (raw == null || raw.isEmpty) return {};
    try {
      final list = jsonDecode(raw) as List<dynamic>;
      return list.map((e) => e as String).toSet();
    } catch (_) {
      return {};
    }
  }

  Future<void> setMuted(String displayName, bool muted) async {
    final s = await mutedNames();
    if (muted) {
      s.add(displayName);
    } else {
      s.remove(displayName);
    }
    final p = await SharedPreferences.getInstance();
    await p.setString(_mutedKey, jsonEncode(s.toList()));
  }

  Future<bool> isMuted(String displayName) async {
    final s = await mutedNames();
    return s.contains(displayName);
  }

  Future<void> addReport({
    required String scope,
    required String detail,
    String? target,
  }) async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_reportsKey);
    List<Map<String, dynamic>> list = [];
    if (raw != null && raw.isNotEmpty) {
      try {
        list = (jsonDecode(raw) as List<dynamic>).map((e) => Map<String, dynamic>.from(e as Map)).toList();
      } catch (_) {}
    }
    list.insert(0, {
      'id': _uuid.v4(),
      'at': DateTime.now().toIso8601String(),
      'scope': scope,
      'detail': detail,
      'target': target,
    });
    while (list.length > 30) {
      list.removeLast();
    }
    await p.setString(_reportsKey, jsonEncode(list));
  }
}
