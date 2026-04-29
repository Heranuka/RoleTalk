import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart';

import '../../models/chat_message.dart';
import '../../models/topic_vote.dart';
import '../../services/auth_service.dart';
import '../../services/history_store.dart';
import '../../services/mute_report_store.dart';
import '../../services/voice_capture.dart';
import '../../theme/app_theme.dart';

/// Комната с людьми: у каждого своя роль, свободный диалог, текст + удержание микрофона (голос без «текста»).
class HumanSessionScreen extends StatefulWidget {
  const HumanSessionScreen({
    super.key,
    required this.topic,
    required this.playerRoles,
  });

  final TopicVote topic;
  /// Имя игрока → роль в сцене.
  final Map<String, String> playerRoles;

  @override
  State<HumanSessionScreen> createState() => _HumanSessionScreenState();
}

class _HumanSessionScreenState extends State<HumanSessionScreen> with SingleTickerProviderStateMixin {
  final List<ChatMessage> _messages = [];
  final ScrollController _scroll = ScrollController();
  final TextEditingController _textCtrl = TextEditingController();
  final FocusNode _textFocus = FocusNode();

  bool _recording = false;
  bool _busy = false;
  DateTime? _micPressStart;
  late AnimationController _pulse;
  late VoiceCapture _voice;
  Set<String> _muted = {};

  static final _textPool = [
    "Sounds good from my side.",
    "Let's keep it in character.",
    "One moment — thinking how to say this.",
    "Ha, okay, I can work with that.",
    "Should we switch to a calmer tone?",
  ];

  List<String> get _names => widget.playerRoles.keys.toList();

  @override
  void initState() {
    super.initState();
    _voice = VoiceCapture();
    _pulse = AnimationController(vsync: this, duration: const Duration(milliseconds: 1000))..repeat(reverse: true);
    MuteReportStore.instance.mutedNames().then((s) {
      if (mounted) setState(() => _muted = s);
    });
    WidgetsBinding.instance.addPostFrameCallback((_) => _bootstrap());
  }

  @override
  void dispose() {
    unawaited(_voice.dispose());
    _pulse.dispose();
    _scroll.dispose();
    _textCtrl.dispose();
    _textFocus.dispose();
    super.dispose();
  }

  void _bootstrap() {
    final buf = StringBuffer('🎭 Роли в сцене «${widget.topic.title}» (импровизация, без жёсткой под-темы):\n');
    widget.playerRoles.forEach((n, r) {
      buf.writeln('• $n — $r');
    });
    buf.writeln('\nГоворите свободно: текст или удерживайте 🎤.');
    setState(() {
      _messages.add(ChatMessage(isUser: false, text: buf.toString(), sentAt: DateTime.now()));
    });
    _scrollBottom();
    unawaited(_pushOtherLine());
  }

