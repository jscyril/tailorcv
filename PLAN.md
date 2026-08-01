# TailorCV Implementation Plan

## Product goal

TailorCV is a local-first desktop application that turns a reusable, evidence-backed career profile into a job-specific LaTeX resume and compiled PDF. It imports useful metadata from GitHub, accepts manually entered experience, projects, certifications, achievements, education, skills, and links, ranks that material against a pasted job description, and produces an editable resume without inventing facts.

## Success criteria

- A user can build and maintain a career profile whose primary copy remains on their computer.
- Public GitHub repositories can be imported and refreshed, with private repositories supported later through explicit authentication.
- A pasted job description produces a ranked selection of relevant experience, projects, and skills.
- AI-generated wording is constrained to facts selected from the profile and remains reviewable before export.
- Every generated resume can be edited as LaTeX, compiled locally, previewed as a PDF, and saved as a versioned application.
- The app works without a cloud account; cloud AI is optional and local Ollama or deterministic matching can be used instead.

## Technology decisions

- **Desktop framework:** Wails v2, using Go directly as the desktop backend.
- **Frontend:** React, TypeScript, and Vite.
- **UI:** Tailwind CSS and shadcn/ui; React Hook Form and Zod for forms and validation.
- **Editors and preview:** CodeMirror 6 for LaTeX editing and PDF.js for embedded PDF preview.
- **Backend:** Go, organized as testable domain services exposed through generated Wails bindings.
- **Storage:** SQLite through `database/sql` and the CGO-free `modernc.org/sqlite` driver. Migrations are embedded in the executable.
- **PDF compiler:** A pinned, platform-specific Tectonic executable and resource bundle shipped with the app.
- **AI providers:** Ollama for local inference and Gemini for optional cloud inference, behind a common Go interface.
- **GitHub:** GitHub REST API for profile, repository, language, topic, and README metadata.
- **Secrets:** Operating-system credential storage; secrets never go into SQLite or exported profile data.

## Architecture

The React frontend calls Go services through Wails-generated bindings. The frontend never accesses SQLite, GitHub, AI providers, the filesystem, or Tectonic directly.

Go services:

- `ProfileService`: CRUD for personal details, experience, education, projects, skills, certifications, and achievements.
- `GitHubService`: profile connection, repository discovery, import, refresh, and review status.
- `JobService`: job-description storage, normalized requirement extraction, and application lifecycle.
- `MatchingService`: deterministic ranking, explanations, and optional embedding-assisted scoring.
- `GenerationService`: provider selection, structured AI requests, response validation, and evidence enforcement.
- `TemplateService`: installed templates, user copies, validation, and template metadata.
- `CompileService`: safe LaTeX rendering, Tectonic execution, diagnostics, and PDF output.
- `SettingsService`: local preferences and references to credentials held by the operating system.

## Data model

Use normalized tables for profiles, contact links, experience, experience bullets, education, projects, project bullets, skills, project-skill links, certifications, achievements, GitHub repositories, jobs, applications, templates, resume versions, and AI runs.

Each factual bullet or claim has a stable ID, provenance (`manual`, `github`, or `imported`), optional source URL, verification state, and timestamps. Generated bullets reference one or more fact IDs and cannot be saved as verified facts automatically.

Each resume version stores:

- The immutable job-description snapshot.
- Selected fact IDs and ranking explanations.
- Structured generated content.
- Final LaTeX source.
- Compiled PDF path and compilation diagnostics.
- Template, prompt, provider, and model identifiers.
- Content hashes and creation time.

User data is stored in the operating system's application-data directory. The app supports lossless JSON export and import of the career profile, templates, and application history.

## Resume generation pipeline

1. Normalize the pasted job description into role, seniority, responsibilities, required skills, preferred skills, and domain terms.
2. Rank stored facts using explainable deterministic signals: exact skill overlap, term frequency, domain relevance, evidence quality, user importance, and recency.
3. Optionally add semantic similarity from an Ollama or Gemini embedding provider. Do not require a vector database for the initial data scale.
4. Allow the user to inspect, add, remove, or lock selected facts before generation.
5. Send only the job requirements and selected facts to the configured text-generation provider.
6. Require provider output to match a versioned JSON Schema. Every generated bullet must include the IDs of supporting facts.
7. Reject unknown IDs, unsupported metrics, new technologies, malformed output, and claims that cannot be traced to selected facts.
8. Present generated wording for review, then render it into a trusted LaTeX template through Go code.
9. Compile and preview the PDF. Save a new immutable resume version on every accepted generation or edit.

