import 'package:flutter/material.dart';
import '../models/topic_vote.dart';
import '../theme/app_theme.dart';
import 'session_screen.dart';

class PrepScreen extends StatefulWidget {
  const PrepScreen({super.key, required this.topic});

  final TopicVote topic;

  @override
  State<PrepScreen> createState() => _PrepScreenState();
}

class _PrepScreenState extends State<PrepScreen> {

  @override
  Widget build(BuildContext context) {
    final t = widget.topic;
    return Scaffold(
      backgroundColor: AppTheme.background,
      appBar: AppBar(
        title: Text('Лобби: ${t.emoji}'),
      ),
      body: ListView(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        children: [
          Text(
            t.title,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          _InfoBlock(
            title: 'СЕТТИНГ',
            content: t.publicContext,
            icon: Icons.map_outlined,
          ),
          const SizedBox(height: 16),
          _InfoBlock(
            title: 'КОНТРАГЕНТ',
            content: t.partnerRole,
            icon: Icons.person_outline,
            bgColor: const Color(0xFFF3F4F6),
          ),
          const SizedBox(height: 24),
          Text(
            'Ваша роль',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: 8),
          Text(
            t.myRole,
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppTheme.primary),
          ),
          const SizedBox(height: 16),
        ],
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
          child: SizedBox(
            width: double.infinity,
            height: 54,
            child: ElevatedButton.icon(
              onPressed: () {
                Navigator.of(context).pushReplacement(
                  MaterialPageRoute<void>(
                    builder: (_) => SessionScreen(topic: t),
                  ),
                );
              },
              icon: const Icon(Icons.play_arrow_rounded),
              label: const Text('Начать тренировку (Start)'),
            ),
          ),
        ),
      ),
    );
  }
}

class _InfoBlock extends StatelessWidget {
  final String title;
  final String content;
  final IconData icon;
  final Color bgColor;

  const _InfoBlock({
    required this.title,
    required this.content,
    required this.icon,
    this.bgColor = AppTheme.primarySoft,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 18, color: AppTheme.textSecondary),
              const SizedBox(width: 6),
              Text(
                title,
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      fontWeight: FontWeight.w800,
                      color: AppTheme.textSecondary,
                      letterSpacing: 0.06,
                    ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(content, style: Theme.of(context).textTheme.bodyMedium?.copyWith(height: 1.45, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}
