import 'package:flutter_test/flutter_test.dart';

import 'package:speak_sim/main.dart';

void main() {
  testWidgets('app builds', (WidgetTester tester) async {
    await tester.pumpWidget(const SpeakSimApp());
    await tester.pump();
    expect(find.byType(SpeakSimApp), findsOneWidget);
  });
}
