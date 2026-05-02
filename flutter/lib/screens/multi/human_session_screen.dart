import 'dart:async';

import 'package:flutter/material.dart';

import '../../models/chat_message.dart';
import '../../models/topic_vote.dart';
import '../../models/voice_room.dart';
import '../../services/auth_service.dart';
import '../../services/voice_capture.dart';
import '../../services/app_localizations.dart';
import '../../services/active_session_manager.dart';
import '../../theme/app_theme.dart';
import '../../widgets/session_widgets.dart';

class HumanSessionScreen extends StatefulWidget {
  const HumanSessionScreen({
    super.key,
    required this.topic,
    required this.playerRoles,
    this.room,
  });

  final TopicVote topic;
  final VoiceRoom? room;
  final Map<String, String> playerRoles;

  @override
  State<HumanSessionScreen> createState() => _HumanSessionScreenState();
}

class _HumanSessionScreenState extends State<HumanSessionScreen> with TickerProviderStateMixin {
  final ScrollController _scroll = ScrollController();
  final TextEditingController _textCtrl = TextEditingController();
  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  bool _recording = false;
  final bool _busy = false;
  bool _isTyping = false;
  DateTime? _micPressStart;
  
  late VoiceCapture _voice;
  Timer? _judgeTimer;
  String? _currentHint;

  final Set<String> _kickedPlayers = {};
  final Map<String, AnimationController> _pulseControllers = {};
  AnimationController? _handAnimController;

  static final _mockUsers = ['Anna', 'David', 'Sophie', 'Mark', 'Elena', 'Kevin', 'Lucas', 'Mila', 'Tanya', 'Igor'];
  static const int _maxSeats = 8;
  
  List<String> get _speakers => ActiveSessionManager.instance.speakers.where((n) => !_kickedPlayers.contains(n)).toList();
  List<String> get _audience => ActiveSessionManager.instance.audience.where((n) => !_kickedPlayers.contains(n)).toList();

