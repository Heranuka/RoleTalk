import 'package:flutter/material.dart';

/// Ограничивает ширину как у телефона (удобно на планшете/десктопе при разработке).
class PhoneShell extends StatelessWidget {
  const PhoneShell({super.key, required this.child});

  final Widget child;

  static const double maxWidth = 420;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: const Color(0xFFE4E6ED),
      child: Center(
        child: Container(
          constraints: const BoxConstraints(maxWidth: maxWidth),
          decoration: BoxDecoration(
            color: Theme.of(context).scaffoldBackgroundColor,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.08),
                blurRadius: 24,
                offset: const Offset(0, 8),
              ),
            ],
          ),
          child: child,
        ),
      ),
    );
  }
}
