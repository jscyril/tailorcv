# TailorCV Project Context

This is the durable handoff file for continuing TailorCV work in a later session. Update it whenever a meaningful implementation set is completed. It intentionally contains project and engineering context only—never personal resume data, credentials, generated resumes, or provider secrets.

## Product direction

TailorCV is a local-first Wails desktop application that turns a reusable, evidence-backed career profile into a job-specific LaTeX resume and PDF. The application must remain useful without AI. Optional AI providers may rewrite selected facts, but every claim must remain traceable to stored evidence and reviewable before export.

The planned v1 stack is Go, Wails v2, React, TypeScript, SQLite, CodeMirror, PDF.js, Tectonic, Ollama, and Gemini. `PLAN.md` is the authoritative product roadmap.

## Repository state at this handoff

- Branch: `main`, tracking `origin/main`.
- Baseline before this implementation set: `de32578 feat: add application lifecycle and GitHub metadata`.
- This handoff includes the certifications, achievements, professional-links, and ranking-signal implementation.
- The implementation set was verified, committed, and pushed together; the worktree is clean at handoff.
- No license has been selected.

## Implemented product capabilities

- Local SQLite career profile with contact details and skills.
- Ordered experience and project evidence with provenance, verification, stable fact IDs, and user-set standard/important/essential ranking priority.
- Education records and live deterministic resume rendering.
- Evidence-backed certifications and achievements with stable IDs, provenance, verification, source metadata, CRUD controls, deterministic rendering, and backup coverage.
- Ordered arbitrary professional links stored as profile records and rendered in resume contact details.
- Public GitHub repository import, complete language detection, language selection, and a review gate.
- JSON backup and atomic restore for profile, jobs, applications, and resume source history.
- Saved jobs, structured requirement extraction, alias-aware deterministic matching, FTS5 evidence search, transparent scores, user-importance and date-derived recency bonuses, and manual evidence selection.
- Read-only built-in templates, editable copies, custom `.tex` imports, template selection, and source export.
- Immutable application resume versions with job snapshots, selected fact IDs, ranking explanations, source hashes, and numbered history.
- New immutable versions created from edited saved source; prior snapshots are never updated.
- Local Tectonic compilation with time, source-size, PDF-size, log-size, isolated-workspace, and untrusted-mode controls.
- Structured compiler diagnostics with clickable navigation into the CodeMirror source editor.
- PDF.js preview rendering with zoom controls; PDF.js loads only when a compiled preview exists.
- Successful compilation of a saved immutable version records engine, duration, diagnostics, timestamp, and a private PDF artifact. JSON backup excludes the generated PDF and machine-specific path.
- Provider-neutral, versioned AI request/response contracts in `internal/ai` with strict structured decoding.
- Evidence validation rejects unselected citations, duplicate targets/citations, unsupported metrics, new recognized technologies, malformed output, and proposals without meaningful evidence overlap.
- Ollama connection/model discovery and structured generation with timeouts, cancellation, response limits, recorded contract tests, and no stored credential.
- Gemini model discovery and JSON-constrained generation behind the shared provider contract, evidence validator, cancellation, response limits, and recorded contract tests.
- Gemini API keys live only in the native operating-system keyring. The React credential field is write-only after save; SQLite stores only provider, endpoint, and model preferences.
- The AI workspace is isolated in `frontend/src/features/ai/AIWorkspace.tsx`; React DOM integration tests cover both provider setup paths, credential states, connection/model selection, cancellation, blocked validation, proposal editing/exclusion, and acceptance.
- `TestLiveProviderContract` is an opt-in Ollama/Gemini harness using fictional evidence and the production decoder. It never prints raw provider output or persists an API key.
- Side-by-side original/proposed evidence review with per-proposal edit/include controls before creating a new immutable resume version.
- Auditable AI runs record provider, model, prompt/schema versions, selected fact IDs, validation outcome, failure category, proposals, acceptance timestamp, and resume-version linkage without raw prompts or secrets.
- Application lifecycle controls transition saved applications between `draft`, `submitted`, and `archived` without changing immutable resume versions.
- Public GitHub sync stores stable repository IDs, visibility, upstream update timestamps, bounded README snapshots, complete languages, and review state. README/language rate-limit fallbacks do not discard previously imported metadata.
- Backup schema 6 includes evidence ranking priorities, professional links, certifications, achievements, GitHub repository metadata, templates, AI-run metadata, and non-secret preferences while remaining compatible with schema 1 through 5 imports. Credentials are explicitly excluded.

