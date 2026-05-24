import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import { useSentinelData } from '../hooks/useSentinelData';

interface GraphCanvasProps {
  onCyReady?: (cy: cytoscape.Core) => void;
  onNodeSelect?: (node: cytoscape.NodeSingular) => void;
  onGraphTap?: () => void;
}

/**
 * Renders a Cytoscape graph inside a React component and initializes the Cytoscape instance on mount.
 *
 * Initializes a Cytoscape core with base styles and a preset layout, stores the instance in state,
 * and cleans it up on unmount. Also invokes `onCyReady` if provided after the instance is created.
 *
 * @param onCyReady - Optional callback invoked with the created Cytoscape `cy` instance once it is initialized
 * @returns The React element containing the Cytoscape mounting container and a "Sentinel Live View" label
 */
export function GraphCanvas({ onCyReady, onNodeSelect, onGraphTap }: GraphCanvasProps) {
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
    onCyReady?.(cy);

    cy.on('tap', 'node', (evt) => {
      onNodeSelect?.(evt.target);
    });
    cy.on('tap', (evt) => {
      if (evt.target === cy) {
        onGraphTap?.();
      }
    });

    return () => {
      cy.removeListener('tap');
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