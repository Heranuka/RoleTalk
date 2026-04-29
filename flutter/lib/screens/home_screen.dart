import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/app_localizations.dart';
import '../services/settings_store.dart';
import '../models/topic_vote.dart'; // Добавь этот импорт
import 'multi/community_themes_screen.dart';
import 'prep_screen.dart'; // Добавь этот импорт

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final _aiController = TextEditingController();

  @override
  void dispose() {
    _aiController.dispose();
    super.dispose();
  }

  // Хелпер для запуска сценария
  void _startScenario(TopicVote topic) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => PrepScreen(topic: topic)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        centerTitle: false,
        title: Text(
          AppLocalizations.of(context, 'tab_ai'), 
          style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 24)
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        children: [
          const SizedBox(height: 12),
          _buildAIQuickStart(context),
          
          const SizedBox(height: 32),
          
          _buildSectionHeader(
            context, 
            title: AppLocalizations.of(context, 'solo_recommendations'),
            onSeeAll: () {
              Navigator.push(
                context, 
                // ИСПРАВЛЕНО: Убрали const, чтобы не было ошибки билда
                MaterialPageRoute(builder: (_) => CommunityThemesScreen())
              );
            }
          ),

          const SizedBox(height: 12),

          // ПЕРЕДАЕМ ДАННЫЕ В КАРТОЧКИ
          _buildSoloSlimCard(
            emoji: '🎯', 
            title: 'Job Interview', 
            level: 'C1', 
            tag: 'Professional',
            context: context,
          ),
          _buildSoloSlimCard(
            emoji: '🏨', 
            title: 'Hotel Check-in', 
            level: 'A2', 
            tag: 'Travel',
            context: context,
          ),
          _buildSoloSlimCard(
            emoji: '📉', 
            title: 'Salary Negotiation', 
            level: 'B2', 
            tag: 'Business',
            context: context,
          ),
          _buildSoloSlimCard(
            emoji: '☕', 
            title: 'Small Talk at Café', 
            level: 'A1', 
            tag: 'Casual',
            context: context,
          ),
          
          const SizedBox(height: 30),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(BuildContext context, {required String title, required VoidCallback onSeeAll}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(title, style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18)),
        TextButton(
          onPressed: onSeeAll,
          style: TextButton.styleFrom(visualDensity: VisualDensity.compact),
          child: Text(
            AppLocalizations.of(context, 'see_all'),
            style: const TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold),
          ),
        ),
      ],
    );
  }

  Widget _buildAIQuickStart(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppTheme.primary.withOpacity(0.3), width: 1.5),
        boxShadow: [
          BoxShadow(
            color: AppTheme.primary.withOpacity(0.05), 
            blurRadius: 20, 
            offset: const Offset(0, 4)
          )
        ],
      ),
      child: TextField(
        controller: _aiController,
        style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w500),
        decoration: InputDecoration(
          hintText: AppLocalizations.of(context, 'ai_input_hint'),
          hintStyle: TextStyle(color: Colors.grey.withOpacity(0.6), fontSize: 14),
          prefixIcon: const Icon(Icons.auto_awesome, color: AppTheme.primary, size: 20),
          suffixIcon: Padding(
            padding: const EdgeInsets.only(right: 8),
            child: IconButton(
              icon: const Icon(Icons.play_circle_fill, color: AppTheme.primary, size: 36),
              onPressed: () {
                if (_aiController.text.isNotEmpty) {
                  _startScenario(TopicVote.custom(_aiController.text));
                }
              },
            ),
          ),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(vertical: 18, horizontal: 16),
        ),
      ),
    );
  }

  Widget _buildSoloSlimCard({
    required String emoji, 
    required String title, 
    required String level, 
    required String tag,
    required BuildContext context,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(color: isDark ? Colors.white10 : Colors.black.withOpacity(0.05)),
      ),
      child: ListTile(
        leading: Text(emoji, style: const TextStyle(fontSize: 24)),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
        subtitle: Text(tag, style: const TextStyle(fontSize: 12, color: Colors.grey)),
        trailing: Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          decoration: BoxDecoration(
            color: AppTheme.primarySoft.withOpacity(isDark ? 0.1 : 1.0), 
            borderRadius: BorderRadius.circular(8)
          ),
          child: Text(
            level, 
            style: const TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold, fontSize: 10)
          ),
        ),
        // ВОТ ТУТ ОНТАП: Создаем тему на лету и переходим к практике
        onTap: () => _startScenario(TopicVote(
          id: title.toLowerCase().replaceAll(' ', '_'),
          title: title,
          emoji: emoji,
          level: level,
          duration: '5-10 min',
          skill: tag,
          goal: 'Practice your $tag skills in this scenario.',
          votes: 0,
          voterIds: [],
          publicContext: 'This is a $level level scenario focused on $tag.',
          myRole: 'Student',
          partnerRole: 'AI Tutor',
          aiRoleName: 'AI Teacher',
          aiEmoji: '🤖',
        )),
      ),
    );
  }
}