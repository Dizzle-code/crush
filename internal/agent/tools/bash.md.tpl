Execute shell commands; long-running commands automatically move to background and return a shell ID.

<cross_platform>
Uses the mvdan/sh interpreter (Bash-compatible on all platforms incl. Windows). Use forward slashes: "ls C:/foo/bar". Common builtins and core utils are available on Windows.
</cross_platform>

<execution_steps>
1. If creating files/directories, verify the parent exists with the LS tool.
2. Dangerous commands ({{ .BannedCommands }}) need explicit user approval via warning popup in normal mode; YOLO mode blocks them without asking; Sysadmin mode approves them without asking. Safe read-only commands run without prompts.
3. Commands exceeding 1 minute (default, configurable via `auto_background_after`) move to background and return a shell ID.
4. Output over {{ .MaxOutputLength }} characters is truncated. The truncation marker names a file with the complete output; read it with the view tool when the omitted middle matters.
5. The result includes errors and metadata with <cwd></cwd> tags.
</execution_steps>

<usage_notes>
- `command` required, `working_dir` optional (defaults to current directory).
- IMPORTANT: Use grep/glob/agent instead of 'find'/'grep'. Use view/ls instead of 'cat'/'head'/'tail'/'ls'.
- Chain with ';' or '&&'; avoid newlines except inside quotes.
- Each command runs in an independent shell (no state between calls). Prefer absolute paths over 'cd' (use 'cd' only if the user asks).
- Run privileged commands with plain `sudo`, never `sudo -n`: the user can approve with Touch ID or another PAM method; without one, sudo fails at once ("a terminal is required") instead of hanging. Each command authenticates separately.
{{- if .RgAvailable }}
- Ripgrep (`rg`) is available; prefer it over `grep`.
{{- end }}
- Good: `pytest /foo/bar/tests`. Bad: `cd /foo/bar && pytest tests`.
</usage_notes>

<background_execution>
- `run_in_background=true` runs the command in a separate shell and returns a shell ID. IMPORTANT: NEVER end a command with `&`; use this parameter.
- Backgrounding buys you the command's runtime, so spend it: start the job, then keep working.
- Do NOT call job_output right after starting; a fresh job has printed nothing. Call it once its result is the next thing you need or you have run out of other work. `wait=true` blocks until the job finishes; use it only when waiting is the useful thing to do.
- A command that auto-backgrounds on timeout is still running; keep working unless you need its result now.
- Terminate with job_kill.
- Background: servers (`npm start`, `node server.js`), watchers (`npm run watch`, `tail -f`), anything that doesn't exit on its own.
- Foreground: builds, test suites, git operations, file operations, short scripts.
</background_execution>

<git_message_quality>
For commit messages, PR titles and PR bodies:
- Must be understandable to someone unfamiliar with the codebase. Litmus test: a new contributor reading only the message should understand the problem, why it matters and the impact, without the diff or internal code names.
- Avoid identifiers, filenames and implementation details unless needed to understand the user-facing impact.
- Bad: "Add NameFromHex with sync.Once lazy init". Good: "Improve color name lookup performance while keeping startup fast".
</git_message_quality>

<commit_messages>
Follow <git_message_quality>, and:
- Concise 1-2 sentences on why the change exists and what it enables, not a file list.
- Accurate verbs ("add"=new capability, "update"=enhancement, "fix"=bug fix); no generic messages.
- First line under 72 characters. Add a body (wrapped at 72) only for reasoning, tradeoffs or important context.
- Internal-only changes: describe the benefit or maintenance outcome, not private code names.
- Bad: "fix: nil pointer in session.go". Good: "fix: prevent session loading from crashing on missing metadata".
</commit_messages>

<git_commits>
When the user asks for a git commit:

1. One message with three tool_use blocks: git status (untracked files), git diff (staged/unstaged), git log (recent message style).
2. Stage relevant untracked files. Don't commit files already modified at conversation start unless relevant.
3. Analyze in <commit_analysis> tags: changed/added files and nature (feature/enhancement/fix/refactor/test/docs), purpose, project impact, sensitive info. Use no tools beyond git context.
4. Draft the message per <commit_messages>; check it against the litmus test.
5. Create the commit{{ if or (eq .Attribution.TrailerStyle "assisted-by") (eq .Attribution.TrailerStyle "co-authored-by")}} with attribution{{ end }} using HEREDOC:
   git commit -m "$(cat <<'EOF'
Commit message here.

{{ if .Attribution.GeneratedWith }}
💘 Generated with Crush
{{ end}}
{{if eq .Attribution.TrailerStyle "assisted-by" }}

Assisted-by: Crush:{{ .ModelID }}
{{ else if eq .Attribution.TrailerStyle "co-authored-by" }}

Co-Authored-By: Crush <crush@charm.land>
{{ end }}
EOF
)"
6. If a pre-commit hook fails, retry ONCE; if it fails again, the hook is blocking the commit. If it succeeds but modified files, MUST amend.
7. Run git status to verify.

Notes: use "git commit -am" when possible; don't stage unrelated files; NEVER update config; don't push; no -i flags; no empty commits; return an empty response; when rebasing always use -m.
</git_commits>

<pull_requests>
{{ if .GhAvailable -}}
Use the `gh` command for ALL GitHub tasks.
{{- end }}

When the user asks to create or update a PR:

1. One message with multiple tool_use blocks: git status, git diff, check that the branch tracks a remote and is up to date, git log and 'git diff main...HEAD' (full history since divergence from main).
2. Create a branch if needed, commit if needed, push with -u if needed.
3. Analyze in <pr_analysis> tags: commits since divergence, nature of changes, purpose, project impact, sensitive info. Use no tools beyond git context.
4. Draft the summary per <git_message_quality>: 1-2 concise bullets on "why", reflecting ALL changes since main divergence, not generic. Check against the litmus test.
5. Create with gh pr create using HEREDOC:
   gh pr create --title "title" --body "$(cat <<'EOF'

<summary>

{{ if .Attribution.GeneratedWith -}}
💘 Generated with Crush
{{- end }}

EOF
)"

Important: return an empty response (the user sees gh output); never update git config.
</pull_requests>
