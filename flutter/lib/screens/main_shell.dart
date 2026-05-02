import 'package:flutter/material.dart';
import '../services/app_localizations.dart';
import '../models/topic_vote.dart';
import '../services/active_session_manager.dart';
import '../widgets/session_widgets.dart';
import 'home_screen.dart';
import 'profile_screen.dart';
import 'multi/people_home_screen.dart';
import 'session_screen.dart';
import 'multi/human_session_screen.dart';

class MainShell extends StatefulWidget {
  const MainShell({super.key, required this.onLogout});
  final VoidCallback onLogout;

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _tab = 0;

  @override
  void initState() {
    super.initState();
    ActiveSessionManager.instance.addListener(_rebuild);
  }

  @override
  void dispose() {
    ActiveSessionManager.instance.removeListener(_rebuild);
    super.dispose();
  }

  void _rebuild() => setState(() {});

  void _reEnterSession() {
    final manager = ActiveSessionManager.instance;
    if (manager.activeRoom != null) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => HumanSessionScreen(
            topic: TopicVote(
              id: manager.activeRoom!.id,
              title: manager.activeRoom!.title,
              goal: manager.activeRoom!.subtitle,
              myRole: 'User',
              partnerRole: 'AI',
              emoji: manager.activeRoom!.emoji,
              level: manager.activeRoom!.levelTag,
              duration: '',
              skill: '',
              votes: 0,
              voterIds: [],
              publicContext: '',
              aiRoleName: '',
              aiEmoji: '🤖',
            ),
            playerRoles: const {'You': 'Host'}, // Mock roles for re-entry
            room: manager.activeRoom,
          ),
        ),
      );
    } else if (manager.activeTopic != null) {
      Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => SessionScreen(topic: manager.activeTopic!)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final manager = ActiveSessionManager.instance;

    return Scaffold(
      body: Stack(
        children: [
          IndexedStack(
            index: _tab,
            children: [
              const HomeScreen(),
              const PeopleHomeScreen(),
              ProfileScreen(onLogout: widget.onLogout),
            ],
          ),
          if (manager.isActive)
            Positioned(
              bottom: 0,
              left: 0,
              right: 0,
              child: FloatingSessionBar(
                title: manager.activeRoom?.title ?? manager.activeTopic?.title ?? 'Session',
                onTap: _reEnterSession,
                onClose: () => manager.closeSession(),
              ),
            ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _tab,
        onDestinationSelected: (i) => setState(() => _tab = i),
        height: 70,
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.psychology_outlined), 
            selectedIcon: const Icon(Icons.psychology), 
            label: AppLocalizations.of(context, 'tab_ai'),
          ),
          NavigationDestination(
            icon: const Icon(Icons.groups_outlined), 
            selectedIcon: const Icon(Icons.groups), 
            label: AppLocalizations.of(context, 'tab_people'),
          ),
          NavigationDestination(
            icon: const Icon(Icons.person_outline), 
            selectedIcon: const Icon(Icons.person), 
            label: AppLocalizations.of(context, 'tab_profile'),
          ),
        ],
      ),
    );
  }
}