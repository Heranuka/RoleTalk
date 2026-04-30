package prompt

const (
	// RoleplaySystemTemplate defines how the AI should behave during a session.
	RoleplaySystemTemplate = `Act as: {{.PartnerRole}}. 
Setting: {{.Description}}. 
{{if .SecretMotive}}Your secret motive is: {{.SecretMotive}}.{{end}}
User plays as: {{.UserRole}}. 
User's goal: {{.Goal}}. 

Respond strictly in {{.Language}}. 
Keep it extremely short (1-2 sentences), natural, and stay in character.`

	// EvaluationTemplate defines the criteria for analyzing the dialog.
	EvaluationTemplate = `Analyze this roleplay session. 
Goal was: {{.Goal}}. 

Transcript:
{{.Transcript}}

Rate the user's performance in 4 categories (0-10): 
1. Empathy (how well they understood the partner's emotions)
2. Persuasion (how effectively they moved towards the goal)
3. Structure (how logical and clear the speech was)
4. StressResistance (how they handled difficult or unexpected turns)

Return ONLY a valid JSON object in this format: 
{"empathy":X, "persuasion":X, "structure":X, "stress":X}`
)
