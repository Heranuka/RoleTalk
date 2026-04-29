import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

/// Локальные уведомления без сервера (нужны разрешения ОС).
class LocalNotificationService {
  LocalNotificationService._();

  static final FlutterLocalNotificationsPlugin _p = FlutterLocalNotificationsPlugin();
  static bool _ready = false;

  static Future<void> init() async {
    if (_ready) return;
    try {
      const android = AndroidInitializationSettings('@mipmap/ic_launcher');
      const ios = DarwinInitializationSettings();
      final ok = await _p.initialize(
        settings: const InitializationSettings(android: android, iOS: ios),
      );
      _ready = ok ?? false;
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('LocalNotificationService init: $e\n$st');
      }
      _ready = false;
    }
  }

  static Future<void> showRoomReady(String topicTitle) async {
    if (!_ready) return;
    try {
      const android = AndroidNotificationDetails(
        'lobby',
        'Лобби',
        channelDescription: 'Комната собралась',
        importance: Importance.high,
        priority: Priority.high,
      );
      const ios = DarwinNotificationDetails();
      await _p.show(
        id: 901,
        title: 'Все готовы',
        body: topicTitle,
        notificationDetails: const NotificationDetails(android: android, iOS: ios),
      );
    } catch (e, st) {
      if (kDebugMode) {
        debugPrint('showRoomReady: $e\n$st');
      }
    }
  }

  static Future<void> showRoundTimerTick(String topicTitle) async {
    if (!_ready) return;
    try {
      const android = AndroidNotificationDetails(
        'round',
        'Раунд',
        channelDescription: 'Таймер сессии',
        importance: Importance.defaultImportance,
        priority: Priority.defaultPriority,
      );
      await _p.show(
        id: 902,
        title: 'Смена под-темы',
        body: topicTitle,
        notificationDetails: const NotificationDetails(android: android, iOS: const DarwinNotificationDetails()),
      );
    } catch (_) {}
  }
}
