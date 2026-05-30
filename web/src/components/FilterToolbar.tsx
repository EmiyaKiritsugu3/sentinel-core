import { useFilterStore, type NodeType } from '../stores';
import './FilterToolbar.css';

const ALL_TYPES: NodeType[] = ['function', 'struct', 'interface', 'unresolved_import', 'file'];

interface FilterToolbarProps {
  /** Package list for dropdown, set by useGraphFilter hook. */
  packages: string[];
}

/**
 * Render a toolbar for filtering graph nodes by type, search text, and package.
 *
 * @param packages - Array of package names used to populate the package dropdown
 * @returns A JSX element containing type checkboxes, a search input, a package select, and a reset button
 */
export function FilterToolbar({ packages }: FilterToolbarProps) {
  const enabledTypes    = useFilterStore(s => s.enabledTypes);
  const searchText      = useFilterStore(s => s.searchText);
  const selectedPackage = useFilterStore(s => s.selectedPackage);
  const toggleType      = useFilterStore(s => s.toggleType);
  const setSearchText   = useFilterStore(s => s.setSearchText);
  const setSelectedPackage = useFilterStore(s => s.setSelectedPackage);
  const reset           = useFilterStore(s => s.reset);

  return (
    <div className="filter-toolbar">
      <div className="filter-toolbar__types">
        {ALL_TYPES.map(t => (
          <label key={t} className="filter-toolbar__checkbox">
            <input
              type="checkbox"
              checked={enabledTypes.size === 0 || enabledTypes.has(t)}
              onChange={() => toggleType(t)}
            />
            <span>{t}</span>
          </label>
        ))}
      </div>

      <input
        type="text"
        className="filter-toolbar__search"
        placeholder="Search nodes..."
        value={searchText}
        onChange={e => setSearchText(e.target.value)}
        aria-label="Search nodes"
      />

      <select
        className="filter-toolbar__select"
        value={selectedPackage ?? ''}
        onChange={e => setSelectedPackage(e.target.value || null)}
        aria-label="Select package"
      >
        <option value="">All packages</option>
        {packages.map(p => (
          <option key={p} value={p}>{p}</option>
        ))}
      </select>

      <button className="filter-toolbar__reset" onClick={reset}>
        Reset
      </button>
    </div>
  );
}
