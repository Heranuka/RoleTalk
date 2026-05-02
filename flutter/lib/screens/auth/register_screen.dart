import 'package:flutter/material.dart';
import '../../services/auth_service.dart';
import '../../theme/app_theme.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key, required this.onSuccess});

  final VoidCallback onSuccess;

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _name = TextEditingController();
  final _email = TextEditingController();
  final _pass = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _name.dispose();
    _email.dispose();
    _pass.dispose();
    super.dispose();
  }

  // Регистрация через Email
  Future<void> _register() async {
    final email = _email.text.trim();
    final password = _pass.text;
    final name = _name.text.trim();

    if (email.isEmpty || password.length < 6) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Введите корректную почту и пароль (мин. 6 симв.)')),
      );
      return;
    }

    setState(() => _loading = true);
    try {
      // Вызываем метод регистрации
      // (Убедись, что на бэкенде в Go поле is_email_verified = true для тестов,
      // иначе автоматический вход после регистрации выдаст ошибку 401)
      await AuthService.instance.register(
        email: email,
        password: password,
        displayName: name.isEmpty ? email.split('@')[0] : name,
      );

      if (!mounted) return;

      // 1. Показываем уведомление об успехе
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Регистрация успешна! Входим...'),
          backgroundColor: Colors.green,
          duration: Duration(seconds: 1),
        ),
      );

      // 2. Небольшая задержка для визуального эффекта успеха
      await Future.delayed(const Duration(milliseconds: 500));

      if (!mounted) return;

      // 3. Закрываем текущую страницу регистрации
      Navigator.of(context).pop();

      // 4. Вызываем колбэк успеха, который в AuthGate переключит экран на Main/Onboarding
      widget.onSuccess();

    } on AuthException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message), backgroundColor: Colors.redAccent),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Произошла непредвиденная ошибка'), backgroundColor: Colors.redAccent),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // Регистрация/Вход через Google
  Future<void> _googleLogin() async {
    setState(() => _loading = true);
    try {
      final warn = await AuthService.instance.signInWithGoogle();
      if (!mounted) return;

      if (warn != null) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(warn)));
        return;
      }

      Navigator.of(context).pop();
      widget.onSuccess();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Ошибка входа через Google'), backgroundColor: Colors.redAccent),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        elevation: 0,
        backgroundColor: Colors.transparent,
        iconTheme: IconThemeData(color: isDark ? Colors.white : Colors.black),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Column(
              children: [
                Text(
                  'Создать аккаунт',
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w900,
                    color: AppTheme.primary,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Начните практиковать язык прямо сейчас',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: AppTheme.textSecondary),
                ),
                const SizedBox(height: 32),

                // Имя
                TextField(
                  controller: _name,
                  textCapitalization: TextCapitalization.words,
                  decoration: InputDecoration(
                    labelText: 'Ваше имя',
                    prefixIcon: const Icon(Icons.person_outline),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ),
                const SizedBox(height: 16),

                // Почта
                TextField(
                  controller: _email,
                  keyboardType: TextInputType.emailAddress,
                  autocorrect: false,
                  decoration: InputDecoration(
                    labelText: 'Почта',
                    prefixIcon: const Icon(Icons.email_outlined),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ),
                const SizedBox(height: 16),

                // Пароль
                TextField(
                  controller: _pass,
                  obscureText: true,
                  decoration: InputDecoration(
                    labelText: 'Пароль (от 6 символов)',
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ),
                const SizedBox(height: 24),

                // Кнопка регистрации
                SizedBox(
                  width: double.infinity,
                  height: 56,
                  child: ElevatedButton(
                    onPressed: _loading ? null : _register,
                    style: ElevatedButton.styleFrom(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      backgroundColor: AppTheme.primary,
                      foregroundColor: Colors.white,
                      elevation: 0,
                    ),
                    child: _loading
                        ? const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                    )
                        : const Text('Создать аккаунт', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  ),
                ),

                const SizedBox(height: 24),

                // Разделитель
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

                // Кнопка Google
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
                        const Text(
                          'Регистрация через Google',
                          style: TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }
}