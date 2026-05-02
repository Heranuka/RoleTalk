import 'package:flutter/material.dart';

import '../services/auth_service.dart';
import 'auth/login_screen.dart';
import 'onboarding_screen.dart';

class AuthGate extends StatefulWidget {
  const AuthGate({super.key});

  @override
  State<AuthGate> createState() => _AuthGateState();
}

class _AuthGateState extends State<AuthGate> {
  bool? _in;

  @override
  void initState() {
    super.initState();
    _check();
  }

  Future<void> _check() async {
    final v = await AuthService.instance.isLoggedIn();
    // Проверка mounted здесь уже есть, это хорошо
    if (mounted) {
      setState(() => _in = v);
    }
  }

  // В файле auth_gate.dart

  @override
  Widget build(BuildContext context) {
    if (_in == null) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    if (!_in!) {
      return LoginScreen(onSuccess: () {
        if (mounted) setState(() => _in = true);
      });
    }

    // Если залогинены — показываем Onboarding
    return OnboardingScreen(
      onLogout: () {
        if (mounted) {
          // 1. Переключаем AuthGate в состояние "не залогинен"
          setState(() => _in = false);
        }
      },
    );
  }
}