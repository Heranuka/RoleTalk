import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'app_localizations.dart';

class SettingsStore extends ChangeNotifier {
  SettingsStore._();
  static final SettingsStore instance = SettingsStore._();

  static const _kNotify = 'set_notify_lobby_v1';
  static const _kVibrate = 'set_vibrate_ready_v1';
  static const _kTheme = 'set_theme_index_v1';
  static const _kLang = 'set_language_v1';
  static const _kPracticeLang = 'practice_lang_v1';

  bool _notifyLobbyReady = true;
  bool _vibrateOnReady = true;
  int _themeIndex = 0;
  String _languageCode = 'ru';
  String _practiceLanguage = 'English';

  bool get notifyLobbyReady => _notifyLobbyReady;
  bool get vibrateOnReady => _vibrateOnReady;
  int get themeIndex => _themeIndex;
  String get languageCode => _languageCode;
  String get practiceLanguage => _practiceLanguage;

  Future<void> init() async {
    final p = await SharedPreferences.getInstance();
    _notifyLobbyReady = p.getBool(_kNotify) ?? true;
    _vibrateOnReady = p.getBool(_kVibrate) ?? true;
    _themeIndex = p.getInt(_kTheme) ?? 0;
    _languageCode = p.getString(_kLang) ?? 'ru';
    _practiceLanguage = p.getString(_kPracticeLang) ?? 'English';
    notifyListeners();
  }

  Future<void> setThemeIndex(int v) async {
    _themeIndex = v;
    final p = await SharedPreferences.getInstance();
    await p.setInt(_kTheme, v);
    notifyListeners();
  }

  Future<void> setLanguage(String lang) async {
    _languageCode = lang;
    await AppLocalizations.load(lang);
    final p = await SharedPreferences.getInstance();
    await p.setString(_kLang, lang);
    notifyListeners();
  }

  Future<void> setPracticeLanguage(String lang) async {
    _practiceLanguage = lang;
    final p = await SharedPreferences.getInstance();
    await p.setString(_kPracticeLang, lang);
    notifyListeners();
  }

  Future<void> setNotifyLobbyReady(bool v) async {
    _notifyLobbyReady = v;
    final p = await SharedPreferences.getInstance();
    await p.setBool(_kNotify, v);
    notifyListeners();
  }

  Future<void> setVibrateOnReady(bool v) async {
    _vibrateOnReady = v;
    final p = await SharedPreferences.getInstance();
    await p.setBool(_kVibrate, v);
    notifyListeners();
  }
}