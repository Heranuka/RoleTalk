import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../services/app_localizations.dart';
import '../../models/voice_room.dart';
import 'room_wait_screen.dart';

class CreateRoomScreen extends StatefulWidget {
  const CreateRoomScreen({super.key});

  @override
  State<CreateRoomScreen> createState() => _CreateRoomScreenState();
}

class _CreateRoomScreenState extends State<CreateRoomScreen> {
  final _nameCtrl = TextEditingController();
  final _topicCtrl = TextEditingController();
  String _selectedEmoji = '🎭';
  bool _aiJudgeEnabled = true;
  String _judgePersonality = 'Balanced';

  final List<String> _personalities = ['Strict', 'Balanced', 'Funny', 'Supportive'];
  final List<String> _emojis = ['🎭', '☕', '🏢', '✈️', '🍕', '🛒', '🎓', '🏥'];

  @override
  void dispose() {
    _nameCtrl.dispose();
    _topicCtrl.dispose();
    super.dispose();
  }

  void _createRoom() {
    if (_nameCtrl.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a room name')),
      );
      return;
    }

    final room = VoiceRoom(
      id: 'room_${DateTime.now().millisecondsSinceEpoch}',
      title: _nameCtrl.text,
      subtitle: _topicCtrl.text.isEmpty ? 'General Discussion' : _topicCtrl.text,
      emoji: _selectedEmoji,
      levelTag: 'All Levels',
      onlineCount: 1,
      accent: AppTheme.primary,
      maxPlayers: 3,
      aiJudgeEnabled: _aiJudgeEnabled,
      judgePersonality: _judgePersonality,
    );

    Navigator.pushReplacement(
      context,
      MaterialPageRoute(
        builder: (_) => RoomWaitScreen(
          room: room,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(
        title: Text(AppLocalizations.of(context, 'create_room_title')),
      ),
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: isDark 
              ? [const Color(0xFF1A1A1A), const Color(0xFF121212)]
              : [const Color(0xFFF9FAFB), const Color(0xFFF3F4F6)],
          ),
        ),
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: theme.cardColor,
                    shape: BoxShape.circle,
                    boxShadow: AppTheme.premiumShadow,
                  ),
                  child: Text(_selectedEmoji, style: const TextStyle(fontSize: 48)),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                height: 60,
                child: ListView.separated(
                  scrollDirection: Axis.horizontal,
                  itemCount: _emojis.length,
                  separatorBuilder: (_, __) => const SizedBox(width: 12),
                  itemBuilder: (context, i) => GestureDetector(
                    onTap: () => setState(() => _selectedEmoji = _emojis[i]),
                    child: Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: _selectedEmoji == _emojis[i] ? AppTheme.primary : theme.cardColor,
                        borderRadius: BorderRadius.circular(16),
                        boxShadow: AppTheme.premiumShadow,
                      ),
                      child: Text(_emojis[i], style: const TextStyle(fontSize: 24)),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 32),

              _buildLabel(AppLocalizations.of(context, 'room_name_label')),
              _buildTextField(_nameCtrl, AppLocalizations.of(context, 'room_name_hint')),
              const SizedBox(height: 24),

              _buildLabel(AppLocalizations.of(context, 'room_topic_label')),
              _buildTextField(_topicCtrl, AppLocalizations.of(context, 'room_topic_hint'), maxLines: 2),
              const SizedBox(height: 32),

              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: theme.cardColor,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: AppTheme.premiumShadow,
                ),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                AppLocalizations.of(context, 'ai_judge_enable'),
                                style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 16),
                              ),
                              Text(
                                AppLocalizations.of(context, 'ai_judge_desc'),
                                style: const TextStyle(fontSize: 12, color: AppTheme.textSecondary),
                              ),
                            ],
                          ),
                        ),
                        Switch.adaptive(
                          value: _aiJudgeEnabled,
                          onChanged: (v) => setState(() => _aiJudgeEnabled = v),
                          activeColor: AppTheme.primary,
                        ),
                      ],
                    ),
                    if (_aiJudgeEnabled) ...[
                      const Divider(height: 32),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          AppLocalizations.of(context, 'judge_personality'),
                          style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14),
                        ),
                      ),
                      const SizedBox(height: 12),
                      Wrap(
                        spacing: 8,
                        children: _personalities.map((p) {
                          final label = AppLocalizations.of(context, 'personality_${p.toLowerCase()}');
                          return ChoiceChip(
                            label: Text(label),
                            selected: _judgePersonality == p,
                            onSelected: (v) {
                              if (v) setState(() => _judgePersonality = p);
                            },
                            selectedColor: AppTheme.primary.withOpacity(0.2),
                            checkmarkColor: AppTheme.primary,
                          );
                        }).toList(),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 48),

              SizedBox(
                width: double.infinity,
                height: 56,
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    gradient: AppTheme.primaryGradient,
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                        color: AppTheme.primary.withOpacity(0.3),
                        blurRadius: 12,
                        offset: const Offset(0, 6),
                      ),
                    ],
                  ),
                  child: ElevatedButton(
                    onPressed: _createRoom,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.transparent,
                      shadowColor: Colors.transparent,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                    ),
                    child: Text(
                      AppLocalizations.of(context, 'create_btn'),
                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w800, fontSize: 18),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildLabel(String text) {
    return Padding(
      padding: const EdgeInsets.only(left: 4, bottom: 8),
      child: Text(
        text,
        style: const TextStyle(fontWeight: FontWeight.w800, fontSize: 14, color: AppTheme.textSecondary),
      ),
    );
  }

  Widget _buildTextField(TextEditingController ctrl, String hint, {int maxLines = 1}) {
    return Container(
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(16),
        boxShadow: AppTheme.premiumShadow,
      ),
      child: TextField(
        controller: ctrl,
        maxLines: maxLines,
        decoration: InputDecoration(
          hintText: hint,
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        ),
      ),
    );
  }
}
