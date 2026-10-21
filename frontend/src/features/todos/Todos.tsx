import { useCallback, useEffect, useState } from "react";
import { DeleteTodo, GetTodosEnabled, ListTodos, SaveTodo, SetTodoDone, SetTodosEnabled } from "../../../wailsjs/go/main/App";
import { domain } from "../../../wailsjs/go/models";
import { Field, FormBlock, PanelHeader } from "../../components/ui";
import { errorMessage } from "../../lib/errors";
import { type Application } from "../../lib/resume";
import { type Job } from "../ai/AIWorkspace";
import { resumeTitle } from "../home/Home";

type Todo = domain.Todo;
type TodoDraft = { id: string; title: string; notes: string; dueDate: string; applicationId: string };
const emptyDraft: TodoDraft = { id: "", title: "", notes: "", dueDate: "", applicationId: "" };

// Reloads when resumes change, since deleting a resume or restoring a backup changes to-dos on the server.
export function useTodos(applications: Application[], onError: (message: string) => void) {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [enabled, setEnabledState] = useState(true);
  const reload = useCallback(async () => {
    try {
      const [items, on] = await Promise.all([ListTodos(), GetTodosEnabled()]);
      setTodos(items ?? []);
      setEnabledState(on);
    } catch (reason) {
      onError(errorMessage(reason));
    }
  }, [onError]);
  useEffect(() => { void reload(); }, [reload, applications]);

  const run = async (action: () => Promise<unknown>) => {
    try {
      await action();
      setTodos((await ListTodos()) ?? []);
      return true;
    } catch (reason) {
      onError(errorMessage(reason));
      return false;
    }
  };
  return {
    todos,
    enabled,
    openCount: todos.filter((todo) => !todo.completedAt).length,
    save: (draft: TodoDraft) => run(() => SaveTodo(new domain.TodoInput(draft))),
    setDone: (id: string, done: boolean) => run(() => SetTodoDone(id, done)),
    remove: (id: string) => run(() => DeleteTodo(id)),
    setEnabled: async (on: boolean) => {
      try {
        await SetTodosEnabled(on);
        setEnabledState(on);
      } catch (reason) {
        onError(errorMessage(reason));
      }
    },
  };
}

function localToday() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}

function dueLabel(dueDate: string, today: string): { text: string; tone: string } {
  const text = new Date(`${dueDate}T00:00:00`).toLocaleDateString(undefined, { month: "short", day: "numeric", year: dueDate.slice(0, 4) === today.slice(0, 4) ? undefined : "numeric" });
  if (dueDate < today) return { text: `Overdue · ${text}`, tone: "overdue" };
  if (dueDate === today) return { text: "Due today", tone: "today" };
  return { text: `Due ${text}`, tone: "" };
}

export function TodoWorkspace({ todos, applications, jobs, onSave, onSetDone, onDelete }: { todos: Todo[]; applications: Application[]; jobs: Job[]; onSave: (draft: TodoDraft) => Promise<boolean>; onSetDone: (id: string, done: boolean) => Promise<boolean>; onDelete: (id: string) => Promise<boolean> }) {
  const [draft, setDraft] = useState<TodoDraft>(emptyDraft);
  const [busy, setBusy] = useState(false);
  const today = localToday();
  const open = todos.filter((todo) => !todo.completedAt);
  const completed = todos.filter((todo) => todo.completedAt);
  const update = (field: keyof TodoDraft, value: string) => setDraft((current) => ({ ...current, [field]: value }));
  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setBusy(true);
    if (await onSave(draft)) setDraft(emptyDraft);
    setBusy(false);
  };
  const row = (todo: Todo) => {
    const application = applications.find((item) => item.id === todo.applicationId);
    const due = todo.dueDate && !todo.completedAt ? dueLabel(todo.dueDate, today) : null;
    return <li key={todo.id} className={todo.completedAt ? "done" : ""}>
      <input type="checkbox" id={`todo-${todo.id}`} checked={Boolean(todo.completedAt)} onChange={(event) => void onSetDone(todo.id, event.target.checked)} />
      <div>
        <label htmlFor={`todo-${todo.id}`}>{todo.title}</label>
        {(due || application) && <small>{due && <span className={`todo-due ${due.tone}`}>{due.text}</span>}{application && <span>{resumeTitle(application, jobs)}</span>}</small>}
        {todo.notes && <p>{todo.notes}</p>}
      </div>
      <div className="todo-actions">
        {!todo.completedAt && <button className="text-button" aria-label={`Edit ${todo.title}`} onClick={() => setDraft({ id: todo.id, title: todo.title, notes: todo.notes, dueDate: todo.dueDate, applicationId: todo.applicationId })}>Edit</button>}
        <button className="text-button danger" aria-label={`Delete ${todo.title}`} onClick={() => { if (draft.id === todo.id) setDraft(emptyDraft); void onDelete(todo.id); }}>Delete</button>
      </div>
    </li>;
  };
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Stay on track" title="To-do" description="Follow-ups, deadlines, and prep for your applications. Link a task to a resume to keep context." />
    <form className="compact-form todo-form" onSubmit={submit}>
      <FormBlock title={draft.id ? "Edit task" : "New task"} description="Only the title is required.">
        <div className="field-grid two">
          <Field label="Task" value={draft.title} onChange={(value) => update("title", value)} placeholder="Follow up with the recruiter" required />
          <Field label="Due date" type="date" value={draft.dueDate} onChange={(value) => update("dueDate", value)} placeholder="YYYY-MM-DD" />
        </div>
        <label className="field"><span>Resume (optional)</span><select value={draft.applicationId} onChange={(event) => update("applicationId", event.target.value)}><option value="">No resume</option>{applications.map((application) => <option key={application.id} value={application.id}>{resumeTitle(application, jobs)}</option>)}</select></label>
        <label className="field"><span>Notes (optional)</span><textarea rows={3} maxLength={8192} value={draft.notes} onChange={(event) => update("notes", event.target.value)} placeholder="Links, contacts, or what to prepare." /></label>
        <div className="panel-actions">
          {draft.id && <button className="secondary-button" type="button" onClick={() => setDraft(emptyDraft)}>Cancel</button>}
          <button className="primary-button" disabled={busy || !draft.title.trim()}>{draft.id ? "Save task" : "Add task"}</button>
        </div>
      </FormBlock>
    </form>
    <section className="todo-list" aria-labelledby="todo-open-title">
      <h2 id="todo-open-title">Open · {open.length}</h2>
      {open.length ? <ul>{open.map(row)}</ul> : <p className="todo-empty">Nothing open. Add a task above.</p>}
      {completed.length > 0 && <details><summary>Completed · {completed.length}</summary><ul>{completed.map(row)}</ul></details>}
    </section>
  </section>;
}
