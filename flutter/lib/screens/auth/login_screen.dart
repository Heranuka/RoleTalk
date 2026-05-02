import 'package:flutter/material.dart';
import '../../services/auth_service.dart';
import '../../theme/app_theme.dart';
import 'register_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, required this.onSuccess});

  final VoidCallback onSuccess;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _email = TextEditingController();
  final _pass = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _email.dispose();
    _pass.dispose();
    super.dispose();
  }

  // --- ЛОГИКА СБРОСА ПАРОЛЯ ---
  void _showForgotPasswordDialog() {
    final emailResetController = TextEditingController();
    bool isSending = false;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          title: const Text('Сброс пароля', style: TextStyle(fontWeight: FontWeight.bold)),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('Введите почту, на которую мы отправим инструкции по сбросу.'),
              const SizedBox(height: 16),
              TextField(
                controller: emailResetController,
                keyboardType: TextInputType.emailAddress,
                decoration: InputDecoration(
                  labelText: 'Email',
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Отмена'),
            ),
            ElevatedButton(
              onPressed: isSending ? null : () async {
                final email = emailResetController.text.trim();
                if (email.isEmpty) return;

                setDialogState(() => isSending = true);
                try {
                  // Вызываем метод сервиса (убедись, что он добавлен в AuthService)
                  await AuthService.instance.requestPasswordReset(email);

                  if (!mounted) return;
                  Navigator.pop(context); // Закрываем диалог

                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Инструкции отправлены на почту!'),
                      backgroundColor: Colors.green,
                    ),
                  );
                } catch (e) {
                  setDialogState(() => isSending = false);
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text(e.toString()), backgroundColor: Colors.redAccent),
                  );
                }
              },
              child: isSending
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Отправить'),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _emailLogin() async {
    if (_email.text.isEmpty || _pass.text.isEmpty) return;

    setState(() => _loading = true);
    try {
      await AuthService.instance.loginWithEmail(
        email: _email.text.trim(),
        password: _pass.text,
      );
      if (!mounted) return;
      widget.onSuccess();
    } on AuthException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message), backgroundColor: Colors.redAccent),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _googleLogin() async {
    setState(() => _loading = true);
    try {
      final warn = await AuthService.instance.signInWithGoogle();
      if (!mounted) return;

      if (warn != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(warn)),
        );
        return;
      }
      widget.onSuccess();
    } on AuthException catch (e) {
      if (!mounted) return;
      if (e.message != 'Вход отменён' && e.message != 'canceled') {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message), backgroundColor: Colors.redAccent),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Google Login Error'), backgroundColor: Colors.redAccent),
      );
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  'Role Talk',
                  textAlign: TextAlign.center,
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w900,
                    letterSpacing: -1,
                    color: AppTheme.primary,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Добро пожаловать назад',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: isDark ? Colors.white70 : AppTheme.textSecondary,
                  ),
                ),
                const SizedBox(height: 48),

                TextField(
                  controller: _email,
                  keyboardType: TextInputType.emailAddress,
                  decoration: InputDecoration(
                    labelText: 'Почта',
                    prefixIcon: const Icon(Icons.email_outlined),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _pass,
                  obscureText: true,
                  decoration: InputDecoration(
                    labelText: 'Пароль',
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ),

                // --- ДОБАВЛЕННАЯ КНОПКА "ЗАБЫЛИ ПАРОЛЬ?" ---
                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: _showForgotPasswordDialog,
                    child: Text(
                      'Забыли пароль?',
                      style: TextStyle(color: theme.hintColor, fontSize: 13),
                    ),
                  ),
                ),

                const SizedBox(height: 12),

                SizedBox(
                  width: double.infinity,
                  height: 56,
                  child: ElevatedButton(
                    onPressed: _loading ? null : _emailLogin,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppTheme.primary,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                    ),
                    child: _loading
                        ? const SizedBox(
                      height: 24,
                      width: 24,
                      child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                    )
                        : const Text('Войти', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  ),
                ),

                const SizedBox(height: 24),

                Row(
                  children: [
                    Expanded(child: Divider(color: isDark ? Colors.white24 : Colors.black12)),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: Text('ИЛИ', style: TextStyle(color: theme.hintColor, fontSize: 12)),
                    ),
                    Expanded(child: Divider(color: isDark ? Colors.white24 : Colors.black12)),
                  ],
                ),

                const SizedBox(height: 24),

                SizedBox(
                  width: double.infinity,
                  height: 56,
                  child: OutlinedButton(
                    onPressed: _loading ? null : _googleLogin,
                    style: OutlinedButton.styleFrom(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      side: BorderSide(color: isDark ? Colors.white24 : Colors.black12),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          padding: const EdgeInsets.all(4),
                          decoration: const BoxDecoration(
                            color: Colors.white,
                            shape: BoxShape.circle,
                          ),
                          child: const Text(
                            ' G ',
                            style: TextStyle(
                              color: Colors.blue,
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        const Flexible(
                          child: Text(
                            'Войти через Google',
                            style: TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),

                const SizedBox(height: 24),

                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text('Нет аккаунта?', style: TextStyle(color: theme.hintColor)),
                    TextButton(
                      onPressed: () => Navigator.push(
                        context,
                        MaterialPageRoute(builder: (_) => RegisterScreen(onSuccess: widget.onSuccess)),
                      ),
                      child: const Text('Создать', style: TextStyle(fontWeight: FontWeight.bold)),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}