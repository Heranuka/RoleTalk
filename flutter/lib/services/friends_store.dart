import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

const _key = 'friends_v1';

class Friend {
  Friend({required this.id, required this.name});

  final String id;
  final String name;

  Map<String, dynamic> toJson() => {'id': id, 'name': name};

  static Friend fromJson(Map<String, dynamic> j) {
    return Friend(id: j['id'] as String, name: j['name'] as String);
  }
}

class FriendsStore {
  FriendsStore._();
  static final FriendsStore instance = FriendsStore._();
  static const _uuid = Uuid();

  Future<List<Friend>> load() async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(_key);
    if (raw == null || raw.isEmpty) return [];
    try {
      final list = jsonDecode(raw) as List<dynamic>;
      return list.map((e) => Friend.fromJson(Map<String, dynamic>.from(e as Map))).toList();
    } catch (_) {
      return [];
    }
  }

  Future<void> _save(List<Friend> list) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_key, jsonEncode(list.map((f) => f.toJson()).toList()));
  }

  Future<void> addByName(String name) async {
    final n = name.trim();
    if (n.isEmpty) return;
    final list = await load();
    if (list.any((f) => f.name.toLowerCase() == n.toLowerCase())) return;
    list.add(Friend(id: _uuid.v4(), name: n));
    await _save(list);
  }

  Future<void> remove(String id) async {
    final list = await load()..removeWhere((f) => f.id == id);
    await _save(list);
  }

  /// Демо-ссылка для шаринга (без deep link в приложении).
  String buildInviteLink({required String topicId}) {
    final token = _uuid.v4();
    return 'speaksim://join?topic=$topicId&token=$token';
  }
}