## Important architecture and invariants

- React calls Go only through Wails bindings. Frontend code must not access SQLite, GitHub, credentials, or the native filesystem directly.
- `internal/domain` owns validation and deterministic business rules.
- `internal/storage` owns SQLite persistence and migrations. Migration 10 adds resume hash, ranking, compilation, and artifact metadata; migration 11 adds auditable AI runs; migration 12 adds public GitHub repository identity and snapshot metadata; migration 13 adds contact links, certifications, and achievements; migration 14 adds evidence importance.
- Evidence importance is stored per experience/project bullet as `standard`, `important`, or `essential`, contributing 0/6/12 ranking points. Recency is derived at analysis time: current work adds 8, an end date within two years adds 6, and an end date within five years adds 3. These bonuses apply only after skill, term, or indexed search establishes relevance.
- The first AI schema intentionally supports only currently selectable experience bullets and reviewed project facts. New profile evidence entities require a later schema version.
- `internal/resume` owns trusted template rendering, LaTeX escaping, compiler execution, and diagnostic parsing.
- Resume source is immutable after a version is created. An edit creates the next numbered version by copying its base version's job snapshot, selected evidence, template, and ranking explanation.
- Compilation metadata is a derived attachment and may be updated without mutating the saved source snapshot.
- `ResumeVersion.PDFPath` is deliberately excluded from JSON. `PDFAvailable` is the safe UI-facing indicator.
- `ResumeContentHash` is SHA-256 over the exact UTF-8 LaTeX source.
- Tectonic resolution order is `TAILORCV_TECTONIC`, `bin/tectonic` beside the app, `tectonic` beside the app, then `PATH`. Release packaging still needs pinned binaries for each platform.
- Built-in templates remain read-only. Editing one must create a user-owned copy.
- Generated facts must never become verified facts automatically.

## Verification baseline

Run these from the repository root unless noted:

```bash
GOCACHE=/tmp/tailorcv-go-cache go test ./...
frontend/node_modules/.bin/vitest run --root frontend
frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json
npm --prefix frontend run build
```

At this handoff:

- All Go package tests pass.
- All 22 frontend unit and React integration tests pass.
- TypeScript checking passes.
- The Vite production build passes.
- The opt-in production-contract test passes against local `gemma4:12b` using fictional evidence.
- The credential adapter test binary cross-compiles with CGO disabled for Linux amd64, macOS amd64/arm64, and Windows amd64. Native packaged-app credential prompts still require verification on each operating system.
- The real Tectonic integration suite remains opt-in with `TAILORCV_TECTONIC_INTEGRATION=1`.

## Next implementation set

The next major milestone is remaining v1 product work and platform hardening:

1. Resolve the documented frontend stack drift deliberately; keep the current custom CSS/component-state approach unless a migration has a concrete maintenance or UX benefit.
2. Add a later AI schema version only if certifications and achievements can retain the same evidence citation and review guarantees.
3. Verify native credential set/get/delete behavior in packaged Windows, macOS, and Linux builds.
4. Expand from focused component/service tests to end-to-end no-AI and provider workflows through Wails bindings.

Authenticated/private GitHub access is explicitly post-v1. The v1 UI remains public-only and stores no GitHub credential.

Release-hardening work that remains:

- Ship and verify pinned Tectonic binaries/resources for Windows, macOS, and Linux rather than only supporting the runtime lookup hook.
- Add a true offline Tectonic integration fixture and platform CI coverage.
- Add end-to-end tests for edit → save new version → compile → reopen → export.
- Consider code splitting the editor bundle further; PDF.js is already lazy-loaded.

## Delivery workflow note

For the implementation set captured here, the user explicitly requested that all resulting changes be staged, committed, and pushed to GitHub. Future sessions should still confirm the current user request before treating a new external push as authorized. Preserve unrelated user changes if the worktree is not clean.
