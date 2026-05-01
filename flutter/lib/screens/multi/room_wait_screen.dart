import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../data/mock_repository.dart';
import '../../models/voice_room.dart';
import '../../services/auth_service.dart';
import '../../services/local_notification_service.dart';
import '../../services/settings_store.dart';
import '../../theme/app_theme.dart';
import 'room_topic_screen.dart';

class _Slot {
  _Slot({required this.name, required this.isBot, this.ready = false});

  final String name;
  final bool isBot;
  bool ready;
}

class RoomWaitScreen extends StatefulWidget {
  const RoomWaitScreen({super.key, required this.room});

  final VoiceRoom room;

  @override
  State<RoomWaitScreen> createState() => _RoomWaitScreenState();
}

class _RoomWaitScreenState extends State<RoomWaitScreen> with SingleTickerProviderStateMixin {
  final _rng = Random();
  late int _max;
  final List<_Slot> _slots = [];
  Timer? _fillTimer;
  Timer? _readyTimer;
  bool _navigated = false;
  bool _full = false;
  late AnimationController _dots;

  @override
  void initState() {
    super.initState();
    _max = widget.room.maxPlayers;
    _dots = AnimationController(vsync: this, duration: const Duration(milliseconds: 1200))..repeat();
    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    _slots.add(_Slot(name: me, isBot: false, ready: false));
    _fillTimer = Timer.periodic(const Duration(milliseconds: 1400), (_) => _tickFill());
  }

  @override
  void dispose() {
    _dots.dispose();
    _fillTimer?.cancel();
    _readyTimer?.cancel();
    super.dispose();
  }

  void _tickFill() {
    if (_slots.length >= _max) {
      _fillTimer?.cancel();
      setState(() => _full = true);
      _startReadyPhase();
      return;
    }
    final pool = MockRepository.users.map((u) => u.name).toList()..shuffle(_rng);
    for (final name in pool) {
      if (!_slots.any((s) => s.name == name)) {
        setState(() => _slots.add(_Slot(name: name, isBot: true, ready: false)));
        break;
      }
    }
    if (_slots.length >= _max) {
      _fillTimer?.cancel();
      setState(() => _full = true);
      _startReadyPhase();
    }
  }

  void _startReadyPhase() {
    _readyTimer?.cancel();
    _readyTimer = Timer.periodic(const Duration(milliseconds: 900), (_) {
      final bots = _slots.where((s) => s.isBot && !s.ready).toList()..shuffle(_rng);
      if (bots.isEmpty) {
        _readyTimer?.cancel();
        _tryProceed();
        return;
      }
      setState(() => bots.first.ready = true);
      _tryProceed();
    });
  }

  void _userReady() {
    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    final i = _slots.indexWhere((s) => !s.isBot && s.name == me);
    if (i < 0) return;
    setState(() => _slots[i].ready = true);
    _tryProceed();
  }