  @override
  void initState() {
    super.initState();
    _voice = VoiceCapture();
    _handAnimController = AnimationController(vsync: this, duration: const Duration(milliseconds: 1000));
    
    final me = AuthService.instance.currentUser?.displayName ?? 'User';

    for (final name in [...widget.playerRoles.keys, ..._mockUsers, me]) {
      _pulseControllers[name] = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));
    }
    _pulseControllers['AI Judge'] = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));

    if (!ActiveSessionManager.instance.isActive) {
      if (widget.room != null) {
        ActiveSessionManager.instance.startRoom(widget.room!);
      } else {
        ActiveSessionManager.instance.startTopic(widget.topic);
      }
      
      ActiveSessionManager.instance.moveToStage(me);
      for (var i = 0; i < 3; i++) {
        ActiveSessionManager.instance.moveToStage(_mockUsers[i]);
      }
      for (var i = 3; i < _mockUsers.length; i++) {
        ActiveSessionManager.instance.joinAudience(_mockUsers[i]);
      }

      _bootstrap();
      _injectInitialChat();
    }

    if (widget.room?.aiJudgeEnabled ?? false) {
      _judgeTimer = Timer.periodic(const Duration(seconds: 40), (_) => _triggerJudgeComment());
    }

    _textCtrl.addListener(() {
      final typing = _textCtrl.text.isNotEmpty;
      if (typing != _isTyping) setState(() => _isTyping = typing);
    });

    ActiveSessionManager.instance.addListener(_onSessionUpdate);
  }

  void _injectInitialChat() {
    final history = [
      {'n': 'Anna', 't': 'Hello everyone! Glad to be here.'},
      {'n': 'David', 't': 'Welcome to the English Chit-Chat!'},
      {'n': 'Sophie', 't': 'Happy New Year !'},
    ];
    for (final m in history) {
      ActiveSessionManager.instance.addMessage(ChatMessage(isUser: false, text: m['t']!, sentAt: DateTime.now().subtract(const Duration(minutes: 5)), peerName: m['n']!));
    }
  }

  @override
  void dispose() {
    ActiveSessionManager.instance.removeListener(_onSessionUpdate);
    unawaited(_voice.dispose());
    for (final controller in _pulseControllers.values) {
      controller.dispose();
    }
    _handAnimController?.dispose();
    _scroll.dispose();
    _textCtrl.dispose();
    _judgeTimer?.cancel();
    super.dispose();
  }

  void _onSessionUpdate() {
    if (mounted) setState(() {});
    _scrollBottom();
  }

  void _triggerJudgeComment() {
    if (!mounted || _busy) return;
    ActiveSessionManager.instance.addMessage(ChatMessage(isUser: false, text: AppLocalizations.of(context, 'judge_comment'), sentAt: DateTime.now(), peerName: 'AI Judge', peerRole: 'Judge'));
    _pulseControllers['AI Judge']?.repeat(reverse: true);
    Future.delayed(const Duration(seconds: 4), () => _pulseControllers['AI Judge']?.stop());
  }

  void _bootstrap() {
    ActiveSessionManager.instance.addMessage(ChatMessage(isUser: false, text: AppLocalizations.of(context, 'notice_welcome'), sentAt: DateTime.now()));
  }

  void _scrollBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_scroll.hasClients) return;
      _scroll.animateTo(_scroll.position.maxScrollExtent, duration: const Duration(milliseconds: 300), curve: Curves.easeOutCubic);
    });
  }

  void _onMicDown(PointerDownEvent e) {
    if (_busy) return;
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
    if (!ActiveSessionManager.instance.speakers.contains(me)) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(AppLocalizations.of(context, 'raise_hand_hint'))));
      return;
    }
    _pulseControllers[me]?.repeat(reverse: true);
    _micPressStart = DateTime.now();
    setState(() => _recording = true);
    unawaited(_voice.start());
  }

  void _onMicUp(PointerEvent e) {
    if (_micPressStart == null) return;
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
    _pulseControllers[me]?.stop();
    _pulseControllers[me]?.value = 0.0;
    final sec = DateTime.now().difference(_micPressStart!).inMilliseconds / 1000.0;
    _micPressStart = null;
    setState(() => _recording = false);
    if (sec < 0.22) { unawaited(_voice.stop()); return; }
    unawaited(_finishVoiceSend(sec));
  }

  Future<void> _finishVoiceSend(double sec) async {
    final path = await _voice.stop();
    ActiveSessionManager.instance.addMessage(ChatMessage(isUser: true, sentAt: DateTime.now(), voiceSeconds: sec.round(), voiceLocalPath: path));
  }

  void _sendText() {
    final t = _textCtrl.text.trim();
    if (t.isEmpty) return;
    _textCtrl.clear();
    ActiveSessionManager.instance.addMessage(ChatMessage(isUser: true, text: t, sentAt: DateTime.now()));
  }

  @override
  Widget build(BuildContext context) {
    final me = AuthService.instance.currentUser?.displayName ?? 'User';
    final isAdmin = (widget.playerRoles[me]?.toLowerCase() == 'host' || me == 'User'); 
    final isSpeaker = ActiveSessionManager.instance.speakers.contains(me);
    final messages = ActiveSessionManager.instance.messages.toList();

    return Scaffold(
      key: _scaffoldKey,
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [Color(0xFF0D1B2A), Color(0xFF1B263B)],
          ),
        ),
        child: Stack(
          children: [
            Opacity(
              opacity: 0.05,
              child: Image.network(
                'https://www.transparenttextures.com/patterns/cubes.png',
                repeat: ImageRepeat.repeat, width: double.infinity, height: double.infinity,
              ),
            ),

            Column(
              children: [
                AppBar(
                  backgroundColor: Colors.transparent,
                  elevation: 0,
                  leading: IconButton(icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 36, color: Colors.white), onPressed: () => Navigator.pop(context)),
                  title: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(widget.topic.title, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w800, color: Colors.white, letterSpacing: 0.5)),
                      Row(
                        children: [
                          Container(width: 6, height: 6, decoration: const BoxDecoration(color: Colors.greenAccent, shape: BoxShape.circle)),
                          const SizedBox(width: 4),
                          Text('${_speakers.length + _audience.length} ${AppLocalizations.of(context, 'online_count')}', style: const TextStyle(fontSize: 11, color: Colors.white60, fontWeight: FontWeight.w600)),
                        ],
                      ),
                    ],
                  ),
                  actions: [
                    IconButton(icon: const Icon(Icons.share_rounded, color: Colors.white, size: 22), onPressed: () {}),
                    IconButton(icon: const Icon(Icons.more_horiz_rounded, color: Colors.white, size: 24), onPressed: () => _scaffoldKey.currentState?.openEndDrawer()),
                  ],
                ),

                // COMPACT SPEAKER GRID
                GridView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 4, crossAxisSpacing: 2, mainAxisSpacing: 0, childAspectRatio: 0.9,
                  ),
                  itemCount: _maxSeats,
                  itemBuilder: (context, i) {
                    if (i < _speakers.length) {
                      final name = _speakers[i];
                      return _HTSpeakerTile(name: name == me ? AppLocalizations.of(context, 'session_you') : name, emoji: name == me ? '👤' : (i % 2 == 0 ? '👨' : '👩'), animation: _pulseControllers[name]);
                    }
                    return const _HTEmptySeat();
                  },
                ),

                // AUDIENCE ROW (Immediate)
                Container(
                  height: 60,
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    itemCount: _audience.length,
                    itemBuilder: (context, i) => Padding(
                      padding: const EdgeInsets.only(right: 12),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          CircleAvatar(radius: 18, backgroundColor: Colors.white10, child: Text(_audience[i][0], style: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.bold))),
                          const SizedBox(height: 2),
                          Text(_audience[i], style: const TextStyle(fontSize: 8, color: Colors.white54, fontWeight: FontWeight.w500)),
                        ],
                      ),
                    ),
                  ),
                ),

                const Spacer(),
                const SizedBox(height: 110),
              ],
            ),

            // CHAT OVERLAY
            Positioned(
              bottom: 110, left: 16, width: MediaQuery.of(context).size.width * 0.8,
              height: 240,
              child: ShaderMask(
                shaderCallback: (rect) => const LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [Colors.transparent, Colors.black, Colors.black]).createShader(rect),
                blendMode: BlendMode.dstIn,
                child: ListView.builder(
                  controller: _scroll,
                  padding: const EdgeInsets.only(top: 100),
                  itemCount: messages.length,
                  itemBuilder: (context, i) {
                    final m = messages[i];
                    if (m.isVoice) return const SizedBox.shrink();
                    final isNotice = m.text?.startsWith('Notice:') == true ||
                        m.text?.startsWith('Уведомление:') == true;
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (!isNotice) CircleAvatar(radius: 14, backgroundColor: Colors.white12, child: Text(m.peerName?[0] ?? 'Y', style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold))),
                          const SizedBox(width: 8),
                          Flexible(
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                              decoration: BoxDecoration(color: isNotice ? Colors.deepPurple.withOpacity(0.3) : Colors.black38, borderRadius: BorderRadius.circular(18)),
                              child: RichText(
                                text: TextSpan(
                                  children: [
                                    if (!isNotice) TextSpan(text: '${m.peerName ?? AppLocalizations.of(context, 'session_you')}: ', style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 12, color: Colors.lightBlueAccent)),
                                    TextSpan(text: m.text ?? "", style: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.w400)),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),
            ),

            // REFINED FOOTER
            Positioned(
              bottom: 0, left: 0, right: 0,
              child: _RefinedHTFooter(
                isSpeaker: isSpeaker, recording: _recording, isTyping: _isTyping,
                textController: _textCtrl,
                handAnimation: _handAnimController,
                onSendText: _sendText, onMicDown: _onMicDown, onMicUp: _onMicUp, onRaiseHand: () {
                  final me = AuthService.instance.currentUser?.displayName ?? 'User';
                  ActiveSessionManager.instance.raiseHand(me);
                  _handAnimController?.repeat(reverse: true);
                  Future.delayed(const Duration(seconds: 3), () => _handAnimController?.stop());
                  setState(() {});
                },
                onHint: () => setState(() => _currentHint = AppLocalizations.of(context, 'session_hint_default')),
              ),
            ),

            if (_currentHint != null)
              Positioned(bottom: 120, left: 24, right: 24, child: HintBox(hint: _currentHint!, onClose: () => setState(() => _currentHint = null))),
          ],
        ),
      ),
      endDrawer: _HTDrawer(
        speakers: _speakers, audience: _audience,
        handRaised: ActiveSessionManager.instance.handRaised, me: me, isAdmin: isAdmin,
        onApprove: (n) => ActiveSessionManager.instance.moveToStage(n),
        onKick: (n) => setState(() => _kickedPlayers.add(n)),
        onEnd: () => ActiveSessionManager.instance.closeSession(),
      ),
    );
  }
}

