import 'package:flutter/material.dart';
import '../../data/mock_repository.dart';
import '../../models/voice_room.dart';
import '../../theme/app_theme.dart';
import '../../services/app_localizations.dart';
import 'room_wait_screen.dart';
import 'community_themes_screen.dart';
import 'create_room_screen.dart';

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
      if (!mounted) return;
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
      if (mounted) setState(() => _loadingTopics = false);
    }
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

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
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: isDark
                ? [const Color(0xFF1A1A1A), const Color(0xFF121212)]
                : [const Color(0xFFF9FAFB), const Color(0xFFF3F4F6)],
          ),
        ),
        child: CustomScrollView(
          physics: const BouncingScrollPhysics(),
          slivers: [
            // 1. HEADER
            SliverAppBar(
              expandedHeight: 120,
              floating: true,
              pinned: true,
              stretch: true,
              backgroundColor: Colors.transparent,
              elevation: 0,
              flexibleSpace: FlexibleSpaceBar(
                centerTitle: false,
                titlePadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                title: Text(
                  AppLocalizations.of(context, 'tab_people'),
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontSize: 28,
                    color: isDark ? Colors.white : AppTheme.textPrimary,
                  ),
                ),
              ),
              actions: [
                Padding(
                  padding: const EdgeInsets.only(right: 12),
                  child: Container(
                    decoration: BoxDecoration(
                      color: AppTheme.primary.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: IconButton(
                      icon: const Icon(Icons.add_rounded, color: AppTheme.primary, size: 28),
                      onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const CreateRoomScreen())),
                    ),
                  ),
                ),
              ],
            ),

            // 2. TRENDING TOPICS
            SliverToBoxAdapter(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    context,
                    title: AppLocalizations.of(context, 'top_community_themes'),
                    showSeeAll: true,
                  ),
                  const SizedBox(height: 12),
                  SizedBox(
                    height: 180,
                    child: _loadingTopics
                        ? const Center(child: CircularProgressIndicator())
                        : _communityTopics.isEmpty
                        ? Center(
                      child: Text(
                        AppLocalizations.of(context, 'no_community_topics'),
                        style: TextStyle(color: theme.hintColor, fontWeight: FontWeight.w600),
                      ),
                    )
                        : ListView.separated(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      scrollDirection: Axis.horizontal,
                      physics: const BouncingScrollPhysics(),
                      itemCount: _communityTopics.length,
                      separatorBuilder: (_, __) => const SizedBox(width: 16),
                      itemBuilder: (context, i) => _buildTrendingCard(topic: _communityTopics[i]),
                    ),
                  ),
                ],
              ),
            ),

            // 3. ACTIVE ROOMS HEADER
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.only(top: 24.0),
                child: _buildSectionHeader(
                  context,
                  title: AppLocalizations.of(context, 'active_rooms'), // Именованный title
                ),
              ),
            ),

            // 4. ROOMS LIST OR EMPTY STATE
            if (rooms.isEmpty)
              SliverFillRemaining(
                hasScrollBody: false,
                child: Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.people_outline_rounded,
                        size: 64,
                        color: AppTheme.textSecondary.withOpacity(0.3),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        AppLocalizations.of(context, 'no_results'),
                        style: const TextStyle(color: AppTheme.textSecondary),
                      ),
                    ],
                  ),
                ),
              )
            else
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 12, 20, 40),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                        (context, i) => _buildLiveRoomPremium(context, rooms[i]),
                    childCount: rooms.length,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // --- HELPER WIDGETS ---

  Widget _buildSectionHeader(
      BuildContext context, {
        required String title, // Теперь это именованный параметр
        bool showSeeAll = false,
      }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontWeight: FontWeight.w900,
              fontSize: 18,
              letterSpacing: -0.5,
            ),
          ),
          if (showSeeAll)
            TextButton(
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) => const CommunityThemesScreen(),
                  ),
                );
              },
              style: TextButton.styleFrom(foregroundColor: AppTheme.primary),
              child: Text(
                AppLocalizations.of(context, 'see_all'),
                style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildTrendingCard({required TopicVote topic}) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Container(
      width: 160,
      decoration: BoxDecoration(
        color: theme.cardColor,
        borderRadius: BorderRadius.circular(24),
        boxShadow: AppTheme.premiumShadow,
        border: Border.all(color: isDark ? Colors.white10 : Colors.black.withOpacity(0.03)),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(24),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const CommunityThemesScreen())),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: AppTheme.primary.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Text(topic.emoji, style: const TextStyle(fontSize: 24)),
                  ),
                  const Spacer(),
                  Text(
                      topic.title,
                      style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 14, height: 1.2),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis
                  ),
                  const SizedBox(height: 4),
                  Text(topic.authorName, style: const TextStyle(fontSize: 11, color: AppTheme.textSecondary)),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      const Icon(Icons.favorite_rounded, color: Colors.pinkAccent, size: 14),
                      const SizedBox(width: 4),
                      Text(
                          topic.votes.toString(),
                          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w900)
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildLiveRoomPremium(BuildContext context, VoiceRoom room) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Container(
        decoration: BoxDecoration(
          color: Theme.of(context).cardColor,
          borderRadius: BorderRadius.circular(24),
          boxShadow: AppTheme.premiumShadow,
          border: Border.all(color: isDark ? Colors.white10 : Colors.black.withOpacity(0.03)),
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(24),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => RoomWaitScreen(room: room))),
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Container(
                      width: 56,
                      height: 56,
                      decoration: BoxDecoration(
                        color: room.accent.withOpacity(0.12),
                        borderRadius: BorderRadius.circular(18),
                      ),
                      alignment: Alignment.center,
                      child: Text(room.emoji, style: const TextStyle(fontSize: 28)),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                              room.title,
                              style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16, height: 1.2)
                          ),
                          const SizedBox(height: 4),
                          Row(
                            children: [
                              Container(
                                width: 8,
                                height: 8,
                                decoration: const BoxDecoration(
                                  color: AppTheme.primary,
                                  shape: BoxShape.circle,
                                ),
                              ),
                              const SizedBox(width: 6),
                              Text(
                                  '${room.onlineCount} ${AppLocalizations.of(context, 'online')} • ${room.levelTag}',
                                  style: const TextStyle(fontSize: 12, color: AppTheme.textSecondary, fontWeight: FontWeight.w600)
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    Icon(Icons.chevron_right_rounded, color: AppTheme.textSecondary.withOpacity(0.5)),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}