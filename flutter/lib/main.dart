import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart'; // Добавь этот импорт
import 'screens/auth_gate.dart';
import 'services/local_notification_service.dart';
import 'services/settings_store.dart';
import 'services/app_localizations.dart';
import 'theme/app_theme.dart';
import 'widgets/phone_shell.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализируем уведомления
  await LocalNotificationService.init();

  // --- КОД ДЛЯ ТЕСТИРОВАНИЯ (СБРОС СЕССИИ) ---
  // Раскомментируй следующие две строки, чтобы принудительно выйти из аккаунта и увидеть Login/Register:
  // final prefs = await SharedPreferences.getInstance();
  // await prefs.clear();
  // ------------------------------------------

  // 1. Инициализируем настройки
  await SettingsStore.instance.init();

  // 2. Загружаем локализацию
  await AppLocalizations.load(SettingsStore.instance.languageCode);

  runApp(const SpeakSimApp());
}

class SpeakSimApp extends StatelessWidget {
  const SpeakSimApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: SettingsStore.instance,
      builder: (context, child) {
        final tIdx = SettingsStore.instance.themeIndex;
        final langCode = SettingsStore.instance.languageCode;

        return MaterialApp(
          title: 'RoleTalk',
          debugShowCheckedModeBanner: false,
          locale: Locale(langCode),
          theme: tIdx == 0
              ? AppTheme.light()
              : tIdx == 1
              ? AppTheme.dark()
              : AppTheme.moon(),
          // PhoneShell ограничивает ширину, AuthGate управляет логикой входа
          home: const PhoneShell(
            child: AuthGate(),
          ),
        );
      },
    );
  }
}