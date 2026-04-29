import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppTheme {
  AppTheme._();

  // --- СТАТИЧЕСКИЕ КОНСТАНТЫ (НЕОБХОДИМЫ ДЛЯ СУЩЕСТВУЮЩИХ ЭКРАНОВ) ---
  static const Color primary = Color(0xFF22C55E); 
  static const Color primarySoft = Color(0xFFF0FDF4);
  static const Color background = Color(0xFFF3F4F6);
  static const Color surface = Colors.white;
  static const Color textPrimary = Color(0xFF111827);
  static const Color textSecondary = Color(0xFF6B7280);
  
  // Цвета для чата и сессий (исправляют твои ошибки билда)
  static const Color online = Color(0xFF22C55E);
  static const Color chatBg = Color(0xFFEDEDED);
  static const Color bubblePartner = Colors.white;
  static const Color bubbleMine = Color(0xFF95EC69);
  static const Color bubbleMineText = Colors.black;

  // --- МЕТОДЫ ГЕНЕРАЦИИ ТЕМ ---
  static ThemeData light() => _base(
    Brightness.light, 
    const Color(0xFFF3F4F6), 
    Colors.white, 
    const Color(0xFF111827), 
    Colors.white
  );

  static ThemeData dark() => _base(
    Brightness.dark, 
    const Color(0xFF121212), 
    const Color(0xFF1E1E1E), 
    Colors.white, 
    const Color(0xFF252525)
  );

  static ThemeData moon() => _base(
    Brightness.dark, 
    const Color(0xFF0D1117), 
    const Color(0xFF161B22), 
    const Color(0xFFC9D1D9), 
    const Color(0xFF21262D)
  );

  static ThemeData _base(Brightness brightness, Color bg, Color surf, Color text, Color cardColor) {
    final bool isDark = brightness == Brightness.dark;

    // Конфигурация шрифта с исправлением высоты строк (height)
    // Это лечит "кривой" английский текст
    final baseTextTheme = GoogleFonts.plusJakartaSansTextTheme().apply(
      bodyColor: text,
      displayColor: text,
    );

    return ThemeData(
      useMaterial3: true,
      brightness: brightness,
      scaffoldBackgroundColor: bg,
      cardColor: cardColor,
      colorScheme: ColorScheme.fromSeed(
        seedColor: primary,
        brightness: brightness,
        surface: surf,
        primary: primary,
      ),
      
      appBarTheme: AppBarTheme(
        backgroundColor: surf,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: true,
        titleTextStyle: GoogleFonts.plusJakartaSans(
          color: text, 
          fontSize: 18, 
          fontWeight: FontWeight.w800,
          height: 1.2, // Фикс вертикального выравнивания
        ),
        iconTheme: IconThemeData(color: text),
      ),

      textTheme: baseTextTheme.copyWith(
        headlineMedium: baseTextTheme.headlineMedium?.copyWith(
          height: 1.2, 
          fontWeight: FontWeight.w900, 
          letterSpacing: -0.5
        ),
        titleMedium: baseTextTheme.titleMedium?.copyWith(
          height: 1.2, 
          fontWeight: FontWeight.w700
        ),
        bodyLarge: baseTextTheme.bodyLarge?.copyWith(height: 1.4),
        bodyMedium: baseTextTheme.bodyMedium?.copyWith(height: 1.4),
        labelSmall: baseTextTheme.labelSmall?.copyWith(height: 1.2, letterSpacing: 0.5),
      ),

      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: surf,
        indicatorColor: primary.withOpacity(isDark ? 0.2 : 0.1),
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          return GoogleFonts.plusJakartaSans(
            fontSize: 12,
            fontWeight: states.contains(WidgetState.selected) ? FontWeight.w800 : FontWeight.w500,
            height: 1.2,
          );
        }),
      ),

      listTileTheme: ListTileThemeData(
        iconColor: primary,
        titleTextStyle: GoogleFonts.plusJakartaSans(
          color: text, 
          fontWeight: FontWeight.w700, 
          fontSize: 15, 
          height: 1.2
        ),
        subtitleTextStyle: GoogleFonts.plusJakartaSans(
          color: text.withOpacity(0.6), 
          fontSize: 12, 
          height: 1.2
        ),
      ),
    );
  }
}