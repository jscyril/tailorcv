# TailorCV Project Context

This is the durable handoff file for continuing TailorCV work in a later session. Update it whenever a meaningful implementation set is completed. It intentionally contains project and engineering context only—never personal resume data, credentials, generated resumes, or provider secrets.

## Product direction

TailorCV is a local-first Wails desktop application that turns a reusable, evidence-backed career profile into a job-specific LaTeX resume and PDF. The application must remain useful without AI. Optional AI providers may rewrite selected facts, but every claim must remain traceable to stored evidence and reviewable before export.

The planned v1 stack is Go, Wails v2, React, TypeScript, SQLite, CodeMirror, PDF.js, Tectonic, Ollama, and Gemini. `PLAN.md` is the authoritative product roadmap.

## Repository state at this handoff

- Branch: `main`, tracking `origin/main`.
- The current hardening set adds malicious LaTeX/evidence tests, transactional restore and migration-failure rollback tests, provider/compiler/artifact recovery coverage, and keyboard/focus semantics across onboarding, navigation, evidence, diagnostics, native-dialog triggers, and PDF controls.
- It also adds native Linux amd64, macOS arm64, and Windows amd64 CI for Go tests and migrations, `go vet`, frontend tests and type-checking/build, packaged Wails artifacts, and offline Tectonic integration checks.
- The hardening set is not committed or pushed. A pre-existing deletion of `frontend/dist/.gitkeep` remains unrelated and deliberately untouched.
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
- Tectonic cache and intermediate writes are confined to the disposable compile workspace. Packaged integration tests attempt arbitrary file reads, outside writes, and shell execution under the production flags.
- Structured compiler diagnostics with clickable navigation into the CodeMirror source editor.
- PDF.js preview rendering with zoom controls; PDF.js loads only when a compiled preview exists.
- Successful compilation of a saved immutable version records engine, duration, diagnostics, timestamp, and a private PDF artifact. JSON backup excludes the generated PDF and machine-specific path.
- Reopening a compiled immutable version reloads its bounded private PDF artifact for preview and export without recompiling. Missing artifacts fall back to source-only reopen; corrupt, non-PDF, non-regular, and unexpected artifact paths are rejected and clear the export payload.
- Backup, LaTeX, PDF, and private-artifact writes use one owner-only atomic writer. Unix uses atomic rename and Windows uses `MoveFileEx` with replace-existing and write-through flags, so an existing destination can be safely replaced.
- Provider-neutral, versioned AI request/response contracts in `internal/ai` with strict structured decoding.
- Evidence validation rejects unselected citations, duplicate targets/citations, unsupported metrics, new recognized technologies, malformed output, and proposals without meaningful evidence overlap.
- Ollama connection/model discovery and structured generation with timeouts, cancellation, response limits, recorded contract tests, and no stored credential.
- Gemini model discovery and JSON-constrained generation behind the shared provider contract, evidence validator, cancellation, response limits, and recorded contract tests.
- Gemini API keys live only in the native operating-system keyring. The React credential field is write-only after save; SQLite stores only provider, endpoint, and model preferences.
- The AI workspace is isolated in `frontend/src/features/ai/AIWorkspace.tsx`; React DOM integration tests cover both provider setup paths, credential states, connection/model selection, cancellation, blocked validation, proposal editing/exclusion, and acceptance.
- `workflow_test.go` exercises deterministic analysis, evidence selection, immutable creation/edit history, public GitHub import/review/refresh preservation, compile/reopen/export, backup/restore, corrupt-artifact rejection, and recorded-Ollama generation/acceptance through the Go `App` service boundary. `frontend/src/App.workflow.test.tsx` exercises onboarding/profile creation, GitHub review gating, deterministic version creation, recorded-Ollama review/acceptance, compile/reopen/export, and backup/restore through mocked generated Wails bindings.
- `TestLiveProviderContract` is an opt-in Ollama/Gemini harness using fictional evidence and the production decoder. It never prints raw provider output or persists an API key.
- Side-by-side original/proposed evidence review with per-proposal edit/include controls before creating a new immutable resume version.
- Auditable AI runs record provider, model, prompt/schema versions, selected fact IDs, validation outcome, failure category, proposals, acceptance timestamp, and resume-version linkage without raw prompts or secrets.
- Application lifecycle controls transition saved applications between `draft`, `submitted`, and `archived` without changing immutable resume versions.
- Public GitHub sync stores stable repository IDs, visibility, upstream update timestamps, bounded README snapshots, complete languages, and review state. README/language rate-limit fallbacks do not discard previously imported metadata.
- First-run onboarding remains visible while its profile draft is edited and closes only after a successful save or an explicit “Explore first” action. Previously, entering the first character made the draft look non-empty and prematurely unmounted the form.
- First-run onboarding is isolated in `frontend/src/features/onboarding/Onboarding.tsx` with modal semantics, initial focus, focus containment, and keyboard tests. Navigation, forms, evidence selection, diagnostics, native-dialog cancellation, and PDF zoom controls have focused accessibility coverage.
- Backup schema 6 includes evidence ranking priorities, professional links, certifications, achievements, GitHub repository metadata, templates, AI-run metadata, and non-secret preferences while remaining compatible with schema 1 through 5 imports. Credentials are explicitly excluded.
- `.github/workflows/ci.yml` runs the verification suite and native Wails packaging on Ubuntu 24.04 amd64, macOS 15 arm64, and Windows 2025 amd64. Each short-lived unsigned artifact contains a checksum-verified Tectonic 0.16.9 executable and curated offline TeX Live 2022.0r0 resource bundle.
- `cmd/packagetectonic` downloads the official platform archive, verifies its pinned SHA-256 digest, hydrates only the resources exercised by the two built-in-template fixtures, writes a deterministic local ZIP bundle, and performs an offline PDF smoke compile before packaging succeeds.
- Packaged application binaries expose narrowly scoped CI verification flags before Wails startup. One runs a disposable profile → GitHub review → immutable edit/compile/reopen/export → backup/restore workflow using the real filesystem and bundled Tectonic; the other performs a random set/get/delete cycle against the native credential store and confirms cleanup without printing secrets.

