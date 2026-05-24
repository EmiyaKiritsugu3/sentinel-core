import { useState, useCallback } from 'react';
import cytoscape from 'cytoscape';
import { GraphCanvas } from './components/GraphCanvas';
import { StatusHUD } from './components/StatusHUD';
import { FilterToolbar } from './components/FilterToolbar';
import { EventLog } from './components/EventLog';
import { InfoPanel } from './components/InfoPanel';
import { useGraphFilter } from './hooks/useGraphFilter';
import './App.css';

interface SelectedNodeInfo {
  id: string;
  label: string;
  type: string;
  file_path?: string;
  start_line?: number;
  end_line?: number;
}
function App() {
  const [cy, setCy] = useState<cytoscape.Core | null>(null);
  const [selectedNode, setSelectedNode] = useState<SelectedNodeInfo | null>(null);
  const handleCyReady = useCallback((instance: cytoscape.Core) => {
    setCy(instance);
  }, []);

  const { packages } = useGraphFilter(cy);

  const port = window.location.port === '5173' ? '8080' : window.location.port;
  const host = `${window.location.hostname}:${port}`;
  const baseUrl = `http://${host}`;

  const handleNodeSelect = useCallback((node: cytoscape.NodeSingular) => {
    const d = node.data();
    setSelectedNode({
      id: d.id as string,
      label: d.label as string,
      type: d.type as string,
      file_path: d.file_path as string | undefined,
      start_line: d.start_line as number | undefined,
      end_line: d.end_line as number | undefined,
    });
  }, []);

  const handleGraphTap = useCallback(() => {
    setSelectedNode(null);
  }, []);

  return (
    <div className="app-layout">
      <div className="header">
        <StatusHUD />
      </div>
      <div className="toolbar">
        <FilterToolbar packages={packages} />
      </div>
      <div className="main">
        <GraphCanvas
          onCyReady={handleCyReady}
          onNodeSelect={handleNodeSelect}
          onGraphTap={handleGraphTap}
        />
      </div>
      <div className="sidebar">
        {selectedNode ? (
          <InfoPanel node={selectedNode} baseUrl={baseUrl} onClose={() => setSelectedNode(null)} />
        ) : (
          <EventLog />
        )}
      </div>

    </div>
  );
}

export default App;
