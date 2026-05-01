import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/app_localizations.dart';
import '../models/topic_vote.dart';
import 'prep_screen.dart';
import 'multi/community_themes_screen.dart';
import '../services/topic_service.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final _aiController = TextEditingController();
  List<TopicVote> _topics = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _fetchTopics();
  }

  Future<void> _fetchTopics() async {
    try {
      final topicsData = await TopicService.instance.getOfficialTopics();
      setState(() {
        _topics = topicsData.map((t) => TopicVote(
          id: t['id'],
          title: t['title'],
          emoji: t['emoji'] ?? '🎯',
          level: t['difficulty_level'] ?? 'A2',
          duration: '5-10 min',
          skill: 'Conversation',
          goal: t['goal'] ?? '',
          votes: 0,
          voterIds: [],
          publicContext: t['description'] ?? '',
          myRole: t['my_role'] ?? 'Student',
          partnerRole: t['partner_role'] ?? 'AI',
          aiRoleName: t['partner_role'] ?? 'AI',
          aiEmoji: t['partner_emoji'] ?? '🤖',
        )).toList();
        _loading = false;
      });
    } catch (e) {
      debugPrint("Error fetching topics: $e");
      setState(() => _loading = false);
    }
  }

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
                MaterialPageRoute(builder: (_) => const CommunityThemesScreen())
              );
            }
          ),

          const SizedBox(height: 12),

          if (_loading)
            const Center(child: CircularProgressIndicator())
          else if (_topics.isEmpty)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(20.0),
                child: Text(
                  AppLocalizations.of(context, 'no_topics_found') ?? 'No scenarios available.',
                  style: const TextStyle(color: Colors.grey),
                ),
              ),
            )
          else
            ..._topics.map((topic) => _buildSoloSlimCard(
              emoji: topic.emoji,
              title: topic.title,
              level: topic.level,
              tag: topic.skill,
              context: context,
              topic: topic,
            )),
          
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
    required TopicVote topic,
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
        onTap: () => _startScenario(topic),
      ),
    );
  }
}