import 'dart:convert';
import 'package:flutter/services.dart';
import 'package:flutter/material.dart';

class AppLocalizations {
  static Map<String, String>? _localizedStrings;

  static Future<void> load(String lang) async {
    try {
      String jsonString = await rootBundle.loadString('assets/translations/$lang.json');
      Map<String, dynamic> jsonMap = json.decode(jsonString);
      _localizedStrings = jsonMap.map((key, value) => MapEntry(key, value.toString()));
    } catch (e) {
      _localizedStrings = {};
      debugPrint("Error loading localization: $e");
    }
  }

  static String of(BuildContext context, String key) {
    return _localizedStrings?[key] ?? key;
  }
}