import 'dart:async';
import 'package:flutter/material.dart';

import '../data/session_scripts.dart';
import '../models/session_line.dart';
import '../models/topic_vote.dart';
import '../services/voice_capture.dart';
import '../services/active_session_manager.dart';
import '../theme/app_theme.dart';
import 'feedback_screen.dart';

// ВОТ ЭТИ ДВА ИМПОРТА ОБЯЗАТЕЛЬНЫ:
import '../services/practice_service.dart';
import '../services/topic_service.dart';

import '../widgets/session_widgets.dart';

class SessionScreen extends StatefulWidget {
  const SessionScreen({super.key, required this.topic});

  final TopicVote topic;

  @override
  State<SessionScreen> createState() => _SessionScreenState();
}

class _SessionScreenState extends State<SessionScreen> with TickerProviderStateMixin {
  late List<SessionLine> _script;

  int _cursor = 0;
  bool _recording = false;
  bool _busy = false;
  DateTime? _micPressStart;

  String? _currentHint;
  String? _currentMood;
  String? _currentReaction;
  String _sceneDescription = "Setting the stage...";
  Timer? _idleTimer;

  late AnimationController _pulsePartner;
  late AnimationController _pulseMe;

  late VoiceCapture _voice;
  String? _backendSessionId;

  final List<String> _quickActions = ["Agree", "Disagree", "Ask why?", "Suggest idea"];

