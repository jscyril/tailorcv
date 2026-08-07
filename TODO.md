# TailorCV TODO

This is the working execution checklist for [PLAN.md](PLAN.md), which remains the authoritative product roadmap. [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) remains the durable engineering handoff. Keep this file short, checkable, and ordered; update the context file after completing a meaningful implementation set.

Last audited: 2026-08-08

## Completed baseline

- [x] Scaffold the Wails v2, Go, React, TypeScript, and SQLite desktop application.
- [x] Persist and validate the local profile, skills, experience evidence, projects, and education.
- [x] Preserve stable evidence IDs, provenance, verification state, ordering, and timestamps.
- [x] Import and refresh public GitHub repositories with language detection and a review gate.
- [x] Save jobs and extract structured requirements with deterministic, alias-aware matching and FTS5 evidence search.
- [x] Let users review and manually select ranked evidence before rendering a resume.
- [x] Provide read-only built-in templates, editable copies, custom `.tex` imports, and source export.
- [x] Reopen any saved resume version and save edits as a new immutable numbered version, preserving prior variations, job snapshots, evidence selections, ranking explanations, and content hashes.
- [x] Compile LaTeX with Tectonic in an isolated untrusted workspace and return structured diagnostics.
- [x] Preview compiled PDFs with lazy-loaded PDF.js and persist private PDF artifacts for saved versions.
- [x] Export and atomically restore versioned JSON backups for the currently supported profile, job, application, and resume-source data.
- [x] Pass the Go tests, 22 frontend unit and React integration tests, TypeScript check, production frontend build, and `go vet` as of the audit date.

## Evidence-constrained AI

The provider contract, Ollama and Gemini paths, persistence, review gate, focused React integration coverage, and opt-in live validation harness are delivered. Full Wails end-to-end coverage remains.

### 1. Freeze the contract and trust boundary

- [x] Decide whether the first AI schema supports only current experience/project facts or also the planned certification, achievement, and contact-link entities. Record the decision in `PLAN.md` before freezing the schema.
- [x] Add an `internal/ai` package with a provider interface that accepts context cancellation and returns provider-neutral results.
- [x] Define versioned structured request/response types and a JSON Schema. Requests must contain only normalized job requirements and explicitly selected evidence.
- [x] Define generated-content types that reference one or more selected fact IDs for every proposed bullet.
- [x] Add strict decoding: reject unknown fields, malformed JSON, oversized responses, duplicate IDs, and unsupported schema versions.

### 2. Enforce evidence before calling a real model

- [x] Validate that every cited fact exists and belongs to the current selected-fact set.
- [x] Reject uncited bullets, unknown facts, unsupported metrics, new technologies, and claims that are not supported by cited evidence.
- [x] Ensure generated wording never becomes verified profile evidence automatically.
- [x] Add adversarial unit tests for prompt injection in job text, fabricated numbers, technology substitution, mixed supported/unsupported claims, duplicate facts, and malformed output.
- [x] Keep deterministic rendering fully usable when AI is disabled or unavailable.

### 3. Persist auditable AI runs

- [x] Add a migration for AI runs and accepted generated content.
- [x] Store provider, model, prompt version, schema version, selected fact IDs, validation outcome, timestamps, and failure category.
- [x] Do not store provider secrets or unnecessary raw private payloads.
- [x] Include portable AI-run metadata in backup/restore while excluding credentials.
- [x] Test migration upgrades, backup round trips, and failure atomicity.

### 4. Implement Ollama

- [x] Add configurable local endpoint validation, health checks, and model discovery.
- [x] Implement generation with explicit connection and total timeouts, response-size limits, cancellation, and useful offline errors.
- [x] Add recorded-response contract tests for valid, invalid, partial, and timeout cases; keep live-model tests opt-in.
- [x] Expose only actually supported providers in the UI. Remove the current Claude/OpenAI placeholders unless they are added to the roadmap.

### 5. Add the review-and-accept workflow

- [x] Replace the disconnected chat placeholder with a task-focused tailoring flow based on the current job and selected evidence.
- [x] Show original evidence beside proposed wording, cited fact IDs, and validation results.
- [x] Allow proposals to be accepted, edited, or rejected individually before rendering.
- [x] On acceptance, create a new immutable resume version without changing its source facts or prior versions.
- [x] Add retry and cancel controls without allowing duplicate accepted versions.
- [x] Add focused React integration coverage for provider setup, write-only credentials, connection/model selection, cancellation, blocked runs, proposal editing, exclusion, and acceptance.
- [x] Add an opt-in live-provider harness that uses fictional evidence and the production validator without logging raw output.
- [ ] Test the full no-AI and Ollama paths at the Go service boundary and in the React UI.