class _HTSpeakerTile extends StatelessWidget {
  const _HTSpeakerTile({required this.name, required this.emoji, this.animation});
  final String name;
  final String emoji;
  final Animation<double>? animation;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        AnimatedBuilder(
          animation: animation ?? const AlwaysStoppedAnimation(0.0),
          builder: (context, child) {
            final pulsing = animation != null && animation!.value > 0.01;
            return Container(
              padding: EdgeInsets.all(pulsing ? 2.5 : 0),
              decoration: BoxDecoration(shape: BoxShape.circle, border: pulsing ? Border.all(color: Colors.greenAccent, width: 2) : null),
              child: CircleAvatar(radius: 35, backgroundColor: Colors.white.withOpacity(0.1), child: Text(emoji, style: const TextStyle(fontSize: 30))),
            );
          },
        ),
        const SizedBox(height: 6),
        Text(name, style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.w700), maxLines: 1, overflow: TextOverflow.ellipsis),
      ],
    );
  }
}

class _HTEmptySeat extends StatelessWidget {
  const _HTEmptySeat();

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 70, height: 70,
          decoration: BoxDecoration(color: Colors.white.withOpacity(0.08), shape: BoxShape.circle),
          child: const Icon(Icons.person_pin_circle_outlined, color: Colors.white24, size: 32),
        ),
        const SizedBox(height: 6),
        Text(AppLocalizations.of(context, 'join_seat'), style: const TextStyle(color: Colors.white24, fontSize: 10, fontWeight: FontWeight.w500)),
      ],
    );
  }
}

