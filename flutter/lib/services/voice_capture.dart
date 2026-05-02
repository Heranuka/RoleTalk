import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/foundation.dart';
import 'package:path_provider/path_provider.dart';
import 'package:record/record.dart';

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
      bool hasPerm = await _rec.hasPermission();
      if (!hasPerm) {
        print("CRITICAL: No microphone permission!");
        return;
      }
      final dir = await getTemporaryDirectory();
      final path = '${dir.path}/vs_${DateTime.now().millisecondsSinceEpoch}.m4a';

      await _rec.start(
        const RecordConfig(encoder: AudioEncoder.aacLc),
        path: path,
      );
      print("DEBUG: Recording started at $path");
    } catch (e) {
      print("CRITICAL: Error starting recorder: $e"); // Добавьте это
    }
  }

  Future<String?> stop() async {
    try {
      final path = await _rec.stop();
      print("DEBUG: Recorder stopped, file: $path");
      return path;
    } catch (e) {
      print("CRITICAL: Error stopping recorder: $e"); // Добавьте это
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