import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class SkillsScreen extends StatelessWidget {
  const SkillsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Мои навыки')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('Прогресс на основе ваших сессий', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 24),
          _skillRow('Эмпатия', 0.8),
          _skillRow('Убедительность', 0.6),
          _skillRow('Логика и Структура', 0.9),
          _skillRow('Стрессоустойчивость', 0.4),
          const SizedBox(height: 40),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(color: AppTheme.primarySoft, borderRadius: BorderRadius.circular(12)),
            child: const Row(
              children: [
                Icon(Icons.lightbulb_outline, color: AppTheme.primary),
                SizedBox(width: 12),
                Expanded(child: Text('Совет: Чтобы поднять логику, выбирайте темы с дебатами.')),
              ],
            ),
          )
        ],
      ),
    );
  }

  Widget _skillRow(String label, double val) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(label, style: const TextStyle(fontWeight: FontWeight.w700)),
              Text('${(val * 100).toInt()}%', style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.primary)),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: val,
            minHeight: 8,
            borderRadius: BorderRadius.circular(10),
            backgroundColor: AppTheme.primarySoft,
            color: AppTheme.primary,
          ),
        ],
      ),
    );
  }
}