import { KeywordCoverageBar } from "./KeywordCoverageBar";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { StreamLanguage, defaultHighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { stex } from "@codemirror/legacy-modes/mode/stex";
import { EditorState } from "@codemirror/state";
import { EditorView, drawSelection, highlightActiveLine, highlightSpecialChars, keymap, lineNumbers } from "@codemirror/view";
import { useEffect, useRef, useState } from "react";
import { domain } from "../../../wailsjs/go/models";
import { PanelHeader } from "../../components/ui";

export function LatexWorkspace({ source, template, result, busy, compiling, versionBusy, hasOpenVersion, versionDirty, onChange, onSave, onSaveVersion, onSaveAsResume, onReload, onCompile, onExport, coverage }: { coverage?: domain.KeywordCoverage | null; source: string; template?: domain.ResumeTemplate; result: domain.CompileResult | null; busy: boolean; compiling: boolean; versionBusy: boolean; hasOpenVersion: boolean; versionDirty: boolean; onChange: (value: string) => void; onSave: () => void; onSaveVersion: () => void; onSaveAsResume: () => void; onReload: () => void; onCompile: () => void; onExport: () => void }) {
  const diagnostics = result?.diagnostics ?? [];
  const compileState = compiling ? "Compiling…" : result?.success ? "PDF ready" : result ? "Needs attention" : "Draft";
  const [diagnosticLine, setDiagnosticLine] = useState(0);
  const focusLine = (line: number) => {
    setDiagnosticLine(line > 0 ? line : 0);
  };
  return <section className="workspace-panel latex-workspace"><PanelHeader eyebrow="Source editor" title="LaTeX" description={`Editing the rendered source from ${template?.name ?? "the selected template"}. Changes compile locally after a short pause.`} action={<div className="panel-actions latex-actions"><button className="text-button" disabled={busy || compiling || !template} onClick={onReload}>Reload data</button><button className="secondary-button" disabled={busy || compiling || !template} onClick={onSave}>{template?.builtIn ? "Save editable copy" : "Save template"}</button>{hasOpenVersion ? <button className="secondary-button" disabled={versionBusy || compiling || !versionDirty} onClick={onSaveVersion}>{versionBusy ? "Saving…" : versionDirty ? "Save new version" : "Version saved"}</button> : <button className="secondary-button" disabled={versionBusy || compiling || !template || !source.trim()} onClick={onSaveAsResume}>{versionBusy ? "Saving…" : "Save as new resume"}</button>}<button className="secondary-button" disabled={compiling} onClick={onExport}>Export .tex</button><button className="primary-button" disabled={compiling} onClick={onCompile}>{compiling ? "Compiling…" : "Compile PDF"}</button></div>} /><div className="editor-tabs"><button className="active">resume.tex</button><span>{hasOpenVersion ? versionDirty ? "Saved version · edited draft" : "Immutable saved version" : template?.builtIn ? "Built-in source · edits are draft-only" : "Custom template · local"}</span></div>{coverage && <KeywordCoverageBar coverage={coverage} />}<div className="latex-editor-body"><LatexCodeEditor value={source} focusLine={diagnosticLine} onChange={onChange} />{diagnostics.length > 0 && <section className="compiler-diagnostics" aria-label="Compiler diagnostics" aria-live="polite"><header><strong>Compiler diagnostics</strong><span>{diagnostics.length} issue{diagnostics.length === 1 ? "" : "s"}</span></header><div>{diagnostics.map((diagnostic, index) => <button type="button" className={`compiler-diagnostic ${diagnostic.severity}`} key={`${diagnostic.line}-${diagnostic.message}-${index}`} onClick={() => focusLine(diagnostic.line)}><span>{diagnostic.severity}</span><strong>{diagnostic.line > 0 ? `Line ${diagnostic.line}` : "Source"}</strong><p>{diagnostic.message}</p></button>)}</div></section>}</div><footer className="editor-status" role="status" aria-live="polite"><span>{compileState}</span><span>UTF-8</span><span>{template?.name ?? "LaTeX draft"}</span><span>{source.split("\n").length} lines</span></footer></section>;
}

export function LatexCodeEditor({ value, focusLine, onChange }: { value: string; focusLine: number; onChange: (value: string) => void }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!hostRef.current) return;
    const view = new EditorView({
      parent: hostRef.current,
      state: EditorState.create({
        doc: value,
        extensions: [
          lineNumbers(), highlightSpecialChars(), history(), drawSelection(), highlightActiveLine(),
          syntaxHighlighting(defaultHighlightStyle), StreamLanguage.define(stex),
          keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
          EditorView.contentAttributes.of({ "aria-label": "LaTeX source editor" }),
          EditorView.lineWrapping,
          EditorView.theme({ "&": { height: "100%" }, ".cm-scroller": { overflow: "auto" } }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) onChangeRef.current(update.state.doc.toString());
          }),
        ],
      }),
    });
    viewRef.current = view;
    return () => { view.destroy(); viewRef.current = null; };
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view || view.state.doc.toString() === value) return;
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
  }, [value]);

  useEffect(() => {
    const view = viewRef.current;
    if (!view || focusLine < 1 || focusLine > view.state.doc.lines) return;
    const line = view.state.doc.line(focusLine);
    view.dispatch({ selection: { anchor: line.from, head: line.to }, scrollIntoView: true });
    view.focus();
  }, [focusLine]);

  return <div className="code-editor" ref={hostRef} role="region" aria-label="LaTeX source" />;
}
