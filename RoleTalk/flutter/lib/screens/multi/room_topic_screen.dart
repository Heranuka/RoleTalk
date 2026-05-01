import 'dart:math';

import 'package:flutter/material.dart';

import '../../data/mock_repository.dart';
import '../../data/multiplayer_config.dart';
import '../../models/topic_vote.dart';
import '../../models/voice_room.dart';
import '../../theme/app_theme.dart';
import 'human_session_screen.dart';

/// После лобби: вся команда вводит (или выбирает) тему — дальше роли и чат.
class RoomTopicScreen extends StatefulWidget {
  const RoomTopicScreen({
    super.key,
    required this.room,
    required this.playerNames,
  });

  final VoiceRoom room;
  final List<String> playerNames;

  @override
  State<RoomTopicScreen> createState() => _RoomTopicScreenState();
}

class _RoomTopicScreenState extends State<RoomTopicScreen> {
  final _ctrl = TextEditingController();
  final _rng = Random();

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  void _go() {
    final topic = TopicVote.custom(_ctrl.text);
    final roles = MultiplayerConfig.assignRoles(widget.playerNames, 'custom', _rng);
    if (!mounted) return;
    Navigator.of(context).pushReplacement(
      MaterialPageRoute<void>(
        builder: (_) => HumanSessionScreen(
          topic: topic,
          playerRoles: roles,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final suggestions = MockRepository.topics.map((t) => t.title).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Тема сессии'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
        children: [
          Row(
            children: [
              Text(widget.room.emoji, style: const TextStyle(fontSize: 36)),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.room.title,
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800),
                    ),
                    Text(
                      'Команда собрана — введите тему сцены; роли уже распределены, дальше свободный диалог.',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppTheme.textSecondary, height: 1.35),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Text(
            'Ваша тема',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _ctrl,
            maxLines: 3,
            minLines: 2,
            textCapitalization: TextCapitalization.sentences,
            decoration: InputDecoration(
              hintText: 'Например: спор о чаевых в ресторане, переговоры о зарплате…',
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide(color: AppTheme.primary.withValues(alpha: 0.25)),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide(color: AppTheme.primary.withValues(alpha: 0.2)),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: const BorderSide(color: AppTheme.primary, width: 1.4),
              ),
            ),
            onSubmitted: (_) => _go(),
          ),
          const SizedBox(height: 16),
          Text(
            'Или подсказка одним нажатием',
            style: Theme.of(context).textTheme.labelLarge?.copyWith(color: AppTheme.textSecondary, fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 10),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final s in suggestions)
                ActionChip(
                  label: Text(s, style: const TextStyle(fontSize: 12.5, fontWeight: FontWeight.w600)),
                  onPressed: () {
                    _ctrl.text = s;
                    setState(() {});
                  },
                ),
            ],
          ),
          const SizedBox(height: 28),
          FilledButton(
            onPressed: _go,
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(52),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            ),
            child: const Text('Перейти в комнату', style: TextStyle(fontWeight: FontWeight.w800, fontSize: 16)),
          ),
        ],
      ),
    );
  }
}
