import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class SessionAvatar extends StatelessWidget {
  const SessionAvatar({
    super.key,
    required this.name,
    required this.emoji,
    this.animation,
    this.isLarge = false,
  });

  final String name;
  final String emoji;
  final Animation<double>? animation;
  final bool isLarge;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final double radius = isLarge ? 48 : 32;

    return Column(
      children: [
        if (animation != null)
          AnimatedBuilder(
            animation: animation!,
            builder: (context, child) {
              final double value = animation!.value;
              return Container(
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: value > 0.01 ? AppTheme.primary : Colors.transparent,
                    width: 3,
                  ),
                  boxShadow: value > 0.01
                      ? [BoxShadow(color: AppTheme.primary.withOpacity(0.2 * value), blurRadius: 15 * value, spreadRadius: 5 * value)]
                      : [],
                ),
                child: CircleAvatar(
                  radius: radius,
                  backgroundColor: theme.cardColor,
                  child: Text(emoji, style: TextStyle(fontSize: radius * 0.8)),
                ),
              );
            },
          )
        else
          CircleAvatar(
            radius: radius,
            backgroundColor: theme.cardColor,
            child: Text(emoji, style: TextStyle(fontSize: radius * 0.8)),
          ),
        const SizedBox(height: 12),
        Text(
          name,
          style: theme.textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w800),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ],
    );
  }
}

class PulseRipple extends StatefulWidget {
  const PulseRipple({super.key, this.color = const Color(0xFFEF4444)});
  final Color color;

  @override
  State<PulseRipple> createState() => _PulseRippleState();
}

class _PulseRippleState extends State<PulseRipple> with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: const Duration(seconds: 1))..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Transform.scale(
          scale: 1.0 + _controller.value * 0.5,
          child: Opacity(
            opacity: 1.0 - _controller.value,
            child: Container(
              width: 100,
              height: 100,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: widget.color.withOpacity(0.4),
              ),
            ),
          ),
        );
      },
    );
  }
}

class HintBox extends StatelessWidget {
  const HintBox({
    super.key,
    required this.hint,
    this.onClose,
  });

  final String hint;
  final VoidCallback? onClose;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return AnimatedOpacity(
      opacity: 1.0,
      duration: const Duration(milliseconds: 300),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isDark ? Colors.white.withOpacity(0.05) : AppTheme.primarySoft,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppTheme.primary.withOpacity(0.2)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
            const Icon(Icons.lightbulb_rounded, color: Colors.amber, size: 24),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                hint,
                style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
              ),
            ),
            if (onClose != null)
              IconButton(
                icon: const Icon(Icons.close_rounded, size: 18),
                onPressed: onClose,
              ),
          ],
        ),
      ),
    );
  }
}

class FloatingSessionBar extends StatelessWidget {
  const FloatingSessionBar({
    super.key,
    required this.title,
    required this.onTap,
    required this.onClose,
  });

  final String title;
  final VoidCallback onTap;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      height: 64,
      margin: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: AppTheme.primaryGradient,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: AppTheme.primary.withOpacity(0.4),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(20),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                const PulseRipple(color: Colors.white),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Active Session',
                        style: TextStyle(color: Colors.white70, fontSize: 10, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        title,
                        style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w900),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded, color: Colors.white70),
                  onPressed: onClose,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