## Product and data-model gaps

These are already named or implied by `PLAN.md`, but are not implemented in the audited code.

- [x] Add certifications and achievements as first-class, evidence-backed profile entities. They remain deferred from the current AI schema.
- [x] Add ordered arbitrary contact links alongside the current fixed website/GitHub/LinkedIn fields. Flexible professional links remain deferred from the current AI schema.
- [x] Add application lifecycle controls for `draft`, `submitted`, and `archived`; status changes preserve immutable resume versions and backup history.
- [x] Extend backup/restore to include custom resume templates and relevant non-secret settings. Bump the backup schema with backward-compatible import tests.
- [x] Complete the planned public GitHub metadata set: bounded README content, visibility, repository update timestamps, and stable repository IDs.
- [x] Defer authenticated/private GitHub access until post-v1; align `PLAN.md` and retain the public-only UI without a GitHub credential.
- [x] Add explicit standard/important/essential priority to evidence bullets and derive explainable recency bonuses from role/project dates without admitting otherwise irrelevant evidence.
- [ ] Resolve stack drift deliberately: Tailwind, shadcn/ui, React Hook Form, and Zod are listed in `PLAN.md` but the app currently uses custom CSS and component state. Avoid a migration unless it has a concrete maintenance or UX benefit.
- [x] Update `README.md` status after the next delivery; it currently says the templates-and-compilation milestone is still being finished even though the handoff marks it complete.

## Gemini and credential storage

- [x] Add an operating-system credential-store abstraction with platform tests or test doubles.
- [x] Add settings that store only credential references and non-secret provider preferences in SQLite.
- [x] Verify credentials are excluded from logs, backups, errors, and AI-run records.
- [x] Implement the Gemini adapter behind the same contract and validator used by Ollama.
- [x] Add recorded Gemini contract tests; keep live-provider tests opt-in.

## Release hardening

- [ ] Package pinned Tectonic executables and offline resources for Windows, macOS, and Linux.
- [ ] Ensure ordinary compilation cannot attempt network access and add a true offline integration fixture.
- [ ] Make atomic file replacement work consistently on Windows, including recompiling or re-exporting to an existing destination.
- [ ] Add platform CI for Go tests, frontend tests/type checking/builds, migrations, and packaged Wails builds.
- [x] Compile-check the OS credential adapter for Linux amd64, macOS amd64/arm64, and Windows amd64.
- [ ] Run native credential set/get/delete verification in packaged Windows, macOS, and Linux applications.
- [ ] Add end-to-end coverage for onboarding, profile creation, GitHub import/review, job analysis, evidence selection, version creation, edit, compile, reopen, backup/restore, and export.
- [ ] Add security tests proving malicious job/model/template content cannot enable shell escape, read arbitrary files, or write outside the compile workspace.
- [ ] Add accessibility testing and keyboard/focus checks for dialogs, navigation, forms, evidence selection, diagnostics, and the PDF workspace.
- [ ] Add failure-recovery tests for interrupted backup restore, missing/corrupt PDF artifacts, unavailable compilers/providers, and database migration failure.
- [ ] Add focused tests around `app.go`; current package-level statement coverage is 7.1%, and the 1,609-line `App.tsx` has no component/integration tests.
- [ ] Split `App.tsx` into feature-focused modules as those tests are added; avoid a standalone rewrite.
- [ ] Revisit editor bundle splitting after AI UI work. PDF.js is already lazy-loaded; the main production JS bundle was about 534 KB at the audit.
- [ ] Select a license and contribution policy before the first public source release.
- [ ] Add signed release artifacts and updater support only after manual installation is stable.

## Definition of done for each checked item

- Implementation is complete at the Go/domain/storage/frontend layers that the item affects.
- Relevant unit, integration, contract, or end-to-end tests pass.
- Data migrations and backup compatibility are covered when persistence changes.
- User-facing behavior and failure states are documented where needed.
- `PROJECT_CONTEXT.md` is updated after a meaningful implementation set.
- No credentials, personal resume data, generated private resumes, PDFs, databases, or compiler/model caches are committed.
