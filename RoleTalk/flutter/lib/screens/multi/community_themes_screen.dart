import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../services/app_localizations.dart';

class CommunityThemesScreen extends StatelessWidget {
  const CommunityThemesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        // Используем локализацию для заголовка
        title: Text(AppLocalizations.of(context, 'top_community_themes'),
          style: const TextStyle(fontWeight: FontWeight.w900)),
        centerTitle: true,
      ),
      body: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
        itemCount: 20, 
        itemBuilder: (context, i) {
          return Card(
            elevation: 0,
            margin: const EdgeInsets.only(bottom: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
              side: BorderSide(color: isDark ? Colors.white10 : Colors.black.withOpacity(0.05)),
            ),
            child: ListTile(
              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              leading: Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppTheme.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                alignment: Alignment.center,
                child: Text(_getMockEmoji(i), style: const TextStyle(fontSize: 24)),
              ),
              title: Text(
                _getMockTitle(i), 
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)
              ),
              subtitle: Text(
                'Level: ${_getMockLevel(i)} • by User_${100 + i}',
                style: const TextStyle(fontSize: 12),
              ),
              trailing: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.favorite, color: Colors.redAccent, size: 16),
                      const SizedBox(width: 4),
                      Text(
                        '${2500 - i * 120}', 
                        style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 13)
                      ),
                    ],
                  ),
                  const Icon(Icons.chevron_right, size: 16, color: Colors.grey),
                ],
              ),
              onTap: () {
                // В будущем: переход к деталям сценария
              },
            ),
          );
        },
      ),
    );
  }

  // Хелперы для генерации моковых данных
  String _getMockEmoji(int i) => ['🧨', '🛸', '💼', '🧛', '🧙', '🚀', '☕', '🏨'][i % 8];
  
  String _getMockTitle(int i) => [
    'Bank Heist', 
    'Alien Contact', 
    'Salary Negotiation', 
    'Interview with Vampire', 
    'Medieval Market', 
    'First Contact',
    'Blind Date',
    'Hotel Check-in'
  ][i % 8];

  String _getMockLevel(int i) => ['A2', 'B1', 'B2', 'C1'][i % 4];
}