## Important architecture and invariants

- React calls Go only through Wails bindings. Frontend code must not access SQLite, GitHub, credentials, or the native filesystem directly.
- The v1 frontend deliberately retains purpose-built CSS and local React component state. Tailwind, shadcn/ui, React Hook Form, and Zod are not required dependencies; Go domain validation remains authoritative. Add a frontend library only for a measured maintenance or UX problem, and split `App.tsx` incrementally as workflow tests are added.
- `internal/domain` owns validation and deterministic business rules.
- `internal/storage` owns SQLite persistence and migrations. Migration 10 adds resume hash, ranking, compilation, and artifact metadata; migration 11 adds auditable AI runs; migration 12 adds public GitHub repository identity and snapshot metadata; migration 13 adds contact links, certifications, and achievements; migration 14 adds evidence importance.
- Evidence importance is stored per experience/project bullet as `standard`, `important`, or `essential`, contributing 0/6/12 ranking points. Recency is derived at analysis time: current work adds 8, an end date within two years adds 6, and an end date within five years adds 3. These bonuses apply only after skill, term, or indexed search establishes relevance.
- The first AI schema intentionally supports only currently selectable experience bullets and reviewed project facts. New profile evidence entities require a later schema version.
- `internal/resume` owns trusted template rendering, LaTeX escaping, compiler execution, and diagnostic parsing.
- Resume source is immutable after a version is created. An edit creates the next numbered version by copying its base version's job snapshot, selected evidence, template, and ranking explanation.
- Compilation metadata is a derived attachment and may be updated without mutating the saved source snapshot.
- `ResumeVersion.PDFPath` is deliberately excluded from JSON. `PDFAvailable` is the safe UI-facing indicator.
- `ResumeContentHash` is SHA-256 over the exact UTF-8 LaTeX source.
- Tectonic resolution order is `TAILORCV_TECTONIC`, `bin/tectonic` beside the app, `tectonic` beside the app, then `PATH`. A local resource bundle is mandatory through `TAILORCV_TECTONIC_BUNDLE` or `tectonic-resources.zip` beside the selected executable, and every compile uses `--only-cached` plus `--untrusted`.
- Runtime Tectonic cache variables point inside the per-compile temporary workspace; no persistent user-cache directory is writable by compilation.
- The packaged resource ZIP intentionally covers the built-in templates rather than the multi-gigabyte upstream TeX Live archive. Custom templates that need other packages fail offline with compiler diagnostics.
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
- All 34 frontend unit and React integration tests pass.
- TypeScript checking passes.
- The Vite production build passes.
- The opt-in production-contract test passes against local `gemma4:12b` using fictional evidence.
- The credential adapter test binary cross-compiles with CGO disabled for Linux amd64, macOS amd64/arm64, and Windows amd64. Native packaged-app credential prompts still require verification on each operating system.
- The atomic writer, backup package, and full app test binaries compile for Windows amd64 with CGO disabled. Native Windows replacement behavior still belongs in packaged platform verification.
- GitHub Actions repeats tests, migrations, vetting, frontend validation, Wails packaging, Tectonic runtime assembly, real offline compilation, disposable packaged workflows, and native credential lifecycle checks on Linux, macOS, and Windows runners. Linux explicitly selects WebKitGTK 4.1 and starts an isolated Secret Service session; Windows uses the browser WebView2 strategy so CI does not bundle a runtime installer.
- The real Tectonic integration suite remains opt-in locally with `TAILORCV_TECTONIC_INTEGRATION=1`; platform CI enables it against the packaged executable and local bundle.

## Next implementation set

The next major milestone is remaining v1 product work and platform hardening:

1. Exercise `RELEASE_CHECKLIST.md` against the packaged Windows amd64, macOS arm64, and Linux amd64 artifacts. Automated UI and scripted filesystem workflows are covered, but native dialog presentation and real WebView focus still require human verification.
2. Select a license and contribution policy before the first public source release.
3. Add a later AI schema version only if certifications and achievements can retain the same evidence citation and review guarantees.

Authenticated/private GitHub access is explicitly post-v1. The v1 UI remains public-only and stores no GitHub credential.

Release-hardening work that remains:

- Complete and record the packaged release checklist on Windows, macOS, and Linux; scripted packaged-binary coverage is complete.
- Consider code splitting the roughly 565 KB main editor bundle further; PDF.js is already lazy-loaded.

## Delivery workflow note

The prior packaged-workflow set was committed and pushed at the user's explicit request. This hardening set is currently uncommitted; future sessions should not treat a commit or push as authorized without a current request. Preserve unrelated user changes if the worktree is not clean.
