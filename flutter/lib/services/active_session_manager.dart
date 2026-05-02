import 'package:flutter/material.dart';
import '../models/topic_vote.dart';
import '../models/voice_room.dart';
import '../models/chat_message.dart';

class ActiveSessionManager extends ChangeNotifier {
  ActiveSessionManager._();
  static final instance = ActiveSessionManager._();

  TopicVote? _activeTopic;
  VoiceRoom? _activeRoom;
  List<ChatMessage> _messages = [];
  bool _isActive = false;

  final Set<String> _speakers = {};
  final Set<String> _audience = {};
  final Set<String> _handRaised = {};
  
  // List of minimized/unfinished sessions for the UI
  final List<TopicVote> _unfinishedSessions = [];

  TopicVote? get activeTopic => _activeTopic;
  VoiceRoom? get activeRoom => _activeRoom;
  List<ChatMessage> get messages => List.unmodifiable(_messages);
  bool get isActive => _isActive;
  
  Set<String> get speakers => _speakers;
  Set<String> get audience => _audience;
  Set<String> get handRaised => _handRaised;
  List<TopicVote> get unfinishedSessions => _unfinishedSessions;

  void startTopic(TopicVote topic) {
    _activeTopic = topic;
    _activeRoom = null;
    _messages = [];
    _isActive = true;
    notifyListeners();
  }

  void startRoom(VoiceRoom room) {
    _activeRoom = room;
    _activeTopic = null;
    _messages = [];
    _isActive = true;
    _speakers.clear();
    _audience.clear();
    _handRaised.clear();
    notifyListeners();
  }

  void addMessage(ChatMessage msg) {
    _messages.add(msg);
    notifyListeners();
  }

  void closeSession() {
    if (_activeTopic != null && _messages.isNotEmpty) {
      // Logic to save to unfinished if not reached end
      // For now we just clear
    }
    _isActive = false;
    _activeTopic = null;
    _activeRoom = null;
    _messages = [];
    _speakers.clear();
    _audience.clear();
    _handRaised.clear();
    notifyListeners();
  }

  // Audience / Speaker Logic
  void joinAudience(String name) {
    _audience.add(name);
    _speakers.remove(name);
    _handRaised.remove(name);
    notifyListeners();
  }

  void raiseHand(String name) {
    _handRaised.add(name);
    notifyListeners();
  }

  void moveToStage(String name) {
    _speakers.add(name);
    _audience.remove(name);
    _handRaised.remove(name);
    notifyListeners();
  }

  void moveToAudience(String name) {
    joinAudience(name);
  }

  // Minimize logic (actually just navigating away keeps it active)
  void markAsUnfinished(TopicVote topic) {
    if (!_unfinishedSessions.any((t) => t.id == topic.id)) {
      _unfinishedSessions.add(topic);
      notifyListeners();
    }
  }

  void removeUnfinished(String id) {
    _unfinishedSessions.removeWhere((t) => t.id == id);
    notifyListeners();
  }
}