The no-AI path selects and orders existing user-authored bullets using deterministic ranking and produces a complete resume from them.

## LaTeX and PDF safety

- Models generate structured resume content, never executable LaTeX.
- Trusted templates use Go `text/template` with custom delimiters and a mandatory LaTeX-escaping function.
- Escape backslashes, braces, dollar signs, ampersands, hashes, percent signs, underscores, tildes, carets, and other control-sensitive input.
- Compile in a fresh temporary directory with a time limit, output-size limit, shell escape disabled, and Tectonic untrusted mode enabled.
- Use a pinned offline resource bundle so ordinary compilation does not depend on the network.
- Default templates are read-only. Editing creates a user-owned copy that can be restored or deleted without affecting built-ins.
- Return structured diagnostics with source line, severity, and a user-readable message when compilation fails.

## Delivery milestones

### 1. Foundation

- Scaffold Wails v2 with React and TypeScript.
- Establish frontend navigation, Go service boundaries, SQLite migrations, application-data paths, logging, and error contracts.
- Implement profile CRUD, validation, JSON backup/restore, and automated tests.

### 2. GitHub ingestion

- Import public profile and repository metadata without authentication where possible.
- Add optional authenticated access stored in the OS credential manager.
- Import descriptions, topics, language totals, README text, URLs, visibility, fork status, and update timestamps.
- Require review before imported repositories become resume-eligible projects.

### 3. Jobs and deterministic matching

- Add job-description and application records.
- Implement requirement extraction, normalized skill aliases, FTS-backed search, ranking explanations, and manual selection controls.
- Produce a complete resume using stored bullets without an LLM.

### 4. Templates and compilation

- Ship one ATS-friendly one-page template and the pinned Tectonic toolchain.
- Add LaTeX source editing, debounced compilation, PDF preview, diagnostics, and export of `.tex` and `.pdf`.
- Add template duplication and multiple saved resume versions.

### 5. AI-assisted tailoring

- Add the provider interface, Ollama adapter, Gemini adapter, credential management, JSON Schema output, validation, retry rules, and cancellation.
- Add side-by-side review of original facts and proposed wording.
- Record provider and prompt metadata while excluding secrets and unnecessary private data.

### 6. Release hardening

- Add Windows, macOS, and Linux CI builds and platform-specific Tectonic packaging.
- Add migration, backup/restore, offline, accessibility, and failure-recovery tests.
- Add signed release artifacts and updater support after the manual-install flow is stable.

## Testing and acceptance

- Unit-test domain validation, LaTeX escaping, skill normalization, scoring, provenance enforcement, and provider-response validation.
- Integration-test SQLite migrations, GitHub fixtures, profile import/export round trips, Tectonic success and failure, and resume-version immutability.
- Contract-test every AI adapter against the same JSON Schema using recorded responses; live-provider tests remain opt-in.
- End-to-end test onboarding, profile creation, GitHub import, job analysis, project selection, generation, editing, PDF compilation, and export.
- Verify that malicious job text and model output cannot enable shell escape, inject template commands, read arbitrary files, or overwrite files outside the application workspace.
- Verify that the primary workflow works with networking disabled once the bundled compiler resources and optional local model are available.

## Initial release boundaries

Included in v1: desktop use, one local profile, public GitHub import, manual career data, job matching, Ollama and Gemini adapters, multiple LaTeX templates, PDF compilation, application history, and JSON backup/restore.

Deferred: LinkedIn scraping, collaborative/cloud accounts, mobile apps, custom model training, automatic job-board submission, recruiter analytics, and parsing arbitrary legacy resume formats. LinkedIn information is entered manually or imported from a user-provided export when a supported importer is added.

## Repository conventions

- Default branch: `main`.
- License, contribution policy, and release automation will be selected before the first public source release.
- Never commit credentials, personal profile databases, generated private resumes, PDFs, compiler caches, or local model data.
- Keep example and test fixtures fictional and clearly labeled.
