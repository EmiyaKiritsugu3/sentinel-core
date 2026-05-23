import { GraphCanvas } from './components/GraphCanvas';
import './App.css';

function App() {
  return (
    <div className="app-layout">
      <div className="header">{/* Sprint 1: Status HUD */}</div>
      <div className="toolbar">{/* Sprint 1: Filter Toolbar */}</div>
      <div className="main">
        <GraphCanvas />
      </div>
      <div className="sidebar">{/* Sprint 2: Event Log */}</div>
    </div>
  );
}

export default App;
