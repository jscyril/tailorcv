# TailorCV packaged release checklist

Use this checklist for each unsigned release candidate on Windows amd64, macOS arm64, and Linux amd64. Test the exact artifact produced by CI, not a development build. Record failures with the platform, artifact checksum, step, observed behavior, and a screenshot when private data is not visible.

## Release candidate

- Version or commit:
- CI run:
- Artifact SHA-256:
- Tester and date:
- Platform and OS version:
- Fresh install or upgrade:

## Preflight

- [ ] The downloaded artifact checksum matches CI.
- [ ] The application launches without a terminal and without a network connection.
- [ ] The bundled Tectonic executable and `tectonic-resources.zip` are present.
- [ ] A disposable profile and fictional evidence are used throughout this check.

## First run and keyboard focus

- [ ] The onboarding dialog is announced with the title “Start with facts you control.”
- [ ] Focus starts in **Full name** and remains inside the dialog while pressing Tab or Shift+Tab.
- [ ] Required fields report validation errors without closing the dialog.
- [ ] **Explore first** closes onboarding and leaves the main workspace operable.
- [ ] Relaunching without a saved profile shows onboarding again.
- [ ] Saving a valid profile closes onboarding; relaunching does not show it again.
- [ ] The active sidebar destination is visually clear and exposed to assistive technology.

## Delivered workflows

- [ ] Create and edit profile, experience, project, education, certification, achievement, skill, and professional-link records using only the keyboard.
- [ ] Import public GitHub repositories, review one, and explicitly enable it for resume selection.
- [ ] Analyze a fictional job description and add/remove ranked evidence using the keyboard.
- [ ] Create a deterministic resume version, edit it, and save a second immutable version.
- [ ] Configure Ollama or a disposable Gemini key, verify provider-unavailable feedback, cancel a run, and review proposals without exposing a credential.
- [ ] Move an application through draft, submitted, and archived without changing resume history.

## Compilation, diagnostics, and PDF

- [ ] Both built-in templates compile with networking disabled.
- [ ] A compiler error is announced; activating a diagnostic moves focus to its source line.
- [ ] PDF zoom controls work from the keyboard and retain visible focus.
- [ ] A saved compiled version reopens with its PDF after relaunch.
- [ ] Temporarily removing its PDF artifact reopens source-only with export disabled.
- [ ] Temporarily making Tectonic unavailable shows a recoverable error and does not destroy the last saved artifact.

## Native dialogs and user-selected paths

For every dialog below, test both **Cancel** and a destination outside the default folder. Focus must return to the initiating control after cancellation or completion.

- [ ] Import a custom `.tex` template.
- [ ] Export LaTeX, including replacement of an existing destination.
- [ ] Export PDF, including replacement of an existing destination.
- [ ] Export a JSON backup, including replacement of an existing destination.
- [ ] Restore a valid JSON backup and confirm the replacement prompt.
- [ ] Reject a wrong extension, malformed backup, and unsupported backup version without changing current data.
- [ ] Paths containing spaces and non-ASCII characters work.

## Platform integration

- [ ] Native credential set/get/delete works and leaves no disposable key behind.
- [ ] Application data, artifacts, and exported files are private to the current user under normal platform permissions.
- [ ] Window focus, resize, minimize/restore, and close behavior are stable.
- [ ] No unexpected console, keyring, firewall, or file-permission prompts appear.

## Result

- [ ] Pass
- [ ] Blocked — linked issues:
- Notes:

Manual installation must pass on all three target platforms before signing artifacts or enabling updater support.
