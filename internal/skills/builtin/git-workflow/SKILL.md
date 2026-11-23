---
name: git-workflow
description: Use when the user asks to create a git commit, amend one, or create or update a pull request.
---

# Git commits and pull requests

The exact commit command, attribution trailer and HEREDOC format are in the bash tool description (`<git_commits>`). Follow them. This skill covers the process and message quality.

## Message quality (commits, PR titles, PR bodies)

- A new contributor reading only the message must understand the problem, why it matters and the impact, without the diff or internal code names.
- Avoid identifiers, filenames and implementation details unless needed for the user-facing impact.
- Bad: "Add NameFromHex with sync.Once lazy init". Good: "Improve color name lookup performance while keeping startup fast".
- Bad: "fix: nil pointer in session.go". Good: "fix: prevent session loading from crashing on missing metadata".

## Commit

1. In one message run in parallel: `git status`, `git diff`, `git log` (recent message style).
2. Stage relevant untracked files. Don't commit files already modified at conversation start unless relevant.
3. Analyze in `<commit_analysis>` tags: changed files and nature (feature, enhancement, fix, refactor, test, docs), purpose, project impact, sensitive info. Use no tools beyond git.
4. Draft 1-2 sentences on why the change exists and what it enables, not a file list. Accurate verbs ("add" = new capability, "update" = enhancement, "fix" = bug fix). First line under 72 characters; body (wrapped at 72) only for reasoning or tradeoffs. Internal-only changes: describe the maintenance benefit.
5. Commit with the HEREDOC from the bash description.
6. If a pre-commit hook fails, retry ONCE; a second failure means the hook blocks the commit. If it succeeds but modified files, amend.
7. Run `git status` to verify.

Notes: use `git commit -am` when possible; don't stage unrelated files; never update git config; don't push; no `-i` flags; no empty commits; when rebasing always use `-m`; reply with an empty response afterwards.

## Pull request

Use `gh` for all GitHub tasks.

1. In one message run in parallel: `git status`, `git diff`, whether the branch tracks a remote and is up to date, `git log` and `git diff main...HEAD` (full history since divergence).
2. Create a branch, commit and push with `-u` only if needed (pushing is allowed here because the user asked for a PR).
3. Analyze in `<pr_analysis>` tags: commits since divergence, nature of changes, purpose, impact, sensitive info.
4. Draft 1-2 bullets on "why", reflecting ALL changes since divergence, not generic.
5. `gh pr create --title "title" --body "$(cat <<'EOF'` then the summary, then `EOF` and `)"`. Append the attribution line from the bash description if one is configured.

Reply with an empty response afterwards (the user sees the `gh` output). Never update git config.
