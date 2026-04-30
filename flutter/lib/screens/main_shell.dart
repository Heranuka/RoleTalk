import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/app_localizations.dart';
import 'home_screen.dart';
import 'profile_screen.dart';
import 'multi/people_home_screen.dart';

class MainShell extends StatefulWidget {
  const MainShell({super.key, required this.onLogout});
  final VoidCallback onLogout;

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _tab,
        children: [
          const HomeScreen(),
          const PeopleHomeScreen(),
          ProfileScreen(onLogout: widget.onLogout),
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