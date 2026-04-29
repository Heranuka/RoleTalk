import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart';

import '../services/friends_store.dart';
import '../theme/app_theme.dart';

class FriendsScreen extends StatefulWidget {
  const FriendsScreen({super.key});

  @override
  State<FriendsScreen> createState() => _FriendsScreenState();
}

class _FriendsScreenState extends State<FriendsScreen> {
  final _name = TextEditingController();
  List<Friend> _list = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _reload();
  }

  @override
  void dispose() {
    _name.dispose();
    super.dispose();
  }

  Future<void> _reload() async {
    final l = await FriendsStore.instance.load();
    if (mounted) setState(() {
      _list = l;
      _loading = false;
    });
  }

  Future<void> _add() async {
    await FriendsStore.instance.addByName(_name.text);
    _name.clear();
    await _reload();
  }

  Future<void> _shareInvite() async {
    final link = FriendsStore.instance.buildInviteLink(topicId: 'any');
    await Share.share(
      'Заходи в SPEAK/SIM (демо-ссылка):\n$link',
      subject: 'Инвайт SPEAK/SIM',
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Друзья'),
        actions: [
          IconButton(
            tooltip: 'Инвайт',
            onPressed: _shareInvite,
            icon: const Icon(Icons.share_outlined),
          ),
        ],
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _name,
                    decoration: const InputDecoration(
                      labelText: 'Имя друга',
                      border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(12))),
                    ),
                    onSubmitted: (_) => _add(),
                  ),
                ),
                const SizedBox(width: 8),
                FilledButton(onPressed: _add, child: const Text('+')),
              ],
            ),
          ),
          Text(
            'Инвайт — шаринг демо-ссылки (deep link в приложении не подключён).',
            style: TextStyle(fontSize: 12, color: AppTheme.textSecondary),
            textAlign: TextAlign.center,
          ),
          const Divider(),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _list.isEmpty
                    ? Center(
                        child: Text('Пока пусто', style: TextStyle(color: AppTheme.textSecondary)),
                      )
                    : ListView.builder(
                        itemCount: _list.length,
                        itemBuilder: (context, i) {
                          final f = _list[i];
                          return Dismissible(
                            key: ValueKey(f.id),
                            background: Container(color: Colors.red.shade100, alignment: Alignment.centerRight, padding: const EdgeInsets.only(right: 20), child: const Icon(Icons.delete)),
                            onDismissed: (_) async {
                              await FriendsStore.instance.remove(f.id);
                              await _reload();
                            },
                            child: ListTile(
                              title: Text(f.name, style: const TextStyle(fontWeight: FontWeight.w700)),
                              trailing: IconButton(
                                icon: const Icon(Icons.share_outlined, size: 20),
                                onPressed: () => Share.share('Добавь меня в SPEAK/SIM: ${f.name}'),
                              ),
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }
}
