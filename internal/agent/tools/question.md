Ask the user a structured question and wait for the answer. Use for clarification, confirmation, or a choice before proceeding.

Always pass a `questions` array (1-5 items). One item renders as a plain question; several render as a tabbed form with a confirmation screen.

Every question MUST have:
- `type`: `yes_no`, `single_choice`, `multi_choice` or `free_text`
- `question`: short, one line
- `description`: markdown context (details, tradeoffs), under 300 chars. Always required; omitting it errors.

Hard limits (violations error and waste a round trip):
- Max 5 choices per question; each choice description under 100 chars.
- `single_choice` and `multi_choice` require choices (at least 2 for single).
- Max 5 questions per batch; split larger sets and tell the user follow-ups are coming.

Types:
- `yes_no`: only for a proposition to accept or reject ("Proceed with deletion?"). Never for A-vs-B or preference questions; use `single_choice` with two choices instead.
- `single_choice`: pick one named alternative, including binary ones ("TypeScript or Go?").
- `multi_choice`: pick one or more.
- `free_text`: open-ended narrative answer, no choices. Never use yes_no for open-ended questions.

Choice questions automatically include a free-text fill-in option. Do not add "Other", "Something else" or "Custom" yourself.

Batches (multiple questions): each item may have `label` (max 3 words, tab header; defaults to the first 3 words of `question`). A confirmation tab is always shown. Set `confirm_title` ("Ready to go?") and `confirm_description` (summarize what will happen, as if you already know the picks).

Example:
```json
{
  "questions": [
    {"label": "Database", "type": "single_choice", "question": "Which database?", "description": "PostgreSQL for relational data, MongoDB for documents.", "choices": [{"id": "pg", "label": "PostgreSQL"}, {"id": "mongo", "label": "MongoDB"}]},
    {"label": "Caching", "type": "yes_no", "question": "Enable caching?", "description": "Reduces latency but adds invalidation complexity."}
  ],
  "confirm_title": "Ready to configure?",
  "confirm_description": "We'll set up PostgreSQL with query caching enabled."
}
```

Use to: confirm destructive or ambiguous actions, resolve multiple valid interpretations, let the user pick options, gather related answers at once.
Do NOT use for: things answerable by reading code or docs, things other tools can find, or asking permission (use the permission system).
