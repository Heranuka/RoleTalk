import 'package:flutter/material.dart';
import 'screens/auth_gate.dart';
import 'services/local_notification_service.dart';
import 'services/settings_store.dart';
import 'services/app_localizations.dart'; // ВАЖНО: Добавь этот импорт
import 'theme/app_theme.dart';
import 'widgets/phone_shell.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await LocalNotificationService.init();
  
  // 1. Инициализируем стор (загружаем настройки темы и языка)
  await SettingsStore.instance.init();
  
  // 2. Загружаем JSON файл перевода
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
          // Передаем локаль в приложение
          locale: Locale(langCode),
          theme: tIdx == 0 
              ? AppTheme.light() 
              : tIdx == 1 
                  ? AppTheme.dark() 
                  : AppTheme.moon(),
          home: const PhoneShell(
            child: AuthGate(),
          ),
        );
      },
    );
  }
}