import 'package:flutter/material.dart';
import '../services/auth_service.dart';
import '../services/settings_store.dart';
import '../services/app_localizations.dart';
import '../theme/app_theme.dart';
import 'friends_screen.dart';
import 'history_screen.dart';
import 'skills_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key, required this.onLogout});
  final VoidCallback onLogout;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  void _showEditNameDialog(String currentName) {
    final controller = TextEditingController(text: currentName);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(AppLocalizations.of(context, 'edit_name')),
        content: TextField(
          controller: controller, 
          autofocus: true,
          decoration: InputDecoration(hintText: AppLocalizations.of(context, 'name_hint'))
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: Text(AppLocalizations.of(context, 'cancel'))),
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(AppLocalizations.of(context, 'save'), style: const TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final u = AuthService.instance.currentUser;
    final store = SettingsStore.instance;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: Text(AppLocalizations.of(context, 'tab_profile'))),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
        children: [
          _buildHeader(u),
          const SizedBox(height: 32),
          
          // РАЗДЕЛ: ОБУЧЕНИЕ (Новый)
          _sectionLabel(AppLocalizations.of(context, 'practice_settings')),
          _buildMenuCard(
            icon: Icons.insights_rounded,
            title: AppLocalizations.of(context, 'skills_title'),
            subtitle: AppLocalizations.of(context, 'skills_subtitle'),
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const SkillsScreen())),
            isHighlight: true,
          ),
          
          // Выбор языка практики (переехал сюда)
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.school_outlined, color: AppTheme.primary),
            title: Text(AppLocalizations.of(context, 'practice_lang')),
            trailing: DropdownButton<String>(
              value: store.practiceLanguage,
              underline: const SizedBox(),
              onChanged: (v) => store.setPracticeLanguage(v!),
              items: ['English', 'Spanish', 'French', 'German', 'Russian']
                  .map((e) => DropdownMenuItem(value: e, child: Text(e, style: const TextStyle(fontWeight: FontWeight.bold))))
                  .toList(),
            ),
          ),

          const SizedBox(height: 24),
          _sectionLabel(AppLocalizations.of(context, 'appearance')),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _themeBtn(0, Icons.wb_sunny_outlined, AppLocalizations.of(context, 'theme_light'), store.themeIndex),
              _themeBtn(1, Icons.nightlight_round, AppLocalizations.of(context, 'theme_dark'), store.themeIndex),
              _themeBtn(2, Icons.auto_awesome, AppLocalizations.of(context, 'theme_moon'), store.themeIndex),
            ],
          ),

          const SizedBox(height: 24),
          _sectionLabel(AppLocalizations.of(context, 'language')),
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: const Icon(Icons.language, color: AppTheme.primary),
            title: Text(AppLocalizations.of(context, 'app_language')),
            trailing: DropdownButton<String>(
              value: store.languageCode.toUpperCase(),
              underline: const SizedBox(),
              onChanged: (v) => store.setLanguage(v!.toLowerCase()),
              items: ['RU', 'EN'].map((e) => DropdownMenuItem(value: e, child: Text(e, style: const TextStyle(fontWeight: FontWeight.bold)))).toList(),
            ),
          ),

          const Divider(height: 40),
          _sectionLabel(AppLocalizations.of(context, 'notifications')),
          _buildSwitchTile(AppLocalizations.of(context, 'push_notify'), store.notifyLobbyReady, store.setNotifyLobbyReady),
          _buildSwitchTile(AppLocalizations.of(context, 'vibration'), store.vibrateOnReady, store.setVibrateOnReady),
          
          _buildMenuCard(
            icon: Icons.people_outline,
            title: AppLocalizations.of(context, 'friends_title'),
            subtitle: AppLocalizations.of(context, 'friends_subtitle'),
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const FriendsScreen())),
          ),
          _buildMenuCard(
            icon: Icons.history,
            title: AppLocalizations.of(context, 'history_title'),
            subtitle: AppLocalizations.of(context, 'history_subtitle'),
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const HistoryScreen())),
          ),

          const SizedBox(height: 32),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(
              onPressed: widget.onLogout,
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.red,
                side: const BorderSide(color: Colors.red),
                padding: const EdgeInsets.symmetric(vertical: 15)
              ),
              child: Text(AppLocalizations.of(context, 'logout'), style: const TextStyle(fontWeight: FontWeight.bold)),
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  // ... (остальные вспомогательные методы: _buildHeader, _sectionLabel, _themeBtn, _buildMenuCard, _buildSwitchTile как в прошлом коде)
  
  Widget _buildHeader(u) {
    return Row(
      children: [
        Stack(
          children: [
            CircleAvatar(
              radius: 35,
              backgroundColor: AppTheme.primarySoft,
              child: Text(
                u?.displayName?[0].toUpperCase() ?? '?', 
                style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.primary)
              ),
            ),
            Positioned(
              bottom: 0,
              right: 0,
              child: GestureDetector(
                onTap: () {}, 
                child: const CircleAvatar(radius: 12, backgroundColor: AppTheme.primary, child: Icon(Icons.camera_alt, size: 12, color: Colors.white)),
              ),
            ),
          ],
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              GestureDetector(
                onTap: () => _showEditNameDialog(u?.displayName ?? ''),
                child: Row(
                  children: [
                    Flexible(
                      child: Text(u?.displayName ?? 'User', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800, height: 1.2)),
                    ),
                    const SizedBox(width: 8),
                    const Icon(Icons.edit, size: 16, color: AppTheme.primary),
                  ],
                ),
              ),
              Text(u?.email ?? '', style: const TextStyle(color: AppTheme.textSecondary, fontSize: 13, height: 1.2)),
            ],
          ),
        ),
      ],
    );
  }

  Widget _sectionLabel(String t) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Text(t.toUpperCase(), style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w900, color: AppTheme.textSecondary, letterSpacing: 0.5)),
  );

  Widget _buildSwitchTile(String title, bool value, Function(bool) onChanged) {
    return SwitchListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      value: value,
      onChanged: onChanged,
    );
  }

  Widget _themeBtn(int idx, IconData icon, String label, int currentIdx) {
    bool isSel = currentIdx == idx;
    bool isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: () => SettingsStore.instance.setThemeIndex(idx),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.28,
        padding: const EdgeInsets.symmetric(vertical: 12),
        decoration: BoxDecoration(
          color: isSel ? AppTheme.primary : (isDark ? Colors.white.withOpacity(0.05) : AppTheme.primarySoft),
          borderRadius: BorderRadius.circular(12),
          border: isSel ? Border.all(color: AppTheme.primary, width: 2) : Border.all(color: Colors.transparent, width: 2),
        ),
        child: Column(
          children: [
            Icon(icon, color: isSel ? Colors.white : AppTheme.primary, size: 20),
            const SizedBox(height: 6),
            Text(label, style: TextStyle(color: isSel ? Colors.white : AppTheme.primary, fontSize: 11, fontWeight: FontWeight.w800, height: 1.2)),
          ],
        ),
      ),
    );
  }

  Widget _buildMenuCard({required IconData icon, required String title, required String subtitle, required VoidCallback onTap, bool isHighlight = false}) {
    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 12),
      color: isHighlight ? AppTheme.primary.withOpacity(0.05) : Theme.of(context).cardColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16), 
        side: BorderSide(color: isHighlight ? AppTheme.primary : Colors.grey.withOpacity(0.15))
      ),
      child: ListTile(
        leading: Icon(icon, color: AppTheme.primary),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w700)),
        subtitle: Text(subtitle, style: const TextStyle(fontSize: 12)),
        trailing: const Icon(Icons.chevron_right, size: 20),
        onTap: onTap,
      ),
    );
  }
}