  void _scrollBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_scroll.hasClients) return;
      _scroll.animateTo(
        _scroll.position.maxScrollExtent,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOutCubic,
      );
    });
  }

  bool _hidden(ChatMessage m) {
    if (m.isUser) return false;
    final n = m.peerName;
    if (n == null) return false;
    return _muted.contains(n);
  }

  Future<void> _pushOtherLine() async {
    if (_messages.length > 18) return;
    setState(() => _busy = true);
    await Future<void>.delayed(Duration(milliseconds: 500 + Random().nextInt(700)));
    if (!mounted) return;

    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    final others = _names.where((n) => n != me).toList();
    if (others.isEmpty) {
      setState(() => _busy = false);
      return;
    }
    var name = others[Random().nextInt(others.length)];
    while (_muted.contains(name) && others.length > 1) {
      others.remove(name);
      name = others[Random().nextInt(others.length)];
    }
    if (_muted.contains(name)) {
      setState(() => _busy = false);
      return;
    }

    final role = widget.playerRoles[name] ?? '';
    final now = DateTime.now();

    if (Random().nextBool()) {
      final sec = 2 + Random().nextInt(18);
      setState(() {
        _messages.add(
          ChatMessage(
            isUser: false,
            sentAt: now,
            voiceSeconds: sec,
            peerName: name,
            peerRole: role,
          ),
        );
        _busy = false;
      });
    } else {
      final line = _textPool[Random().nextInt(_textPool.length)];
      setState(() {
        _messages.add(
          ChatMessage(
            isUser: false,
            text: line,
            sentAt: now,
            peerName: name,
            peerRole: role,
          ),
        );
        _busy = false;
      });
    }
    _scrollBottom();
  }

  Future<void> _sendVoice(double holdSec, {String? localPath}) async {
    if (_busy) return;
    final secs = holdSec.clamp(0.35, 120.0).round().clamp(1, 120);
    setState(() {
      _busy = true;
      _recording = false;
      _messages.add(ChatMessage(isUser: true, sentAt: DateTime.now(), voiceSeconds: secs, voiceLocalPath: localPath));
    });
    if (localPath != null) {
      unawaited(playVoiceFile(localPath));
    }
    _scrollBottom();
    await Future<void>.delayed(const Duration(milliseconds: 350));
    if (!mounted) return;
    setState(() => _busy = false);
    if (Random().nextBool()) await _pushOtherLine();
  }

  void _sendText() {
    final t = _textCtrl.text.trim();
    if (t.isEmpty || _busy) return;
    _textCtrl.clear();
    setState(() {
      _messages.add(ChatMessage(isUser: true, text: t, sentAt: DateTime.now()));
    });
    _scrollBottom();
    if (Random().nextBool()) unawaited(_pushOtherLine());
  }

  void _onMicDown(PointerDownEvent e) {
    if (_busy) return;
    _micPressStart = DateTime.now();
    setState(() => _recording = true);
    unawaited(_voice.start());
  }

  void _onMicUp(PointerEvent e) {
    if (_micPressStart == null) return;
    final sec = DateTime.now().difference(_micPressStart!).inMilliseconds / 1000.0;
    _micPressStart = null;
    setState(() => _recording = false);
    if (sec < 0.22) {
      unawaited(_voice.stop());
      return;
    }
    unawaited(_finishVoiceSend(sec));
  }

  Future<void> _finishVoiceSend(double sec) async {
    final path = await _voice.stop();
    await _sendVoice(sec, localPath: path);
  }

  Future<void> _exit() async {
    final summary = widget.playerRoles.entries.map((e) => '${e.key}: ${e.value}').join('; ');
    await HistoryStore.instance.addMultiplayerSession(
      topicTitle: widget.topic.title,
      subtitle: summary,
      players: widget.playerRoles.length,
    );
    if (!mounted) return;
    Navigator.of(context).popUntil((r) => r.isFirst);
  }

  Future<void> _shareInvite() async {
    final link = 'speaksim://join?topic=${widget.topic.id}&room=demo';
    await Share.share('Комната «${widget.topic.title}»:\n$link', subject: 'SPEAK/SIM');
  }

  Future<void> _openMutePicker() async {
    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    final others = _names.where((n) => n != me).toList();
    if (!mounted) return;
    await showModalBottomSheet<void>(
      context: context,
      builder: (ctx) {
        return SafeArea(
          child: StatefulBuilder(
            builder: (ctx2, setModal) {
              return ListView(
                shrinkWrap: true,
                children: [
                  const Padding(
                    padding: EdgeInsets.all(16),
                    child: Text('Заглушить', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
                  ),
                  for (final n in others)
                    SwitchListTile(
                      title: Text('$n · ${widget.playerRoles[n] ?? ""}'),
                      value: _muted.contains(n),
                      onChanged: (v) async {
                        await MuteReportStore.instance.setMuted(n, v);
                        final s = await MuteReportStore.instance.mutedNames();
                        if (mounted) setState(() => _muted = s);
                        setModal(() {});
                      },
                    ),
                ],
              );
            },
          ),
        );
      },
    );
  }

  Future<void> _reportRoom() async {
    final ctrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Пожаловаться'),
        content: TextField(controller: ctrl, decoration: const InputDecoration(hintText: 'Что не так?'), maxLines: 3),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Отправить')),
        ],
      ),
    );
    if (ok == true && ctrl.text.trim().isNotEmpty) {
      await MuteReportStore.instance.addReport(scope: 'room', detail: ctrl.text.trim(), target: widget.topic.title);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Записано локально (демо)')));
    }
    ctrl.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final me = AuthService.instance.currentUser?.displayName ?? 'Вы';
    final myRole = widget.playerRoles[me] ?? '';

    final visible = _messages.where((m) => !_hidden(m)).toList();

    return Scaffold(
      backgroundColor: AppTheme.chatBg,
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Комната', style: TextStyle(fontSize: 16)),
            Text(widget.topic.title, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w500)),
          ],
        ),
        actions: [
          IconButton(onPressed: _shareInvite, icon: const Icon(Icons.link), tooltip: 'Инвайт'),
          PopupMenuButton<String>(
            onSelected: (v) {
              if (v == 'mute') _openMutePicker();
              if (v == 'report') _reportRoom();
            },
            itemBuilder: (_) => const [
              PopupMenuItem(value: 'mute', child: Text('Заглушить…')),
              PopupMenuItem(value: 'report', child: Text('Пожаловаться')),
            ],
          ),
          TextButton(onPressed: _exit, child: const Text('Выйти')),
        ],
      ),
      body: Column(
        children: [
          if (myRole.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Chip(
                  label: Text('Вы: $myRole', style: const TextStyle(fontWeight: FontWeight.w700)),
                  backgroundColor: AppTheme.primarySoft,
                  side: BorderSide.none,
                ),
              ),
            ),
          Expanded(
            child: ListView.builder(
              controller: _scroll,
              padding: const EdgeInsets.fromLTRB(12, 4, 12, 8),
              itemCount: visible.length,
              itemBuilder: (context, i) {
                final msg = visible[i];
                final time = '${msg.sentAt.hour.toString().padLeft(2, '0')}:${msg.sentAt.minute.toString().padLeft(2, '0')}';
                return _MsgBubble(message: msg, time: time);
              },
            ),
          ),
          _InputPanel(
            recording: _recording,
            busy: _busy,
            pulse: _pulse,
            textController: _textCtrl,
            focusNode: _textFocus,
            onSendText: _sendText,
            onMicDown: _onMicDown,
            onMicUp: _onMicUp,
          ),
        ],
      ),
    );
  }
}

