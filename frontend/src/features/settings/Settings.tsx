import { FormBlock, PanelHeader } from "../../components/ui";

export function SettingsWorkspace({ todosEnabled, onTodosEnabled }: { todosEnabled: boolean; onTodosEnabled: (enabled: boolean) => void }) {
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Preferences" title="Settings" description="Choose which tools appear in TailorCV." />
    <div className="compact-form">
      <FormBlock title="Features" description="Turning a feature off hides it; its data is kept.">
        <label className="checkbox-field"><input type="checkbox" checked={todosEnabled} onChange={(event) => onTodosEnabled(event.target.checked)} /><span>Show the to-do list</span></label>
      </FormBlock>
    </div>
  </section>;
}