  @override
  void initState() {
    super.initState();
    _voice = VoiceCapture();
    _script = SessionScripts.forTopic(widget.topic.id);

    _pulsePartner = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));
    _pulseMe = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));

    _initAll();
  }

  Future<void> _initAll() async {
    if (!ActiveSessionManager.instance.isActive) {
      ActiveSessionManager.instance.startTopic(widget.topic);
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _startSession());
  }

  @override
  void dispose() {
    _idleTimer?.cancel();
    _voice.dispose();
    _pulsePartner.dispose();
    _pulseMe.dispose();
    super.dispose();
  }

  Future<void> _startSession() async {
    setState(() => _busy = true);
    try {
      String realTopicId = widget.topic.id;

      // 1. Проверяем, если ID временный (начинается на 'custom_')
      if (realTopicId.startsWith('custom_')) {
        debugPrint("Creating topic in DB for custom prompt...");

        // Создаем тему на бэкенде с ПРАВИЛЬНЫМИ параметрами
        final topicResponse = await TopicService.instance.createTopic(
          title: widget.topic.title,
          description: widget.topic.publicContext.isEmpty
              ? "Custom user-generated topic"
              : widget.topic.publicContext,
          // Вместо prompt используем goal (как требует сервис)
          goal: widget.topic.goal.isEmpty ? widget.topic.title : widget.topic.goal,
          myRole: widget.topic.myRole,
          partnerRole: widget.topic.partnerRole,
          partnerEmoji: widget.topic.aiEmoji,
        );

        // Получаем реальный UUID, созданный базой данных Go
        realTopicId = topicResponse['id'];
        debugPrint("Topic created successfully in DB with UUID: $realTopicId");
      }

      // 2. Теперь стартуем сессию с НОРМАЛЬНЫМ UUID
      debugPrint("Starting practice session with topic UUID: $realTopicId");
      final sessionData = await PracticeService.instance.startSession(realTopicId);

      if (mounted) {
        setState(() {
          _backendSessionId = sessionData['id'];
          _busy = false;
        });
        debugPrint("Session initialized on backend: $_backendSessionId");
      }

    } catch (e) {
      debugPrint("CRITICAL ERROR during session startup: $e");
      if (mounted) {
        setState(() => _busy = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text("Error: $e"), backgroundColor: Colors.redAccent),
        );
      }
    }

    // Запускаем сценарий (реплики партнера)
    await _flushPartnerLines();
  }

  Future<void> _finishVoiceSend() async {
    if (_backendSessionId == null) return;
    setState(() => _busy = true);
    final path = await _voice.stop();
    if (path != null) {
      try {
        await PracticeService.instance.sendVoiceTurn(
          sessionId: _backendSessionId!,
          audioPath: path,
          language: 'English',
        );
      } catch (e) {
        debugPrint("Error in sendVoiceTurn: $e");
      }
    }
    if (mounted) setState(() => _busy = false);
    _startIdleTimer();
    unawaited(_flushPartnerLines());
  }

  Future<void> _flushPartnerLines() async {
    if (_cursor >= _script.length) return;
    while (_cursor < _script.length && !_script[_cursor].isUser) {
      final line = _script[_cursor];
      if (line.type == LineType.system) {
        setState(() => _sceneDescription = line.content.replaceAll("Режиссер: ", ""));
        _cursor++;
      } else {
        setState(() { _currentMood = line.mood; _currentReaction = line.reaction; });
        _pulsePartner.repeat(reverse: true);
        await Future<void>.delayed(Duration(milliseconds: 1000 + (line.content.length * 40)));
        if (!mounted) return;
        _pulsePartner.stop(); _pulsePartner.value = 0.0;
        _cursor++;
      }
      await Future<void>.delayed(const Duration(milliseconds: 400));
    }
    if (mounted) { setState(() {}); _startIdleTimer(); }
  }

  void _onMicDown(PointerDownEvent e) {
    if (_busy) return;
    _resetIdleTimer();
    _micPressStart = DateTime.now();
    setState(() => _recording = true);
    _pulseMe.repeat(reverse: true);
    _voice.start();
  }

  void _onMicUp(PointerEvent e) {
    if (_micPressStart == null) return;
    final sec = DateTime.now().difference(_micPressStart!).inMilliseconds / 1000.0;
    _micPressStart = null;
    setState(() => _recording = false);
    _pulseMe.stop(); _pulseMe.value = 0.0;
    if (sec < 0.4) { _voice.stop(); _startIdleTimer(); return; }
    _finishVoiceSend();
  }

  void _resetIdleTimer() {
    _idleTimer?.cancel();
    setState(() => _currentHint = null);
  }

  void _startIdleTimer() {
    _resetIdleTimer();
    _idleTimer = Timer(const Duration(seconds: 10), () {
      if (!mounted || _recording || _busy) return;
      _showHint();
    });
  }

  void _showHint() {
    final line = _cursor < _script.length ? _script[_cursor] : null;
    setState(() => _currentHint = line?.hint ?? "Your turn!");
  }

  void _goFeedback() {
    ActiveSessionManager.instance.removeUnfinished(widget.topic.id);
    Navigator.pushReplacement(context, MaterialPageRoute(builder: (_) => FeedbackScreen(topic: widget.topic)));
  }

  void _minimize() {
    ActiveSessionManager.instance.markAsUnfinished(widget.topic);
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final progress = (_cursor / _script.length).clamp(0.01, 1.0);

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      appBar: AppBar(
        leading: IconButton(icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 32), onPressed: _minimize),
        title: Text(widget.topic.title, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w900)),
        actions: [
          IconButton(icon: const Icon(Icons.check_circle_outline_rounded, color: Colors.greenAccent), onPressed: _goFeedback),
        ],
      ),
      body: Stack(
        children: [
          Positioned.fill(
            child: AnimatedContainer(
              duration: const Duration(seconds: 2),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter, end: Alignment.bottomCenter,
                  colors: [
                    _currentMood == 'Serious' ? Colors.red.withOpacity(0.05) : AppTheme.primary.withOpacity(0.05),
                    theme.scaffoldBackgroundColor,
                  ],
                ),
              ),
            ),
          ),
          SafeArea(
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 4),
                  child: LinearProgressIndicator(
                      value: progress, minHeight: 3,
                      backgroundColor: theme.dividerColor.withOpacity(0.1), color: AppTheme.primary
                  ),
                ),
                Container(
                  width: double.infinity,
                  margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: isDark ? Colors.white.withOpacity(0.05) : Colors.white,
                    borderRadius: BorderRadius.circular(24),
                    boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10)],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.auto_awesome_rounded, size: 14, color: AppTheme.primary),
                          const SizedBox(width: 6),
                          const Text('SCENE', style: TextStyle(fontSize: 9, fontWeight: FontWeight.w900, color: Colors.grey)),
                          const Spacer(),
                          if (_busy) const SizedBox(width: 10, height: 10, child: CircularProgressIndicator(strokeWidth: 2)),
                        ],
                      ),
                      const SizedBox(height: 6),
                      Text(_sceneDescription, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, height: 1.3), maxLines: 3, overflow: TextOverflow.ellipsis),
                    ],
                  ),
                ),
                Expanded(
                  child: SingleChildScrollView(
                    physics: const BouncingScrollPhysics(),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Stack(
                            alignment: Alignment.center,
                            children: [
                              SessionAvatar(name: widget.topic.partnerRole, emoji: widget.topic.aiEmoji, animation: _pulsePartner, isLarge: true),
                              if (_currentReaction != null)
                                Positioned(top: 0, right: -10, child: Text(_currentReaction!, style: const TextStyle(fontSize: 36))),
                            ],
                          ),
                          const SizedBox(height: 20),
                          SingleChildScrollView(
                            scrollDirection: Axis.horizontal,
                            padding: const EdgeInsets.symmetric(horizontal: 24),
                            child: Row(
                              children: _quickActions.map((a) => Padding(
                                padding: const EdgeInsets.only(right: 8),
                                child: ActionChip(
                                  label: Text(a), onPressed: () {},
                                  backgroundColor: AppTheme.primary.withOpacity(0.05),
                                  labelStyle: const TextStyle(fontSize: 11, color: AppTheme.primary, fontWeight: FontWeight.bold),
                                ),
                              )).toList(),
                            ),
                          ),
                          const SizedBox(height: 20),
                          SessionAvatar(name: "You (${widget.topic.myRole})", emoji: '👤', animation: _pulseMe, isLarge: false),
                        ],
                      ),
                    ),
                  ),
                ),
                if (_currentHint != null)
                  Padding(
                    padding: const EdgeInsets.fromLTRB(20, 0, 20, 10),
                    child: HintBox(hint: _currentHint!, onClose: () => setState(() => _currentHint = null)),
                  ),
                Padding(
                  padding: const EdgeInsets.only(bottom: 24),
                  child: Listener(
                    onPointerDown: _busy ? null : _onMicDown,
                    onPointerUp: _busy ? null : _onMicUp,
                    child: Stack(
                      alignment: Alignment.center,
                      children: [
                        if (_recording) const PulseRipple(),
                        AnimatedContainer(
                          duration: const Duration(milliseconds: 200),
                          width: _recording ? 94 : 82, height: _recording ? 94 : 82,
                          decoration: BoxDecoration(
                            gradient: _recording ? const LinearGradient(colors: [Colors.red, Colors.redAccent]) : AppTheme.primaryGradient,
                            shape: BoxShape.circle,
                            boxShadow: [BoxShadow(color: (_recording ? Colors.red : AppTheme.primary).withOpacity(0.3), blurRadius: 15, offset: const Offset(0, 6))],
                          ),
                          child: Icon(_busy ? Icons.hourglass_empty_rounded : (_recording ? Icons.stop_rounded : Icons.mic_rounded), color: Colors.white, size: 36),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}