  Future<void> _tryProceed() async {
    if (_navigated || !_full) return;
    if (!_slots.every((s) => s.ready)) return;
    _readyTimer?.cancel();
    _fillTimer?.cancel();
    if (_navigated || !mounted) return;
    _navigated = true;

    if (SettingsStore.instance.vibrateOnReady) {
      await HapticFeedback.heavyImpact();
    }
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Команда готова — ${widget.room.title}'),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
    if (SettingsStore.instance.notifyLobbyReady) {
      await LocalNotificationService.showRoomReady(widget.room.title);
    }

    if (!mounted) return;
    final names = _slots.map((s) => s.name).toList();
    Navigator.of(context).pushReplacement(
      MaterialPageRoute<void>(
        builder: (_) => RoomTopicScreen(
          room: widget.room,
          playerNames: names,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    final allReady = _full && _slots.isNotEmpty && _slots.every((s) => s.ready);
    _Slot? meSlot;
    for (final s in _slots) {
      if (!s.isBot && s.name == me) {
        meSlot = s;
        break;
      }
    }
    final userReady = meSlot?.ready ?? false;
    final r = widget.room;

    return Scaffold(
      backgroundColor: AppTheme.background,
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            pinned: true,
            expandedHeight: 168,
            backgroundColor: r.accent,
            foregroundColor: Colors.white,
            iconTheme: const IconThemeData(color: Colors.white),
            flexibleSpace: FlexibleSpaceBar(
              titlePadding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
              title: Text(
                r.title,
                style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 17, shadows: [Shadow(blurRadius: 8, color: Colors.black26)]),
              ),
              background: Stack(
                fit: StackFit.expand,
                children: [
                  Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          r.accent,
                          Color.lerp(r.accent, Colors.black, 0.22)!,
                        ],
                      ),
                    ),
                  ),
                  Positioned(
                    right: -12,
                    top: 24,
                    child: Text(r.emoji, style: const TextStyle(fontSize: 96, shadows: [Shadow(blurRadius: 16, color: Colors.black12)])),
                  ),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            sliver: SliverToBoxAdapter(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    r.subtitle,
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(color: AppTheme.textSecondary, fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Icon(Icons.people_outline_rounded, size: 18, color: AppTheme.primary.withValues(alpha: 0.9)),
                      const SizedBox(width: 6),
                      Text(
                        '${r.onlineCount} онлайн в похожих комнатах · до $_max в вашей',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppTheme.textSecondary, fontWeight: FontWeight.w600),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: [
                        BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 12, offset: const Offset(0, 4)),
                      ],
                    ),
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                          decoration: BoxDecoration(
                            color: r.accent.withValues(alpha: 0.12),
                            borderRadius: BorderRadius.circular(10),
                          ),
                          child: Text(r.levelTag, style: TextStyle(fontWeight: FontWeight.w800, color: r.accent, fontSize: 12)),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            _full ? 'Нажмите «Готов» — затем введите тему сессии' : 'Подбираем участников в комнату…',
                            style: const TextStyle(fontWeight: FontWeight.w700, height: 1.35),
                          ),
                        ),
                        if (!_full)
                          AnimatedBuilder(
                            animation: _dots,
                            builder: (_, __) {
                              final i = (_dots.value * 3).floor() % 3;
                              return Text(['·  ', '·· ', '···'][i], style: TextStyle(color: r.accent, fontWeight: FontWeight.w900));
                            },
                          ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, i) {
                  final s = _slots[i];
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 10),
                    child: Material(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                      elevation: 0,
                      shadowColor: Colors.transparent,
                      child: ListTile(
                        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                        leading: CircleAvatar(
                          radius: 24,
                          backgroundColor: s.isBot ? const Color(0xFFF3F4F6) : r.accent.withValues(alpha: 0.2),
                          child: Text(
                            s.name.isNotEmpty ? s.name[0].toUpperCase() : '?',
                            style: TextStyle(
                              fontWeight: FontWeight.w900,
                              fontSize: 18,
                              color: s.isBot ? AppTheme.textSecondary : r.accent,
                            ),
                          ),
                        ),
                        title: Text(s.name, style: const TextStyle(fontWeight: FontWeight.w800)),
                        subtitle: s.isBot
                            ? Text('Бот', style: TextStyle(color: AppTheme.textSecondary.withValues(alpha: 0.85), fontSize: 12))
                            : const Text('Вы', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 12)),
                        trailing: Icon(
                          s.ready ? Icons.check_circle_rounded : Icons.pending_outlined,
                          color: s.ready ? AppTheme.primary : AppTheme.textSecondary,
                          size: 26,
                        ),
                      ),
                    ),
                  );
                },
                childCount: _slots.length,
              ),
            ),
          ),
          if (!_full)
            const SliverPadding(
              padding: EdgeInsets.all(32),
              sliver: SliverToBoxAdapter(child: Center(child: CircularProgressIndicator())),
            ),
          if (_full && !allReady)
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
              sliver: SliverToBoxAdapter(
                child: FilledButton(
                  onPressed: userReady ? null : _userReady,
                  style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(52),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  ),
                  child: Text(userReady ? 'Вы готовы — ждём остальных' : 'Готов — задать тему', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
