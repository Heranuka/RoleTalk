enum LineType { text, system }

class SessionLine {
  const SessionLine({
    required this.type,
    required this.isUser,
    required this.content,
    this.hint,
  });

  factory SessionLine.partner(String content, {String? hint}) =>
      SessionLine(type: LineType.text, isUser: false, content: content, hint: hint);

  factory SessionLine.user(String userMock, {String? hint}) =>
      SessionLine(type: LineType.text, isUser: true, content: userMock, hint: hint);

  factory SessionLine.director(String content) =>
      SessionLine(type: LineType.system, isUser: false, content: content);

  final LineType type;
  final bool isUser;
  final String content; // Text of the message or system event
  final String? hint;   // SOS Hint for the user if they get stuck
}