class _RefinedHTFooter extends StatelessWidget {
  const _RefinedHTFooter({
    required this.isSpeaker, required this.recording, required this.isTyping,
    required this.textController, required this.handAnimation, required this.onSendText, required this.onMicDown, required this.onMicUp, required this.onRaiseHand, required this.onHint,
  });

  final bool isSpeaker; final bool recording; final bool isTyping;
  final TextEditingController textController;
  final AnimationController? handAnimation;
  final VoidCallback onSendText; final void Function(PointerDownEvent) onMicDown; final void Function(PointerEvent) onMicUp;
  final VoidCallback onRaiseHand; final VoidCallback onHint;

  @override
  Widget build(BuildContext context) {
    final hasHand = ActiveSessionManager.instance.handRaised.contains(AuthService.instance.currentUser?.displayName);
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 36),
      decoration: const BoxDecoration(
        color: Color(0xFF0D1B2A),
        borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        boxShadow: [BoxShadow(color: Colors.black45, blurRadius: 15, offset: Offset(0, -5))],
      ),
      child: Row(
        children: [
          Expanded(
            child: Container(
              height: 46,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              decoration: BoxDecoration(color: Colors.white.withOpacity(0.06), borderRadius: BorderRadius.circular(24)),
              child: TextField(
                controller: textController,
                style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w500),
                decoration: InputDecoration(hintText: AppLocalizations.of(context, 'write_comment'), border: InputBorder.none, hintStyle: const TextStyle(color: Colors.white24, fontSize: 12)),
                onSubmitted: (_) => onSendText(),
              ),
            ),
          ),
          const SizedBox(width: 12),
          if (isTyping)
            GestureDetector(
              onTap: onSendText,
              child: Container(width: 46, height: 46, decoration: const BoxDecoration(gradient: AppTheme.primaryGradient, shape: BoxShape.circle), child: const Icon(Icons.send_rounded, color: Colors.white, size: 20)),
            )
          else ...[
            if (handAnimation != null)
              ScaleTransition(
                scale: Tween<double>(begin: 1.0, end: 1.2).animate(CurvedAnimation(parent: handAnimation!, curve: Curves.easeInOut)),
                child: IconButton(icon: Icon(Icons.pan_tool_rounded, color: hasHand ? Colors.amber : Colors.white70, size: 24), onPressed: onRaiseHand),
              )
            else
              IconButton(icon: Icon(Icons.pan_tool_rounded, color: hasHand ? Colors.amber : Colors.white70, size: 24), onPressed: onRaiseHand),
            IconButton(icon: const Icon(Icons.lightbulb_outline_rounded, color: Colors.amber, size: 24), onPressed: onHint),
            const SizedBox(width: 4),
            Opacity(
              opacity: isSpeaker ? 1.0 : 0.3,
              child: Listener(
                onPointerDown: onMicDown, onPointerUp: onMicUp,
                child: Container(
                  width: 52, height: 52,
                  decoration: BoxDecoration(gradient: recording ? const LinearGradient(colors: [Colors.red, Colors.redAccent]) : AppTheme.primaryGradient, shape: BoxShape.circle),
                  child: Icon(recording ? Icons.stop_rounded : Icons.mic_rounded, color: Colors.white, size: 26),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _HTDrawer extends StatelessWidget {
  const _HTDrawer({required this.speakers, required this.audience, required this.handRaised, required this.me, required this.isAdmin, required this.onApprove, required this.onKick, required this.onEnd});
  final List<String> speakers; final List<String> audience; final Set<String> handRaised; final String me; final bool isAdmin; final Function(String) onApprove; final Function(String) onKick; final VoidCallback onEnd;

  @override
  Widget build(BuildContext context) {
    return Drawer(
      backgroundColor: const Color(0xFF0D1B2A),
      child: Column(
        children: [
          DrawerHeader(
            decoration: const BoxDecoration(gradient: AppTheme.primaryGradient),
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.groups_rounded, color: Colors.white, size: 48),
                  const SizedBox(height: 8),
                  Text(AppLocalizations.of(context, 'room_management'), style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 20)),
                ],
              ),
            ),
          ),
          Expanded(
            child: ListView(
              children: [
                ListTile(title: Text(AppLocalizations.of(context, 'speakers'), style: const TextStyle(color: Colors.white70, fontSize: 11, fontWeight: FontWeight.bold))),
                for (final n in speakers)
                  ListTile(
                    leading: CircleAvatar(radius: 16, child: Text(n[0])),
                    title: Text(n == me ? AppLocalizations.of(context, 'session_you') : n, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
                    trailing: isAdmin && n != me ? IconButton(icon: const Icon(Icons.gavel, color: Colors.redAccent), onPressed: () => onKick(n)) : null,
                  ),
                const Divider(color: Colors.white10),
                ListTile(title: Text(AppLocalizations.of(context, 'explore'), style: const TextStyle(color: Colors.white70, fontSize: 11, fontWeight: FontWeight.bold))),
                ListTile(
                  leading: const Icon(Icons.explore_rounded, color: Colors.blueAccent),
                  title: Text(AppLocalizations.of(context, 'discover_topics'), style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w500)),
                  onTap: () => Navigator.of(context).popUntil((route) => route.isFirst),
                ),
                const Divider(color: Colors.white10),
                ListTile(title: Text(AppLocalizations.of(context, 'audience'), style: const TextStyle(color: Colors.white70, fontSize: 11, fontWeight: FontWeight.bold))),
                for (final n in audience)
                  ListTile(
                    leading: CircleAvatar(radius: 16, child: Text(n[0])),
                    title: Text(n, style: const TextStyle(color: Colors.white)),
                    trailing: isAdmin && handRaised.contains(n) ? IconButton(icon: const Icon(Icons.stars_rounded, color: Colors.greenAccent), onPressed: () => onApprove(n)) : null,
                  ),
              ],
            ),
          ),
          if (isAdmin)
            Padding(padding: const EdgeInsets.all(24), child: ElevatedButton.icon(onPressed: onEnd, icon: const Icon(Icons.power_settings_new_rounded), label: Text(AppLocalizations.of(context, 'end_session')), style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent, foregroundColor: Colors.white, minimumSize: const Size.fromHeight(56), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16))))),
        ],
      ),
    );
  }
}
