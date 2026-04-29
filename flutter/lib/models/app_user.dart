class AppUser {
  const AppUser({
    required this.email,
    required this.displayName,
    required this.provider,
  });

  final String email;
  final String displayName;
  /// `email` | `google`
  final String provider;

  Map<String, dynamic> toJson() => {
        'email': email,
        'displayName': displayName,
        'provider': provider,
      };

  static AppUser fromJson(Map<String, dynamic> j) {
    return AppUser(
      email: j['email'] as String,
      displayName: j['displayName'] as String,
      provider: j['provider'] as String,
    );
  }
}
