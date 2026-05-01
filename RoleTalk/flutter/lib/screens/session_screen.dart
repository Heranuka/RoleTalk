import 'dart:async';
import 'package:flutter/material.dart';

import '../data/session_scripts.dart';
import '../models/session_line.dart';
import '../models/topic_vote.dart';
import '../services/voice_capture.dart';
import '../services/app_localizations.dart'; // Локализация
import '../theme/app_theme.dart';
import 'feedback_screen.dart';
import '../services/practice_service.dart';

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
  Timer? _idleTimer;

  late AnimationController _pulsePartner;
  late AnimationController _pulseMe;

  late VoiceCapture _voice;
  String? _backendSessionId;

  @override
  void initState() {
    super.initState();
    _voice = VoiceCapture();
    _script = SessionScripts.forTopic(widget.topic.id);

    _pulsePartner = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));
    _pulseMe = AnimationController(vsync: this, duration: const Duration(milliseconds: 900));

    WidgetsBinding.instance.addPostFrameCallback((_) => _startSession());
  }

  @override
  void dispose() {
    _idleTimer?.cancel();
    unawaited(_voice.dispose());
    _pulsePartner.dispose();
    _pulseMe.dispose();
    super.dispose();
  }

  // Переход на экран фидбека (исправлено)
  void _goFeedback() {
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (_) => FeedbackScreen(topic: widget.topic)),
    );
  }

  void _resetIdleTimer() {
    _idleTimer?.cancel();
    setState(() {
      _currentHint = null;
    });
  }

  void _startIdleTimer() {
    _resetIdleTimer();
    _idleTimer = Timer(const Duration(seconds: 7), () {
      if (!mounted || _recording || _busy) return;
      _showHint();
    });
  }

  void _showHint() {
    if (_cursor < _script.length && _script[_cursor].isUser && _script[_cursor].hint != null) {
      setState(() {
        _currentHint = '💡 ${_script[_cursor].hint}';
      });
    } else {
      setState(() {
        _currentHint = '💡 ${AppLocalizations.of(context, 'session_hint_default')}';
      });
    }
  }

  Future<void> _startSession() async {
    try {
      final sessionData = await PracticeService.instance.startSession(widget.topic.id);
      _backendSessionId = sessionData['id'];
    } catch (e) {
      debugPrint("Error starting session: $e");
      // Fallback or show error
    }
    await Future<void>.delayed(const Duration(milliseconds: 800));
    await _flushPartnerLines();
  }

  Future<void> _flushPartnerLines() async {
    if (_cursor >= _script.length) return;

    while (_cursor < _script.length && !_script[_cursor].isUser) {
      final line = _script[_cursor];

      if (line.type == LineType.system) {
        _cursor++;
      } else {
        _pulsePartner.repeat(reverse: true);
        final durationMs = 1500 + (line.content.length * 60);
        await Future<void>.delayed(Duration(milliseconds: durationMs));

        if (!mounted) return;
        _pulsePartner.stop();
        _pulsePartner.value = 0.0;
        _cursor++;
      }
      await Future<void>.delayed(const Duration(milliseconds: 500));
    }

    if (mounted) {
      setState(() {});
      _startIdleTimer();
    }
  }

  void _onMicDown(PointerDownEvent e) {
    if (_busy) return;
    _resetIdleTimer();
    _micPressStart = DateTime.now();
    setState(() => _recording = true);
    _pulseMe.repeat(reverse: true);
    unawaited(_voice.start());
  }

  void _onMicUp(PointerEvent e) {
    if (_micPressStart == null) return;
    final sec = DateTime.now().difference(_micPressStart!).inMilliseconds / 1000.0;
    _micPressStart = null;
    setState(() => _recording = false);
    _pulseMe.stop();
    _pulseMe.value = 0.0;

    if (sec < 0.4) {
      unawaited(_voice.stop());
      _startIdleTimer();
      return;
    }
    unawaited(_finishVoiceSend());
  }

  Future<void> _finishVoiceSend() async {
    if (_backendSessionId == null) return;
    setState(() => _busy = true);

    // 1. Останавливаем запись
    final recordPath = await _voice.stop();
    if (recordPath != null) {
      try {
        // 2. Отправляем на бэкенд
        final response = await PracticeService.instance.sendVoiceTurn(
          sessionId: _backendSessionId!,
          audioPath: recordPath,
          language: 'English', // Можно брать из настроек
        );

        final aiText = response['ai_response']['text'];
        final aiAudioUrl = response['ai_response']['audio_url'];

        if (aiAudioUrl != null) {
          // 3. Имитируем, что партнер начал говорить
          _pulsePartner.repeat(reverse: true);

          // 4. Проигрываем голос ИИ (ApiClient.baseUrl может понадобиться если URL относительный)
          // Если аудио_url полный - ок, если нет - надо склеить.
          String fullAudioUrl = aiAudioUrl;
          if (!fullAudioUrl.startsWith('http')) {
            fullAudioUrl = "http://localhost:8080$aiAudioUrl";
          }
          
          await playVoiceFile(fullAudioUrl);

          _pulsePartner.stop();
          _pulsePartner.value = 0.0;
        }
      } catch (e) {
        debugPrint("Error in voice turn: $e");
      }
    }

    if (mounted) setState(() => _busy = false);
    _startIdleTimer();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: theme.scaffoldBackgroundColor,
      body: SafeArea(
        child: Column(
          children: [
            // TOP BAR
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          widget.topic.title,
                          style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w900),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 2),
                        Text(
                          '${AppLocalizations.of(context, 'session_goal')}: ${widget.topic.goal}',
                          style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppTheme.primary),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    icon: Icon(Icons.close_rounded, color: theme.hintColor),
                    onPressed: _goFeedback,
                  ),
                ],
              ),
            ),

            // AVATARS AREA (Call Style)
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  _RoomAvatar(
                    name: widget.topic.partnerRole,
                    emoji: widget.topic.aiEmoji,
                    animation: _pulsePartner,
                    isLarge: true,
                  ),
                  const SizedBox(height: 60),
                  _RoomAvatar(
                    name: "${AppLocalizations.of(context, 'session_you')} (${widget.topic.myRole})",
                    emoji: '👤',
                    animation: _pulseMe,
                    isLarge: false,
                  ),
                ],
              ),
            ),

            // AI HINT BOX
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24),
              child: SizedBox(
                height: 80, // Фиксированная высота, чтобы UI не прыгал
                child: _currentHint != null
                    ? AnimatedOpacity(
                  opacity: 1.0,
                  duration: const Duration(milliseconds: 300),
                  child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: isDark ? Colors.white.withOpacity(0.05) : AppTheme.primarySoft,
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: AppTheme.primary.withOpacity(0.2)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.lightbulb_outline, color: AppTheme.primary, size: 20),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            _currentHint!,
                            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                          ),
                        ),
                      ],
                    ),
                  ),
                )
                    : const SizedBox.shrink(),
              ),
            ),

            // BOTTOM CONTROLS
            Padding(
              padding: const EdgeInsets.only(bottom: 40, top: 20),
              child: Column(
                children: [
                  if (_busy && !_recording)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 16),
                      child: Text(
                          AppLocalizations.of(context, 'session_processing'),
                          style: TextStyle(color: theme.hintColor, fontSize: 12, fontWeight: FontWeight.bold)
                      ),
                    ),
                  Listener(
                    behavior: HitTestBehavior.opaque,
                    onPointerDown: _busy ? null : _onMicDown,
                    onPointerUp: _busy ? null : _onMicUp,
                    onPointerCancel: _busy ? null : _onMicUp,
                    child: Stack(
                      alignment: Alignment.center,
                      children: [
                        if (_recording)
                          _PulseRipple(), // Вынес круги пульсации в отдельный виджет
                        Container(
                          width: 80,
                          height: 80,
                          decoration: BoxDecoration(
                            color: _recording ? const Color(0xFFEF4444) : AppTheme.primary,
                            shape: BoxShape.circle,
                            boxShadow: [
                              BoxShadow(
                                  color: AppTheme.primary.withOpacity(0.3),
                                  blurRadius: 20,
                                  offset: const Offset(0, 8)
                              )
                            ],
                          ),
                          child: Icon(
                              _recording ? Icons.stop_rounded : Icons.mic_rounded,
                              color: Colors.white,
                              size: 40
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Виджет пульсации для микрофона
class _PulseRipple extends StatefulWidget {
  @override
  State<_PulseRipple> createState() => _PulseRippleState();
}

class _PulseRippleState extends State<_PulseRipple> with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: const Duration(seconds: 1))..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Transform.scale(
          scale: 1.0 + _controller.value * 0.5,
          child: Opacity(
            opacity: 1.0 - _controller.value,
            child: Container(
              width: 100,
              height: 100,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: const Color(0xFFEF4444).withOpacity(0.4),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _RoomAvatar extends StatelessWidget {
  const _RoomAvatar({
    required this.name,
    required this.emoji,
    this.animation,
    this.isLarge = false,
  });

  final String name;
  final String emoji;
  final Animation<double>? animation;
  final bool isLarge;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final double radius = isLarge ? 65 : 45;

    return Column(
      children: [
        AnimatedBuilder(
          animation: animation!,
          builder: (context, child) {
            final double value = animation!.value;
            return Container(
              padding: const EdgeInsets.all(4),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: value > 0.01 ? AppTheme.primary : Colors.transparent,
                  width: 3,
                ),
                boxShadow: value > 0.01
                    ? [BoxShadow(color: AppTheme.primary.withOpacity(0.2 * value), blurRadius: 15 * value, spreadRadius: 5 * value)]
                    : [],
              ),
              child: CircleAvatar(
                radius: radius,
                backgroundColor: theme.cardColor,
                child: Text(emoji, style: TextStyle(fontSize: radius * 0.8)),
              ),
            );
          },
        ),
        const SizedBox(height: 12),
        Text(
          name,
          style: theme.textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w800),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ],
    );
  }
}