import 'package:flutter/material.dart';

import '../services/app_localizations.dart'; // Импорт локализации
import '../theme/app_theme.dart';
import 'main_shell.dart';

class OnboardingScreen extends StatelessWidget {
  const OnboardingScreen({super.key, required this.onLogout});

  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    // Получаем текущую тему
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      // Автоматически подстраивается: белый, черный или темно-синий фон
      backgroundColor: theme.scaffoldBackgroundColor,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Spacer(),
              // Иконка масок всегда в фирменном зеленом цвете
              const Icon(
                Icons.theater_comedy_rounded, 
                size: 80, 
                color: AppTheme.primary,
              ),
              const SizedBox(height: 32),
              
              // Заголовок (Локализован)
              Text(
                AppLocalizations.of(context, 'onboarding_welcome'),
                style: theme.textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.w900,
                  // Цвет текста подстроится под фон автоматически
                  color: theme.textTheme.headlineMedium?.color,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              
              // Описание (Локализовано)
              Text(
                AppLocalizations.of(context, 'onboarding_description'),
                style: theme.textTheme.bodyLarge?.copyWith(
                  // Делаем второстепенный текст чуть прозрачнее
                  color: theme.textTheme.bodyLarge?.color?.withOpacity(0.7),
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 48),
              
              // Кнопка (Локализована)
              FilledButton(
                onPressed: () => _goNext(context),
                style: FilledButton.styleFrom(
                  backgroundColor: AppTheme.primary,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  minimumSize: const Size(double.infinity, 54),
                  elevation: 0,
                ),
                child: Text(
                  AppLocalizations.of(context, 'onboarding_start'), 
                  style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 18),
                ),
              ),
              const Spacer(flex: 2),
            ],
          ),
        ),
      ),
    );
  }

  void _goNext(BuildContext context) {
    Navigator.of(context).pushReplacement(
      MaterialPageRoute<void>(builder: (_) => MainShell(onLogout: onLogout)),
    );
  }
}