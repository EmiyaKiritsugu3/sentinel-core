import { useState, useCallback } from 'react';
import cytoscape from 'cytoscape';
import { GraphCanvas } from './components/GraphCanvas';
import { StatusHUD } from './components/StatusHUD';
import { FilterToolbar } from './components/FilterToolbar';
import { useGraphFilter } from './hooks/useGraphFilter';
import './App.css';

/**
 * Renders the main application layout and wires the Cytoscape core instance to child components.
 *
 * Manages `cy` state (the Cytoscape core), provides a `handleCyReady` callback to receive the instance,
 * and passes filtered graph package data into the toolbar and the Cytoscape initializer into the canvas.
 *
 * @returns The root JSX element containing the header, toolbar, main graph canvas, and sidebar.
 */
function App() {
  const [cy, setCy] = useState<cytoscape.Core | null>(null);
  const handleCyReady = useCallback((instance: cytoscape.Core) => {
    setCy(instance);
  }, []);

  const { packages } = useGraphFilter(cy);

  return (
    <div className="app-layout">
      <div className="header">
        <StatusHUD />
      </div>
      <div className="toolbar">
        <FilterToolbar packages={packages} />
      </div>
      <div className="main">
        <GraphCanvas onCyReady={handleCyReady} />
      </div>
      <div className="sidebar">{/* Sprint 2: Event Log */}</div>
    </div>
  );
}

export default App;
