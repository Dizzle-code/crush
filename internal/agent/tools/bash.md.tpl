Execute shell commands; long-running commands automatically move to background and return a shell ID.

<cross_platform>
mvdan/sh interpreter (Bash-compatible, incl. Windows). Use forward slashes: "ls C:/foo/bar". Common core utils are built in.
</cross_platform>

<usage_notes>
- `command` required, `working_dir` optional. Each call is an independent shell (no state); prefer absolute paths over 'cd'.
- Use grep/glob/agent instead of 'find'/'grep', and view/ls instead of 'cat'/'head'/'tail'/'ls'.
- Chain with ';' or '&&'; avoid newlines outside quotes. Verify parent directories with ls before creating files.
- Dangerous commands ({{ .BannedCommands }}) need user approval in normal mode, are blocked in YOLO mode and auto-approved in Sysadmin mode. Safe read-only commands run without prompts.
- Output over {{ .MaxOutputLength }} characters is truncated; the marker names a file with the full output (read it with view).
- Commands over 1 minute (default, configurable via `auto_background_after`) move to background and return a shell ID.
- Privileged commands: plain `sudo`, never `sudo -n` (the user can approve via Touch ID or PAM; otherwise it fails at once instead of hanging).
{{- if .RgAvailable }}
- Ripgrep (`rg`) is available; prefer it over `grep`.
{{- end }}
</usage_notes>

<background_execution>
- `run_in_background=true` returns a shell ID; NEVER end a command with `&`. Then keep working.
- Don't call job_output right after starting (nothing printed yet); call it when you need the result or have no other work. `wait=true` blocks until done. Stop with job_kill. An auto-backgrounded command is still running.
- Background: servers, watchers, anything that doesn't exit. Foreground: builds, tests, git, file operations, short scripts.
</background_execution>

<git_commits>
Only when the user asks for a commit. If a git skill is listed, load it first for the process and message quality. Commit with a HEREDOC{{ if or (eq .Attribution.TrailerStyle "assisted-by") (eq .Attribution.TrailerStyle "co-authored-by")}} and attribution{{ end }}:
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
Never push unless the user explicitly asks.
</git_commits>
