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
    if (mounted) setState(() => _in = v);
  }

  @override
  Widget build(BuildContext context) {
    if (_in == null) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (!_in!) {
      return LoginScreen(onSuccess: () => setState(() => _in = true));
    }
    return OnboardingScreen(onLogout: () => setState(() => _in = false));
  }
}
