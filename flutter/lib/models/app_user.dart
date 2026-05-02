class AppUser {
  final String email;
  final String displayName;
  final String provider;
  final bool isVerified; // Добавь это

  AppUser({
    required this.email,
    required this.displayName,
    required this.provider,
    this.isVerified = false, // По умолчанию false
  });

  factory AppUser.fromJson(Map<String, dynamic> json) {
    return AppUser(
      email: json['email'],
      displayName: json['display_name'],
      provider: json['provider'] ?? 'email',
      isVerified: json['is_email_verified'] ?? false, // Читай из ответа бэкенда
    );
  }

  Map<String, dynamic> toJson() => {
    'email': email,
    'display_name': displayName,
    'provider': provider,
    'is_email_verified': isVerified,
  };
}