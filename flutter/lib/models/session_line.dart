enum LineType { text, system }

class SessionLine {
  const SessionLine({
    required this.type,
    required this.isUser,
    required this.content,
    this.hint,
    this.mood,
    this.reaction,
  });

  factory SessionLine.partner(String content, {String? hint, String? mood, String? reaction}) =>
      SessionLine(type: LineType.text, isUser: false, content: content, hint: hint, mood: mood, reaction: reaction);

  factory SessionLine.user(String userMock, {String? hint, String? mood, String? reaction}) =>
      SessionLine(type: LineType.text, isUser: true, content: userMock, hint: hint, mood: mood, reaction: reaction);

  factory SessionLine.director(String content) =>
      SessionLine(type: LineType.system, isUser: false, content: content);

  final LineType type;
  final bool isUser;
  final String content; // Text of the message or system event
  final String? hint;   // SOS Hint for the user if they get stuck
  final String? mood;   // Contextual mood (e.g., "Tense", "Friendly")
  final String? reaction; // AI Emoji reaction
}
