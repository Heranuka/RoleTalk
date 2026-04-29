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

  Future<String?> sendVoiceToBackend(String path) async {
    try {
      final request = http.MultipartRequest(
        'POST',
        Uri.parse('http://10.0.2.2:8080/voice'), // 10.0.2.2 для Android, localhost для iOS
      );
      request.files.add(await http.MultipartFile.fromPath('file', path));

      final response = await request.send();

      if (response.statusCode == 200) {
        final bytes = await response.stream.toBytes();
        final dir = await getTemporaryDirectory();
        final responsePath = '${dir.path}/ai_res_${DateTime.now().millisecondsSinceEpoch}.wav';

        final file = File(responsePath);
        await file.writeAsBytes(bytes);
        return responsePath;
      }
    } catch (e) {
      debugPrint("Error sending voice: $e");
    }
    return null;
  }
}

Future<void> playVoiceFile(String path) async {
  final player = AudioPlayer();
  try {
    await player.play(DeviceFileSource(path));
    await player.onPlayerComplete.first;
  } catch (e) {
    debugPrint("Error playing voice: $e");
  } finally {
    await player.dispose();
  }
}