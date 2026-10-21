import pdfWorkerURL from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { useEffect, useRef, useState } from "react";
import { domain } from "../../../wailsjs/go/models";
import { AchievementDraft, CertificationDraft } from "../../lib/credentials";
import { Education, EducationDraft } from "../../lib/education";
import { errorMessage } from "../../lib/errors";
import { Experience, ExperienceDraft } from "../../lib/experience";
import { Profile } from "../../lib/profile";
import { Project, ProjectDraft } from "../../lib/project";

export function ResumePreview({ profile, experiences, projects, educations,certifications,achievements, compileResult }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[];certifications:CertificationDraft[];achievements:AchievementDraft[]; compileResult: domain.CompileResult | null }) {
  const [zoom, setZoom] = useState(1);
  const visibleExperiences = experiences.slice(0, 2);
  const visibleProjects = projects.slice(0, 3);
  const compiled = Boolean(compileResult?.success && compileResult.pdfBase64);
  const previewTitle = compiled ? "Compiled PDF" : compileResult ? "Compilation failed" : "Layout preview";
  const previewDetail = compiled ? `${compileResult?.engine} · ${compileResult?.durationMs} ms` : compileResult ? `${compileResult.diagnostics?.length || 1} compiler issue${compileResult.diagnostics?.length === 1 ? "" : "s"}` : "Draft · compile to verify";
  return <aside className="preview-pane" aria-label="Resume preview"><header className="preview-toolbar"><div role="status" aria-live="polite"><strong>{previewTitle}</strong><small>{previewDetail}</small></div>{!compiled && <div className="preview-controls" role="group" aria-label="Resume preview controls"><button aria-label="Zoom out" disabled={zoom <= 0.7} onClick={() => setZoom((value) => Math.max(0.7, value - 0.1))}>−</button><output aria-live="polite">{Math.round(zoom * 100)}%</output><button aria-label="Zoom in" disabled={zoom >= 1.5} onClick={() => setZoom((value) => Math.min(1.5, value + 0.1))}>+</button><i /><span>{compileResult ? "Error" : "Draft"}</span></div>}</header>{compiled ? <PDFPreview data={compileResult?.pdfBase64 ?? ""} /> : <div className="preview-stage"><article className="resume-paper" style={{ zoom }}>
    <header className="resume-header"><h1>{profile.name || "Your Name"}</h1><p>{profile.headline || "Backend Engineer"}</p><small>{[profile.email || "your@email.com", profile.phone, profile.location, profile.website].filter(Boolean).join("  ·  ")}</small></header>
    {profile.summary && <ResumeSection title="Summary"><p className="resume-summary">{profile.summary}</p></ResumeSection>}
    <ResumeSection title="Experience">{visibleExperiences.length ? visibleExperiences.map((experience) => <div className="resume-entry" key={experience.key}><div><strong>{experience.title || "Role"} · {experience.company || "Company"}</strong><span>{experience.startDate} — {experience.current ? "Present" : experience.endDate}</span></div>{experience.location && <em>{experience.location}</em>}<ul>{experience.bullets.slice(0, 3).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Add experience and evidence bullets" />}</ResumeSection>
    <ResumeSection title="Projects">{visibleProjects.length ? visibleProjects.map((project) => <div className="resume-entry project" key={project.key}><div><strong>{project.name || "Project"}</strong><span>{project.skills.slice(0, 3).join(" · ")}</span></div>{project.description && <p>{project.description}</p>}<ul>{project.bullets.slice(0, 2).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Select projects to include them" />}</ResumeSection>
    {educations.length > 0 && <ResumeSection title="Education">{educations.slice(0, 2).map((education) => <div className="resume-entry education" key={education.key}><div><strong>{education.degree}{education.fieldOfStudy ? `, ${education.fieldOfStudy}` : ""}</strong><span>{education.startDate}{education.startDate && (education.current || education.endDate) ? " — " : ""}{education.current ? "Present" : education.endDate}</span></div><em>{education.institution}{education.location ? ` · ${education.location}` : ""}</em>{education.details && <p>{education.details}</p>}</div>)}</ResumeSection>}
    {certifications.length>0&&<ResumeSection title="Certifications">{certifications.slice(0,3).map(item=><div className="resume-entry" key={item.key}><div><strong>{item.name} · {item.issuer}</strong><span>{item.issueDate}</span></div>{item.description&&<p>{item.description}</p>}</div>)}</ResumeSection>}
    {achievements.length>0&&<ResumeSection title="Achievements"><ul>{achievements.slice(0,3).map(item=><li key={item.key}><strong>{item.title}:</strong> {item.description}</li>)}</ul></ResumeSection>}
    <ResumeSection title="Skills"><p className="resume-skills">{profile.skills.length ? profile.skills.join("  ·  ") : "Add skills to your profile"}</p></ResumeSection>
  </article></div>}</aside>;
}

export function PDFPreview({ data }: { data: string }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pageCount, setPageCount] = useState(0);
  const [renderError, setRenderError] = useState("");

  useEffect(() => {
    let cancelled = false;
    const bytes = Uint8Array.from(atob(data), (character) => character.charCodeAt(0));
    let loadingTask: { promise: Promise<any>; destroy: () => Promise<void> } | undefined;
    let renderTask: { cancel: () => void; promise: Promise<unknown> } | undefined;
    import("pdfjs-dist").then(({ getDocument, GlobalWorkerOptions }) => {
      GlobalWorkerOptions.workerSrc = pdfWorkerURL;
      loadingTask = getDocument({ data: bytes });
      return loadingTask.promise;
    }).then(async (document) => {
        if (cancelled) return;
        setPageCount(document.numPages);
        const page = await document.getPage(1);
        if (cancelled || !canvasRef.current) return;
        const viewport = page.getViewport({ scale: zoom * 1.35 });
        const pixelRatio = window.devicePixelRatio || 1;
        const canvas = canvasRef.current;
        canvas.width = Math.floor(viewport.width * pixelRatio);
        canvas.height = Math.floor(viewport.height * pixelRatio);
        canvas.style.width = `${viewport.width}px`;
        canvas.style.height = `${viewport.height}px`;
        const context = canvas.getContext("2d");
        if (!context) throw new Error("Canvas rendering is unavailable");
        const activeRenderTask = page.render({ canvasContext: context, viewport, transform: pixelRatio === 1 ? undefined : [pixelRatio, 0, 0, pixelRatio, 0, 0] });
        renderTask = activeRenderTask;
        await activeRenderTask.promise;
        setRenderError("");
      }).catch((reason) => {
      if (!cancelled && reason?.name !== "RenderingCancelledException") setRenderError(errorMessage(reason));
    });
    return () => {
      cancelled = true;
      renderTask?.cancel();
      if (loadingTask) void loadingTask.destroy();
    };
  }, [data, zoom]);

  return <div className="pdf-preview"><div className="pdf-controls" role="group" aria-label="PDF preview controls"><button aria-label="Zoom out" disabled={zoom <= 0.7} onClick={() => setZoom((value) => Math.max(0.7, value - 0.15))}>−</button><output aria-live="polite">{Math.round(zoom * 100)}%</output><button aria-label="Zoom in" disabled={zoom >= 1.6} onClick={() => setZoom((value) => Math.min(1.6, value + 0.15))}>+</button><i /><span>{pageCount || 1} page{pageCount === 1 ? "" : "s"}</span></div><div className="preview-stage compiled">{renderError ? <div className="pdf-error" role="alert">{renderError}</div> : <canvas ref={canvasRef} aria-label="Compiled resume PDF page 1" />}</div></div>;
}

export function ResumeSection({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="resume-section"><h2>{title}</h2>{children}</section>;
}

export function ResumePlaceholder({ text }: { text: string }) {
  return <p className="resume-placeholder">{text}</p>;
}
