import 'package:flutter/material.dart';
import '../../data/mock_repository.dart';
import '../../models/voice_room.dart';
import '../../theme/app_theme.dart';
import '../../services/app_localizations.dart';
import 'room_wait_screen.dart';
import 'community_themes_screen.dart';

import '../../services/topic_service.dart';
import '../../models/topic_vote.dart';

class PeopleHomeScreen extends StatefulWidget {
  const PeopleHomeScreen({super.key});

  @override
  State<PeopleHomeScreen> createState() => _PeopleHomeScreenState();
}

class _PeopleHomeScreenState extends State<PeopleHomeScreen> {
  final _search = TextEditingController();
  List<TopicVote> _communityTopics = [];
  bool _loadingTopics = true;

  @override
  void initState() {
    super.initState();
    _fetchCommunityTopics();
  }

  Future<void> _fetchCommunityTopics() async {
    try {
      final data = await TopicService.instance.getCommunityTopics();
      setState(() {
        _communityTopics = data.map((t) => TopicVote(
          id: t['id'],
          title: t['title'],
          emoji: t['emoji'] ?? '🎭',
          level: t['difficulty_level'] ?? 'All',
          duration: '5-10 min',
          skill: 'Community',
          goal: t['goal'] ?? '',
          votes: t['likes_count'] ?? 0,
          voterIds: [],
          publicContext: t['description'] ?? '',
          myRole: t['my_role'] ?? 'User',
          partnerRole: t['partner_role'] ?? 'AI',
          aiRoleName: t['partner_role'] ?? 'AI',
          aiEmoji: t['partner_emoji'] ?? '🤖',
          rating: (t['likes_count'] ?? 0).toDouble(),
          isOfficial: false,
          authorName: 'Community User',
        )).toList();
        _loadingTopics = false;
      });
    } catch (e) {
      debugPrint("Error fetching community topics: $e");
      setState(() => _loadingTopics = false);
    }
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  // Логика фильтрации комнат
  List<VoiceRoom> get _rooms {
    final q = _search.text.trim().toLowerCase();
    final all = MockRepository.voiceRooms;
    if (q.isEmpty) return all;
    return all.where((r) =>
        r.title.toLowerCase().contains(q) ||
        r.subtitle.toLowerCase().contains(q) ||
        r.levelTag.toLowerCase().contains(q)
    ).toList();
  }

  @override
  Widget build(BuildContext context) {
    final rooms = _rooms;
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        centerTitle: false,
        title: Text(
          AppLocalizations.of(context, 'tab_people'),
          style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 24),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.add_box_outlined, color: AppTheme.primary, size: 28),
            onPressed: () => _showCreateOptions(context),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          // 1. ПОИСК
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: TextField(
                controller: _search,
                onChanged: (_) => setState(() {}),
                decoration: InputDecoration(
                  hintText: AppLocalizations.of(context, 'search_hint'),
                  prefixIcon: const Icon(Icons.search_rounded),
                  filled: true,
                  fillColor: theme.cardColor,
                  contentPadding: const EdgeInsets.symmetric(vertical: 0),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: isDark ? Colors.white10 : Colors.black12),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide(color: isDark ? Colors.white10 : Colors.black12),
                  ),
                ),
              ),
            ),
          ),

          // 2. ТОП СЦЕНАРИЕВ СООБЩЕСТВА (UGC + Лайки)
          SliverToBoxAdapter(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildSectionHeader(
                  AppLocalizations.of(context, 'top_community_themes'),
                  showSeeAll: true, // ВКЛЮЧАЕМ КНОПКУ "ВСЕ"
                ),
                SizedBox(
                  height: 160,
                  child: _loadingTopics 
                    ? const Center(child: CircularProgressIndicator())
                    : _communityTopics.isEmpty
                      ? const Center(child: Text("No community topics yet"))
                      : ListView.separated(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        scrollDirection: Axis.horizontal,
                        itemCount: _communityTopics.length, 
                        separatorBuilder: (_, __) => const SizedBox(width: 12),
                        itemBuilder: (context, i) {
                          final topic = _communityTopics[i];
                          return _buildTrendingCard(
                            emoji: topic.emoji,
                            title: topic.title,
                            author: topic.authorName,
                            likes: topic.votes,
                            topic: topic,
                          );
                        },
                      ),
                ),
              ],
            ),
          ),

          // 3. СПИСОК АКТИВНЫХ КОМНАТ
          SliverToBoxAdapter(
            child: _buildSectionHeader(AppLocalizations.of(context, 'active_rooms')),
          ),
          
          if (rooms.isEmpty)
            SliverFillRemaining(
              hasScrollBody: false,
              child: Center(child: Text(AppLocalizations.of(context, 'no_results'))),
            )
          else
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 40),
              sliver: SliverList(
                delegate: SliverChildBuilderDelegate(
                  (context, i) => _buildLiveRoomSlim(context, rooms[i]),
                  childCount: rooms.length,
                ),
              ),
            ),
        ],
      ),
    );
  }

  // ОБНОВЛЕННЫЙ ЗАГОЛОВОК С КНОПКОЙ "ВСЕ"
