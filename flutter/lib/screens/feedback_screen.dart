import 'package:flutter/material.dart';
import '../data/feedback_data.dart';
import '../models/topic_vote.dart';
import '../services/history_store.dart';
import '../theme/app_theme.dart';

class FeedbackScreen extends StatefulWidget {
  const FeedbackScreen({super.key, required this.topic});

  final TopicVote topic;

  @override
  State<FeedbackScreen> createState() => _FeedbackScreenState();
}

class _FeedbackScreenState extends State<FeedbackScreen> {
  bool _animated = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) setState(() => _animated = true);
      HistoryStore.instance.addAiSession(topicTitle: widget.topic.title);
    });
  }

  @override
  Widget build(BuildContext context) {
    final f = FeedbackRepository.forTopic(widget.topic.id);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Разбор: Вердикт ИИ-режиссера'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppTheme.primarySoft,
              borderRadius: BorderRadius.circular(16),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    Icon(Icons.analytics_rounded, color: AppTheme.primary),
                    SizedBox(width: 8),
                    Text(
                      'Оценки отыгрыша',
                      style: TextStyle(fontWeight: FontWeight.w800, fontSize: 16, color: AppTheme.primary),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _BarRow(label: 'Следование роли', value: f.roleplayScore, animated: _animated, color: AppTheme.primary),
                const SizedBox(height: 14),
                _BarRow(label: 'Логика и факты', value: f.logicScore, animated: _animated, color: const Color(0xFF10B981)),
                const SizedBox(height: 14),
                _BarRow(label: 'Эмпатия/Тактичность', value: f.empathyScore, animated: _animated, color: const Color(0xFFF59E0B)),
              ],
            ),
          ),
          const SizedBox(height: 32),
          Text(
            'Что получилось отлично',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800, color: const Color(0xFF059669)),
          ),
          const SizedBox(height: 12),
          for (final text in f.highlights)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('✅ ', style: TextStyle(fontSize: 18)),
                  Expanded(
                    child: Text(
                      text,
                      style: const TextStyle(fontSize: 15, height: 1.4, color: AppTheme.textPrimary),
                    ),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 24),
          Text(
            'Зоны роста (Заметки ИИ)',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800, color: const Color(0xFFDC2626)),
          ),
          const SizedBox(height: 12),
          for (final text in f.improvements)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('💡 ', style: TextStyle(fontSize: 18)),
                  Expanded(
                    child: Text(
                      text,
                      style: const TextStyle(fontSize: 15, height: 1.4, color: AppTheme.textPrimary),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
          child: ElevatedButton(
            onPressed: () {
              Navigator.of(context).popUntil((route) => route.isFirst);
            },
            child: const Text('Завершить'),
          ),
        ),
      ),
    );
  }
}

class _BarRow extends StatelessWidget {
  const _BarRow({
    required this.label,
    required this.value,
    required this.animated,
    required this.color,
  });

  final String label;
  final int value;
  final bool animated;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: 140,
          child: Text(
            label,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: AppTheme.textSecondary,
                ),
          ),
        ),
        Expanded(
          child: ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: TweenAnimationBuilder<double>(
              key: ValueKey<bool>(animated),
              tween: Tween(begin: 0, end: animated ? value / 100.0 : 0),
              duration: const Duration(milliseconds: 900),
              curve: Curves.easeOutCubic,
              builder: (context, t, _) {
                return LinearProgressIndicator(
                  value: t,
                  minHeight: 10,
                  backgroundColor: const Color(0xFFE8EAEF),
                  color: color,
                );
              },
            ),
          ),
        ),
        const SizedBox(width: 12),
        SizedBox(
          width: 36,
          child: Text(
            '$value%',
            textAlign: TextAlign.end,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
        ),
      ],
    );
  }
}
