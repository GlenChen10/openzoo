import React from "react";
import { listWorkspaces, setWorkspaceId, type Workspace } from "@openzoo/core";

export function WorkspacePicker({ onSelect }: { onSelect?: (ws: Workspace) => void }) {
  const [workspaces, setWorkspaces] = React.useState<Workspace[]>([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    listWorkspaces()
      .then(setWorkspaces)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="p-4 text-sm text-muted-foreground">Loading workspaces...</div>;
  if (workspaces.length === 0) return <div className="p-4 text-sm text-muted-foreground">No workspaces found.</div>;

  return (
    <div className="space-y-1">
      {workspaces.map((ws) => (
        <button
          key={ws.id}
          className="w-full text-left px-3 py-2 rounded-md hover:bg-accent text-sm transition-colors"
          onClick={() => {
            setWorkspaceId(ws.id);
            onSelect?.(ws);
          }}
        >
          {ws.name}
        </button>
      ))}
    </div>
  );
}
