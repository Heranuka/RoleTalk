class MockUser {
  const MockUser({
    required this.id,
    required this.name,
    required this.initials,
    required this.accentColor,
  });

  final String id;
  final String name;
  final String initials;
  final int accentColor; // 0xAARRGGBB
}