Widget _buildSectionHeader(String title, {bool showSeeAll = false}) {
  return Padding(
    padding: const EdgeInsets.fromLTRB(16, 24, 16, 12),
    child: Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(title, style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 18)),
        if (showSeeAll)
          GestureDetector(
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => CommunityThemesScreen())),
            child: Text(
              AppLocalizations.of(context, 'see_all'),
              style: const TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold, fontSize: 13),
            ),
          ),
      ],
    ),
  );
}

  // КАРТОЧКА С ЛАЙКАМИ (Твоя гениальная идея)
  Widget _buildTrendingCard({
    required String emoji, 
    required String title, 
    required String author, 
    required int likes,
    required TopicVote topic,
  }) {
    final theme = Theme.of(context);
    return GestureDetector(
      onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => CommunityThemesScreen())),
      child: Container(
        width: 150,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: theme.cardColor,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: theme.dividerColor.withOpacity(0.08)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: AppTheme.primary.withOpacity(0.1), 
                shape: BoxShape.circle
              ),
              child: Text(emoji, style: const TextStyle(fontSize: 22)),
            ),
            const Spacer(),
            Text(
              title, 
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, height: 1.1), 
              maxLines: 1, 
              overflow: TextOverflow.ellipsis
            ),
            Text(author, style: const TextStyle(fontSize: 10, color: Colors.grey)),
            const SizedBox(height: 10),
            Row(
              children: [
                const Icon(Icons.favorite, color: Colors.redAccent, size: 14),
                const SizedBox(width: 4),
                Text(
                  likes.toString(), 
                  style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w900)
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  // ТОНКАЯ КАРТОЧКА КОМНАТЫ (Slim Design)
  Widget _buildLiveRoomSlim(BuildContext context, VoiceRoom room) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 10),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(color: isDark ? Colors.white10 : Colors.black.withOpacity(0.05)),
      ),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        leading: Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: room.accent.withOpacity(0.12),
            borderRadius: BorderRadius.circular(12),
          ),
          alignment: Alignment.center,
          child: Text(room.emoji, style: const TextStyle(fontSize: 22)),
        ),
        title: Text(
          room.title, 
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15, height: 1.2)
        ),
        subtitle: Text(
          '${room.onlineCount} ${AppLocalizations.of(context, 'online')} • ${room.levelTag}', 
          style: const TextStyle(fontSize: 11)
        ),
        trailing: const Icon(Icons.arrow_forward_ios_rounded, size: 14, color: Colors.grey),
        onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => RoomWaitScreen(room: room))),
      ),
    );
  }

  void _showCreateOptions(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 12),
            Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2))),
            ListTile(
              leading: const Icon(Icons.add_to_photos_rounded, color: AppTheme.primary),
              title: const Text('Опубликовать сценарий', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: const Text('Ваша идея попадет в ленту сообщества'),
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.groups_rounded, color: AppTheme.primary),
              title: const Text('Создать комнату', style: TextStyle(fontWeight: FontWeight.bold)),
              subtitle: const Text('Мгновенное лобби для общения с людьми'),
              onTap: () => Navigator.pop(context),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }
}