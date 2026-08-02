# TailorCV

TailorCV is a local-first desktop resume builder that turns a reusable career profile into an evidence-backed, job-specific LaTeX resume and PDF.

The application is being built with Wails, Go, React, TypeScript, SQLite, and Tectonic. Cloud AI is optional: TailorCV will support local generation through Ollama, cloud generation through Gemini, and a deterministic no-AI workflow.

## Current status

TailorCV is in its foundation stage. The first vertical slice provides:

- A Wails desktop shell with a React interface.
- A locally persisted career profile backed by SQLite.
- Manual experience records with ordered, provenance-aware evidence bullets.
- Reviewable projects with skills, source links, eligibility, and ordered evidence.
- Profile skill management and basic job-description skill matching.
- Clear service boundaries for GitHub import, resume generation, templates, and PDF compilation.

See [PLAN.md](PLAN.md) for the product architecture and delivery milestones.

## Principles

- **Local first:** Career data stays on the user's computer by default.
- **Evidence backed:** Generated claims must reference stored facts.
- **Provider optional:** Core resume building must work without a cloud AI provider.
- **Editable output:** Users retain the LaTeX source and compiled PDF.
- **Safe compilation:** Models produce structured content, not executable LaTeX.

## Prerequisites

- Go 1.25 or newer
- Node.js 20 or newer
- pnpm
- Wails v2 CLI
- Linux development packages required by Wails, or the corresponding Windows/macOS toolchain

Install the Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
```

## Development

Install frontend dependencies:

```bash
cd frontend
pnpm install
```

Run the desktop application with live reload:

```bash
wails dev
```

Run the checks used during development:

```bash
go test ./...
pnpm --dir frontend test
pnpm --dir frontend build
```

Build the desktop application:

```bash
wails build
```

On Linux distributions that provide WebKitGTK 4.1 rather than 4.0, use:

```bash
wails build -tags webkit2_41
```

## Local data

The desktop application stores its SQLite database below the current operating system's user configuration directory in a `tailorcv` folder. Development and test databases, generated resumes, credentials, and user PDFs must not be committed.

## Repository layout

```text
frontend/          React and TypeScript interface
internal/domain/   Core profile and matching rules
internal/storage/  SQLite persistence and migrations
app.go             Wails-facing application service
main.go            Desktop entry point
PLAN.md            Product and implementation roadmap
```

## Security

Do not report sensitive personal resume data in a public issue. TailorCV will never store provider tokens in its SQLite database; provider credentials will be kept in the operating system credential store.

## License

No license has been selected yet. Until one is added, the source remains publicly viewable but is not granted an open-source license.
