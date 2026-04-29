import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:share_plus/share_plus.dart';

import '../models/history_entry.dart';
import '../services/history_store.dart';
import '../theme/app_theme.dart';

class HistoryScreen extends StatefulWidget {
  const HistoryScreen({super.key});

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  List<HistoryEntry> _items = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final l = await HistoryStore.instance.load();
    if (mounted) {
      setState(() {
        _items = l;
        _loading = false;
      });
    }
  }

  Future<void> _export() async {
    final json = await HistoryStore.instance.exportJson();
    await Share.share(json, subject: 'speak_sim_history.json');
  }

  Future<void> _copy() async {
    final json = await HistoryStore.instance.exportJson();
    await Clipboard.setData(ClipboardData(text: json));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('JSON в буфере обмена')));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('История сессий'),
        actions: [
          IconButton(onPressed: _copy, icon: const Icon(Icons.copy_outlined), tooltip: 'Копировать JSON'),
          IconButton(onPressed: _export, icon: const Icon(Icons.ios_share_outlined), tooltip: 'Экспорт'),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _items.isEmpty
              ? Center(child: Text('Пока нет записей', style: TextStyle(color: AppTheme.textSecondary)))
              : RefreshIndicator(
                  onRefresh: _load,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _items.length,
                    itemBuilder: (context, i) {
                      final e = _items[i];
                      return Card(
                        margin: const EdgeInsets.only(bottom: 10),
                        child: ListTile(
                          leading: Icon(
                            e.kind == 'ai' ? Icons.smart_toy_outlined : Icons.groups_outlined,
                            color: AppTheme.primary,
                          ),
                          title: Text(e.title, style: const TextStyle(fontWeight: FontWeight.w800)),
                          subtitle: Text(
                            '${e.kind == 'ai' ? 'ИИ' : 'Люди'} · ${e.subtitle}\n${_fmt(e.at)}',
                            style: TextStyle(height: 1.35, color: AppTheme.textSecondary),
                          ),
                          isThreeLine: true,
                        ),
                      );
                    },
                  ),
                ),
    );
  }

  String _fmt(DateTime t) {
    return '${t.day.toString().padLeft(2, '0')}.${t.month.toString().padLeft(2, '0')}.${t.year} ${t.hour}:${t.minute.toString().padLeft(2, '0')}';
  }
}