class _MsgBubble extends StatelessWidget {
  const _MsgBubble({required this.message, required this.time});

  final ChatMessage message;
  final String time;

  @override
  Widget build(BuildContext context) {
    final user = message.isUser;
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        mainAxisAlignment: user ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Flexible(
            child: Column(
              crossAxisAlignment: user ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                if (!user && message.peerName != null)
                  Padding(
                    padding: const EdgeInsets.only(left: 4, bottom: 4),
                    child: Text(
                      '${message.peerName} · ${message.peerRole ?? ""}',
                      style: TextStyle(fontSize: 11, fontWeight: FontWeight.w800, color: AppTheme.textSecondary),
                    ),
                  ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    color: user ? AppTheme.bubbleMine : AppTheme.bubblePartner,
                    borderRadius: BorderRadius.only(
                      topLeft: const Radius.circular(18),
                      topRight: const Radius.circular(18),
                      bottomLeft: Radius.circular(user ? 18 : 4),
                      bottomRight: Radius.circular(user ? 4 : 18),
                    ),
                    boxShadow: [
                      BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 4, offset: const Offset(0, 1)),
                    ],
                  ),
                  child: message.isVoice
                      ? Material(
                          color: Colors.transparent,
                          child: InkWell(
                            onTap: user && message.voiceLocalPath != null ? () => playVoiceFile(message.voiceLocalPath!) : null,
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(Icons.graphic_eq_rounded, size: 22, color: user ? AppTheme.bubbleMineText : AppTheme.primary),
                                const SizedBox(width: 8),
                                Text(
                                  '${message.voiceSeconds} с',
                                  style: TextStyle(
                                    fontWeight: FontWeight.w700,
                                    color: user ? AppTheme.bubbleMineText : AppTheme.textPrimary,
                                  ),
                                ),
                                if (user && message.voiceLocalPath != null) ...[
                                  const SizedBox(width: 6),
                                  Icon(Icons.play_circle_fill_rounded, size: 20, color: user ? AppTheme.bubbleMineText.withValues(alpha: 0.9) : AppTheme.primary),
                                ],
                              ],
                            ),
                          ),
                        )
                      : Text(
                          message.text ?? '',
                          style: TextStyle(
                            fontSize: 15,
                            height: 1.35,
                            color: user ? AppTheme.bubbleMineText : AppTheme.textPrimary,
                          ),
                        ),
                ),
                const SizedBox(height: 4),
                Text(time, style: TextStyle(fontSize: 11, color: AppTheme.textSecondary.withValues(alpha: 0.85))),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _InputPanel extends StatelessWidget {
  const _InputPanel({
    required this.recording,
    required this.busy,
    required this.pulse,
    required this.textController,
    required this.focusNode,
    required this.onSendText,
    required this.onMicDown,
    required this.onMicUp,
  });

  final bool recording;
  final bool busy;
  final Animation<double> pulse;
  final TextEditingController textController;
  final FocusNode focusNode;
  final VoidCallback onSendText;
  final void Function(PointerDownEvent) onMicDown;
  final void Function(PointerEvent) onMicUp;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppTheme.surface,
      elevation: 10,
      shadowColor: Colors.black26,
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Expanded(
                    child: TextField(
                      controller: textController,
                      focusNode: focusNode,
                      minLines: 1,
                      maxLines: 4,
                      textInputAction: TextInputAction.send,
                      onSubmitted: (_) => onSendText(),
                      decoration: InputDecoration(
                        hintText: 'Сообщение…',
                        filled: true,
                        fillColor: const Color(0xFFF3F4F6),
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(20), borderSide: BorderSide.none),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  FilledButton(
                    onPressed: busy ? null : onSendText,
                    style: FilledButton.styleFrom(
                      padding: const EdgeInsets.all(14),
                      minimumSize: const Size(48, 48),
                      shape: const CircleBorder(),
                    ),
                    child: const Icon(Icons.send_rounded, size: 20),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      recording
                          ? 'Удерживайте… отпустите = голосовое'
                          : busy
                              ? 'Подождите…'
                              : 'Удерживайте микрофон — только голос, без текста',
                      style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppTheme.textSecondary),
                    ),
                  ),
                  Listener(
                    behavior: HitTestBehavior.opaque,
                    onPointerDown: busy ? null : onMicDown,
                    onPointerUp: busy ? null : onMicUp,
                    onPointerCancel: busy ? null : onMicUp,
                    child: AnimatedBuilder(
                      animation: pulse,
                      builder: (_, __) {
                        final s = recording ? 1.0 + pulse.value * 0.06 : 1.0;
                        return Transform.scale(
                          scale: s,
                          child: Material(
                            color: recording ? const Color(0xFFE53935) : AppTheme.primary,
                            shape: const CircleBorder(),
                            elevation: recording ? 6 : 2,
                            child: SizedBox(
                              width: 50,
                              height: 50,
                              child: Icon(
                                Icons.mic_rounded,
                                color: Colors.white,
                                size: 26,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
