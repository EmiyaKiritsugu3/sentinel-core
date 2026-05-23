import { useEffect, useState } from 'react';
import cytoscape from 'cytoscape';
import { useFilterStore } from '../stores';

function applyFilter(
  cy: cytoscape.Core,
  enabledTypes: Set<string>,
  searchText: string,
  selectedPackage: string | null,
) {
  const search = searchText.toLowerCase();
  cy.batch(() => {
    cy.nodes().forEach(node => {
      let visible = true;

      // Type filter: non-empty set = only show those types
      if (enabledTypes.size > 0 && !enabledTypes.has(node.data('type') as string)) {
        visible = false;
      }

      // Text search: match label or file_path (case-insensitive, partial)
      if (search) {
        const label = (node.data('label') as string).toLowerCase();
        const path  = (node.data('file_path') as string).toLowerCase();
        if (!label.includes(search) && !path.includes(search)) {
          visible = false;
        }
      }

      // Package filter: match top-level directory prefix
      if (selectedPackage) {
        const p = node.data('file_path') as string;
        const root = p.includes('/') ? p.split('/')[0] : '(root)';
        if (root !== selectedPackage) {
          visible = false;
        }
      }

      node.style('display', visible ? 'element' : 'none');
    });
  });
}

function extractPackages(cy: cytoscape.Core): string[] {
  return [...new Set(
    cy.nodes()
      .map(n => n.data('file_path') as string)
      .filter(p => p && p.length > 0)
      .map(p => p.includes('/') ? p.split('/')[0] : '(root)')
  )].sort();
}

/**
 * Reads filterStore and applies visibility to cytoscape nodes.
 * Re-filters when store changes or when new nodes arrive via WebSocket.
 * Returns a reactive package list for the FilterToolbar dropdown.
 */
export function useGraphFilter(cy: cytoscape.Core | null) {
  const enabledTypes    = useFilterStore(s => s.enabledTypes);
  const searchText      = useFilterStore(s => s.searchText);
  const selectedPackage = useFilterStore(s => s.selectedPackage);

  const [packages, setPackages] = useState<string[]>([]);

  // Apply filter on every store change
  useEffect(() => {
    if (!cy) return;
    applyFilter(cy, enabledTypes, searchText, selectedPackage);
    setPackages(extractPackages(cy));
  }, [cy, enabledTypes, searchText, selectedPackage]);

  // Re-apply filter when new nodes arrive via WebSocket
  useEffect(() => {
    if (!cy) return;
    const handler = () => {
      const s = useFilterStore.getState();
      applyFilter(cy, s.enabledTypes, s.searchText, s.selectedPackage);
      setPackages(extractPackages(cy));
    };
    cy.on('add', 'node', handler);
    return () => { cy.off('add', 'node', handler); };
  }, [cy]);

  return { packages };
}
