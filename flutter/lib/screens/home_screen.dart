import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/app_localizations.dart';
import '../models/topic_vote.dart';
import '../services/active_session_manager.dart';
import 'prep_screen.dart';
import 'multi/community_themes_screen.dart';
import '../services/topic_service.dart';
import 'session_screen.dart';

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
    ActiveSessionManager.instance.addListener(_rebuild);
  }

  @override
  void dispose() {
    ActiveSessionManager.instance.removeListener(_rebuild);
    _aiController.dispose();
    super.dispose();
  }

  void _rebuild() => setState(() {});

  Future<void> _fetchTopics() async {
    try {
      final topicsData = await TopicService.instance.getOfficialTopics();
      if (!mounted) return;
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
      if (mounted) setState(() => _loading = false);
    }
  }

  void _startScenario(TopicVote topic) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => PrepScreen(topic: topic)),
    );
  }

  void _resumeSession(TopicVote topic) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => SessionScreen(topic: topic)),
    );
  }

  // --- МЕТОД ДЛЯ НАСТРОЙКИ КАСТОМНОЙ КОМНАТЫ С ИИ ---
  void _showConfigureAIChatSheet(BuildContext context, String initialPrompt) {
    String selectedLevel = 'B1';
    String selectedEmoji = '🎭';
    final _titleController = TextEditingController(text: "My Custom Scenario");
    final _promptController = TextEditingController(text: initialPrompt);

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return StatefulBuilder(
            builder: (context, setModalState) {
              final isDark = Theme.of(context).brightness == Brightness.dark;

              return Container(
                decoration: BoxDecoration(
                  color: Theme.of(context).scaffoldBackgroundColor,
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
                ),
                padding: EdgeInsets.only(
                  left: 24, right: 24, top: 24,
                  bottom: MediaQuery.of(context).viewInsets.bottom + 24,
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Center(
                      child: Container(
                        width: 40, height: 4,
                        decoration: BoxDecoration(color: Colors.grey.withOpacity(0.3), borderRadius: BorderRadius.circular(2)),
                      ),
                    ),
                    const SizedBox(height: 24),
                    const Text(
                      'Configure AI Room',
                      style: TextStyle(fontWeight: FontWeight.w900, fontSize: 22),
                    ),
                    const SizedBox(height: 16),

                    // Название комнаты
                    TextField(
                      controller: _titleController,
                      decoration: const InputDecoration(
                        labelText: 'Topic Title',
                        border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(12))),
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Промпт для ИИ
                    TextField(
                      controller: _promptController,
                      maxLines: 3,
                      decoration: const InputDecoration(
                        labelText: 'What do you want to practice?',
                        border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(12))),
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Выбор сложности
                    const Text('Difficulty Level', style: TextStyle(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 8),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: ['A1', 'A2', 'B1', 'B2', 'C1'].map((level) {
                        final isSelected = selectedLevel == level;
                        return ChoiceChip(
                          label: Text(level),
                          selected: isSelected,
                          onSelected: (_) => setModalState(() => selectedLevel = level),
                          selectedColor: AppTheme.primary,
                          // Вместо textColor используем labelStyle
                          labelStyle: TextStyle(
                            color: isSelected ? Colors.white : (isDark ? Colors.white70 : Colors.black87),
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                          ),
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 24),

                    // Кнопка запуска
                    SizedBox(
                      width: double.infinity,
                      height: 56,
                      child: ElevatedButton(
                        onPressed: () {
                          Navigator.pop(context); // Закрываем шторку

                          // Создаем объект для PrepScreen
                          final customTopic = TopicVote(
                            id: 'custom_${DateTime.now().millisecondsSinceEpoch}',
                            title: _titleController.text,
                            emoji: selectedEmoji,
                            level: selectedLevel,
                            duration: '5-10 min',
                            skill: 'Conversation',
                            goal: _promptController.text,
                            votes: 0,
                            voterIds: [],
                            publicContext: _promptController.text,
                            myRole: 'Student',
                            partnerRole: 'AI Assistant',
                            aiRoleName: 'AI Assistant',
                            aiEmoji: '🤖',
                            isOfficial: false,
                          );

                          // Отправляем на экран подготовки
                          _startScenario(customTopic);
                        },
                        style: ElevatedButton.styleFrom(
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                          backgroundColor: AppTheme.primary,
                          foregroundColor: Colors.white,
                        ),
                        child: const Text('Start Practice', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                      ),
                    ),
                  ],
                ),
              );
            }
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final unfinished = ActiveSessionManager.instance.unfinishedSessions;

    return Scaffold(
      appBar: AppBar(
        centerTitle: false,
        title: Text(
          AppLocalizations.of(context, 'ai_coach'),
          style: const TextStyle(
            fontWeight: FontWeight.w900,
            fontSize: 24,
          ),
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        children: [
          const SizedBox(height: 12),
          _buildAIQuickStart(context),

          if (unfinished.isNotEmpty) ...[
            const SizedBox(height: 32),
            _buildSectionHeader(
              context,
              title: 'Unfinished Dialogues',
              onSeeAll: () {},
            ),
            const SizedBox(height: 12),
            SizedBox(
              height: 140,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                physics: const BouncingScrollPhysics(),
                itemCount: unfinished.length,
                itemBuilder: (context, i) => _buildUnfinishedCard(unfinished[i]),
              ),
            ),
          ],

          const SizedBox(height: 32),
          _buildSectionHeader(
              context,
              title: AppLocalizations.of(context, 'solo_recommendations'),
              onSeeAll: () {
                Navigator.push(context, MaterialPageRoute(builder: (_) => const CommunityThemesScreen()));
              }
          ),

          const SizedBox(height: 12),

          if (_loading)
            const Center(
              child: Padding(
                padding: EdgeInsets.all(32.0),
                child: CircularProgressIndicator(),
              ),
            )
          else if (_topics.isEmpty)
            Center(
              child: Padding(
                padding: const EdgeInsets.all(40.0),
                child: Text(
                  AppLocalizations.of(context, 'no_scenarios'),
                  style: TextStyle(color: theme.hintColor),
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
            )).toList(),

          const SizedBox(height: 30),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(
      BuildContext context, {
        required String title,
        required VoidCallback onSeeAll,
      }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          title,
          style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18),
        ),
        if (title != 'Unfinished Dialogues')
          TextButton(
            onPressed: onSeeAll,
            child: Text(
              AppLocalizations.of(context, 'see_all'),
              style: const TextStyle(
                color: AppTheme.primary,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildUnfinishedCard(TopicVote topic) {
    return GestureDetector(
      onTap: () => _resumeSession(topic),
      child: Container(
        width: 160,
        margin: const EdgeInsets.only(right: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppTheme.primary.withOpacity(0.05),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: AppTheme.primary.withOpacity(0.2)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(topic.emoji, style: const TextStyle(fontSize: 28)),
            const Spacer(),
            Text(
                topic.title,
                style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 14),
                maxLines: 2,
                overflow: TextOverflow.ellipsis
            ),
            const SizedBox(height: 4),
            const Text(
                'Resume session',
                style: TextStyle(color: AppTheme.primary, fontSize: 10, fontWeight: FontWeight.bold)
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAIQuickStart(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppTheme.primary.withOpacity(0.3), width: 1.5),
      ),
      child: TextField(
        controller: _aiController,
        decoration: InputDecoration(
          hintText: AppLocalizations.of(context, 'ai_input_hint'),
          prefixIcon: const Icon(Icons.auto_awesome, color: AppTheme.primary, size: 20),
          suffixIcon: IconButton(
            icon: const Icon(Icons.play_circle_fill, color: AppTheme.primary, size: 36),
            onPressed: () {
              if (_aiController.text.isNotEmpty) {
                // ВМЕСТО перехода сразу на экран, открываем шторку настроек
                _showConfigureAIChatSheet(context, _aiController.text);
              }
            },
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
              color: AppTheme.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8)
          ),
          child: Text(
              level,
              style: const TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold, fontSize: 10)
          ),
        ),
        onTap: () => _startScenario(topic),
      ),
    );
  }
}