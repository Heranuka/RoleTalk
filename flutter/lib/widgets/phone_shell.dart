import 'package:flutter/material.dart';

class PhoneShell extends StatelessWidget {
  const PhoneShell({super.key, required this.child});

  final Widget child;

  static const double maxWidth = 420;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold( // Добавили Scaffold как основу для всей оболочки
      backgroundColor: const Color(0xFFE4E6ED), // Цвет фона "вокруг" телефона
      body: Center(
        child: Container(
          constraints: const BoxConstraints(maxWidth: maxWidth),
          decoration: BoxDecoration(
            color: theme.scaffoldBackgroundColor,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.08),
                blurRadius: 24,
                offset: const Offset(0, 8),
              ),
            ],
          ),
          child: ClipRect( // Чтобы контент не вылезал за границы maxWidth
            child: Material( // Чтобы TextField и другие виджеты работали корректно
              child: child,
            ),
          ),
        ),
      ),
    );
  }
}