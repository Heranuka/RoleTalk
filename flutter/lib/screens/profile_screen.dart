import 'package:flutter/material.dart';
import '../services/auth_service.dart';
import '../services/settings_store.dart';
import '../services/app_localizations.dart';
import '../theme/app_theme.dart';
import 'friends_screen.dart';
import 'history_screen.dart';
import 'skills_screen.dart';
import 'auth_gate.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key, required this.onLogout});
  final VoidCallback onLogout;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  bool _isVerifying = false;
  bool _isVerificationSent = false;

  // Главный метод обработки верификации
  Future<void> _handleVerificationAction() async {
    final u = AuthService.instance.currentUser;
    if (u == null) return;

    setState(() => _isVerifying = true);

    try {
      if (!_isVerificationSent) {
        // ШАГ 1: Отправляем письмо
        debugPrint("Sending verification email to: ${u.email}");
        await AuthService.instance.resendVerification(u.email);

        if (mounted) {
          setState(() {
            _isVerificationSent = true;
            _isVerifying = false;
          });
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Verification link sent! Check your email.')),
          );
        }
      } else {
        // ШАГ 2: Обновляем профиль (проверяем статус в БД)
        debugPrint("Refreshing profile to check verification status...");
        await AuthService.instance.refreshProfile();

        if (mounted) {
          setState(() {
            _isVerifying = false;
            // Если после refreshProfile() поле isVerified стало true,
            // баннер исчезнет автоматически при перерисовке.
          });

          final updatedUser = AuthService.instance.currentUser;
          if (updatedUser?.isVerified ?? false) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Email verified! Banner will now disappear.'), backgroundColor: Colors.green),
            );
          } else {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Not verified yet. Did you click the link in your email?')),
            );
          }
        }
      }
    } catch (e) {
      debugPrint("Verification action error: $e");
      if (mounted) {
        setState(() => _isVerifying = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.redAccent),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final u = AuthService.instance.currentUser;
    final store = SettingsStore.instance;
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    // ПРИНТ ДЛЯ ОТЛАДКИ (удали потом)
    debugPrint("UI Build: user verified = ${u?.isVerified}");

    return Scaffold(
      appBar: AppBar(title: Text(AppLocalizations.of(context, 'tab_profile'))),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
        children: [
          _buildHeader(u),
          const SizedBox(height: 24),

          // --- ДИНАМИЧНЫЙ БАННЕР ВЕРИФИКАЦИИ ---
          if (u != null && !u.isVerified)
            AnimatedContainer(
              duration: const Duration(milliseconds: 400),
              margin: const EdgeInsets.only(bottom: 24),
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: _isVerificationSent ? Colors.blue.withOpacity(0.05) : Colors.orange.withOpacity(0.08),
                borderRadius: BorderRadius.circular(28),
                border: Border.all(
                  color: _isVerificationSent ? Colors.blue.withOpacity(0.2) : Colors.orange.withOpacity(0.2),
                  width: 1.5,
                ),
              ),
              child: Column(
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: _isVerificationSent ? Colors.blue.withOpacity(0.1) : Colors.orange.withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          _isVerificationSent ? Icons.mark_email_read_outlined : Icons.mail_lock_outlined,
                          color: _isVerificationSent ? Colors.blue : Colors.orange,
                          size: 24,
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _isVerificationSent ? 'Check your inbox' : 'Verify your email',
                              style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 16),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _isVerificationSent
                                  ? 'Click refresh after confirming in email.'
                                  : 'Full access requires a verified account.',
                              style: TextStyle(fontSize: 13, color: theme.hintColor, height: 1.2),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _isVerificationSent ? Colors.blue : Colors.orange,
                        foregroundColor: Colors.white,
                        elevation: 0,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                      ),
                      onPressed: _isVerifying ? null : _handleVerificationAction,
                      child: _isVerifying
                          ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                          : Text(
                        _isVerificationSent ? 'REFRESH / CHECK STATUS' : 'SEND VERIFICATION EMAIL',
                        style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 12, letterSpacing: 0.5),
                      ),
                    ),
                  ),
                ],
              ),
            ),

          // --- РАЗДЕЛ: ОБУЧЕНИЕ ---
          _sectionLabel(AppLocalizations.of(context, 'practice_settings')),
          _buildMenuCard(
            icon: Icons.insights_rounded,
            title: AppLocalizations.of(context, 'skills_title'),
            subtitle: AppLocalizations.of(context, 'skills_subtitle'),
            onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const SkillsScreen())),
            isHighlight: true,
          ),

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
              onPressed: () async {
                await AuthService.instance.logout();
                if (!mounted) return;
                widget.onLogout();
                Navigator.of(context, rootNavigator: true).pushAndRemoveUntil(
                  MaterialPageRoute(builder: (context) => const AuthGate()),
                      (route) => false,
                );
              },
              style: OutlinedButton.styleFrom(
                  foregroundColor: Colors.red,
                  side: const BorderSide(color: Colors.red),
                  padding: const EdgeInsets.symmetric(vertical: 15)
              ),
              child: Text(
                  AppLocalizations.of(context, 'logout'),
                  style: const TextStyle(fontWeight: FontWeight.bold)),
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  // --- ВСПОМОГАТЕЛЬНЫЕ ВИДЖЕТЫ ---

  void _showEditNameDialog(String currentName) {
    final controller = TextEditingController(text: currentName);
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Edit display name'),
        content: TextField(
            controller: controller,
            autofocus: true,
            decoration: const InputDecoration(hintText: 'Enter name')
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Save', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(u) {
    return Row(
      children: [
        CircleAvatar(
          radius: 35,
          backgroundColor: AppTheme.primarySoft,
          child: Text(
              u?.displayName?[0].toUpperCase() ?? '?',
              style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.primary)
          ),
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