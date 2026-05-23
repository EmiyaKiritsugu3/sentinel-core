import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import { useSentinelData } from '../hooks/useSentinelData';

export function GraphCanvas() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [cyInstance, setCyInstance] = useState<cytoscape.Core | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    const cy = cytoscape({
      container: containerRef.current,
      style: [
        {
          selector: 'node',
          style: {
            'label': 'data(label)',
            'background-color': '#00ADB5',
            'color': '#fff',
            'text-valign': 'center',
            'text-halign': 'center',
            'font-size': '12px',
            'text-wrap': 'wrap',
            'text-max-width': '80px'
          }
        },
        {
          selector: 'edge',
          style: {
            'label': 'data(label)',
            'width': 1.5,
            'line-color': '#555',
            'target-arrow-color': '#555',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
            'font-size': '10px',
            'color': '#999',
            'text-rotation': 'autorotate'
          }
        },
        {
          selector: 'node[type="unresolved_import"]',
          style: {
            'background-color': '#B53737'
          }
        }
      ],
      layout: { name: 'preset' } // Use preset initially, then trigger layout in hook
    });

    setCyInstance(cy);

    return () => {
      cy.destroy();
    };
  }, []);

  useSentinelData(cyInstance);

  return (
    <>
      <div style={{ position: 'absolute', top: 12, left: 12, color: '#888', fontFamily: 'monospace', fontSize: 11, zIndex: 10 }}>
        Sentinel Live View
      </div>
      <div ref={containerRef} className="graph-container" />
    </>
  );
}