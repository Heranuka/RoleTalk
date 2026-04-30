import 'dart:io';
import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/foundation.dart';
import 'package:path_provider/path_provider.dart';
import 'package:record/record.dart';
import 'package:http/http.dart' as http;

class VoiceCapture {
  VoiceCapture({AudioRecorder? recorder}) : _rec = recorder ?? AudioRecorder();

  final AudioRecorder _rec;

  bool get isSupported => !kIsWeb;

  Future<bool> hasPermission() async {
    if (!isSupported) return false;
    try {
      return await _rec.hasPermission();
    } catch (_) {
      return false;
    }
  }

  Future<void> start() async {
    if (!isSupported) return;
    try {
      if (!await _rec.hasPermission()) return;
      final dir = await getTemporaryDirectory();
      final path = '${dir.path}/vs_${DateTime.now().millisecondsSinceEpoch}.m4a';
      await _rec.start(
        const RecordConfig(encoder: AudioEncoder.aacLc),
        path: path,
      );
    } catch (_) {}
  }

  Future<String?> stop() async {
    if (!isSupported) return null;
    try {
      return await _rec.stop();
    } catch (_) {
      return null;
    }
  }

  Future<void> dispose() async {
    await _rec.dispose();
  }
}

Future<void> playVoiceFile(String pathOrUrl) async {
  final player = AudioPlayer();
  try {
    Source source;
    if (pathOrUrl.startsWith('http')) {
      source = UrlSource(pathOrUrl);
    } else {
      source = DeviceFileSource(pathOrUrl);
    }
    await player.play(source);
    await player.onPlayerComplete.first;
  } catch (e) {
    debugPrint("Error playing voice: $e");
  } finally {
    await player.dispose();
  }
}