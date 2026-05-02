import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../data/mock_repository.dart';
import '../../models/voice_room.dart';
import '../../services/auth_service.dart';
import '../../services/local_notification_service.dart';
import '../../services/settings_store.dart';
import '../../services/app_localizations.dart';
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
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
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
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
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
          content: Text('${AppLocalizations.of(context, 'lobby_team_ready')} — ${widget.room.title}'),
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
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
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
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: isDark 
              ? [const Color(0xFF1A1A1A), const Color(0xFF121212)]
              : [const Color(0xFFF9FAFB), const Color(0xFFF3F4F6)],
          ),
        ),
        child: CustomScrollView(
          physics: const BouncingScrollPhysics(),
          slivers: [
            SliverAppBar(
              pinned: true,
              expandedHeight: 200,
              backgroundColor: Colors.transparent,
              elevation: 0,
              flexibleSpace: FlexibleSpaceBar(
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
                            Color.lerp(r.accent, Colors.black, 0.3)!,
                          ],
                        ),
                      ),
                    ),
                    Positioned(
                      right: -20,
                      bottom: -20,
                      child: Opacity(
                        opacity: 0.2,
                        child: Text(r.emoji, style: const TextStyle(fontSize: 180)),
                      ),
                    ),
                    Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const SizedBox(height: 40),
                          Text(r.emoji, style: const TextStyle(fontSize: 64)),
                          const SizedBox(height: 8),
                          Text(
                            r.title,
                            style: const TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.w900,
                              fontSize: 24,
                              shadows: [Shadow(blurRadius: 10, color: Colors.black26)],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 24, 20, 16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (r.aiJudgeEnabled)
                      Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        decoration: BoxDecoration(
                          color: Colors.amber.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: Colors.amber.withOpacity(0.3)),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.gavel_rounded, color: Colors.amber, size: 16),
                            const SizedBox(width: 8),
                            Text(
                              '${AppLocalizations.of(context, 'judge_personality')}: ${AppLocalizations.of(context, 'personality_${r.judgePersonality.toLowerCase()}')}',
                              style: const TextStyle(color: Colors.amber, fontWeight: FontWeight.w800, fontSize: 12),
                            ),
                          ],
                        ),
                      ),
                    Text(
                      r.subtitle,
                      style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      AppLocalizations.of(context, 'lobby_waiting'),
                      style: const TextStyle(color: AppTheme.textSecondary, fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
              ),
            ),
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              sliver: SliverList(
                delegate: SliverChildBuilderDelegate(
                  (context, i) {
                    final s = _slots[i];
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Container(
                        decoration: BoxDecoration(
                          color: theme.cardColor,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: AppTheme.premiumShadow,
                        ),
                        child: ListTile(
                          contentPadding: const EdgeInsets.all(12),
                          leading: Container(
                            width: 50,
                            height: 50,
                            decoration: BoxDecoration(
                              gradient: s.isBot 
                                ? null 
                                : AppTheme.primaryGradient,
                              color: s.isBot ? theme.dividerColor.withOpacity(0.1) : null,
                              shape: BoxShape.circle,
                            ),
                            alignment: Alignment.center,
                            child: Text(
                              s.name.isNotEmpty ? s.name[0].toUpperCase() : '?',
                              style: TextStyle(
                                fontWeight: FontWeight.w900,
                                fontSize: 20,
                                color: s.isBot ? AppTheme.textSecondary : Colors.white,
                              ),
                            ),
                          ),
                          title: Text(s.name, style: const TextStyle(fontWeight: FontWeight.w800)),
                          subtitle: Text(
                            s.isBot ? AppLocalizations.of(context, 'lobby_bot') : AppLocalizations.of(context, 'lobby_you'),
                            style: const TextStyle(fontSize: 12, color: AppTheme.textSecondary),
                          ),
                          trailing: AnimatedContainer(
                            duration: const Duration(milliseconds: 300),
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                            decoration: BoxDecoration(
                              color: s.ready ? AppTheme.primary.withOpacity(0.1) : Colors.transparent,
                              borderRadius: BorderRadius.circular(10),
                              border: Border.all(
                                color: s.ready ? AppTheme.primary : theme.dividerColor.withOpacity(0.2),
                              ),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                if (s.ready) const Icon(Icons.check_rounded, color: AppTheme.primary, size: 14),
                                if (s.ready) const SizedBox(width: 4),
                                Text(
                                  s.ready ? AppLocalizations.of(context, 'lobby_status_ready') : AppLocalizations.of(context, 'lobby_status_waiting'),
                                  style: TextStyle(
                                    fontSize: 10,
                                    fontWeight: FontWeight.w900,
                                    color: s.ready ? AppTheme.primary : AppTheme.textSecondary,
                                  ),
                                ),
                              ],
                            ),
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
                padding: EdgeInsets.all(40),
                sliver: SliverToBoxAdapter(child: Center(child: CircularProgressIndicator())),
              ),
            if (_full && !allReady)
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 20, 20, 40),
                sliver: SliverToBoxAdapter(
                  child: Container(
                    decoration: BoxDecoration(
                      gradient: userReady ? null : AppTheme.primaryGradient,
                      color: userReady ? theme.cardColor : null,
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: userReady ? null : [
                        BoxShadow(
                          color: AppTheme.primary.withOpacity(0.3),
                          blurRadius: 12,
                          offset: const Offset(0, 6),
                        ),
                      ],
                    ),
                    child: ElevatedButton(
                      onPressed: userReady ? null : _userReady,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.transparent,
                        shadowColor: Colors.transparent,
                        minimumSize: const Size.fromHeight(56),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      ),
                      child: Text(
                        userReady ? AppLocalizations.of(context, 'lobby_waiting_btn') : AppLocalizations.of(context, 'lobby_ready_btn'),
                        style: TextStyle(
                          color: userReady ? AppTheme.textSecondary : Colors.white,
                          fontWeight: FontWeight.w900,
                          fontSize: 16,
                        ),
                      ),
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
