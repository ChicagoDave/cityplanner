import type * as THREE from 'three';
import type { SceneState } from '../scene/loader';

export interface ControlsOptions {
  ground?: THREE.Object3D;
}

export function initControls(container: HTMLElement, state: SceneState, options?: ControlsOptions): void {
  const panel = document.createElement('div');
  panel.id = 'controls-panel';

  // Build reverse indices for fast visibility updates.
  const entityLayerMap = new Map<string, string>();
  const entitySystemMap = new Map<string, string>();

  for (const [layer, ids] of Object.entries(state.groups.layers)) {
    for (const id of ids) entityLayerMap.set(id, layer);
  }
  for (const [system, ids] of Object.entries(state.groups.systems)) {
    for (const id of ids) entitySystemMap.set(id, system);
  }

  // Visibility state: entity is visible only if layer ON and (no system or system ON).
  const enabledLayers = new Set<string>(['surface']);
  const enabledSystems = new Set<string>([
    'water', 'sewage', 'electrical', 'telecom', 'vehicle', 'pedestrian', 'bicycle', 'shuttle',
  ]);

  function updateVisibility(): void {
    for (const [id, obj] of state.entityMap) {
      const layer = entityLayerMap.get(id) ?? 'surface';
      const system = entitySystemMap.get(id);
      const layerOn = enabledLayers.has(layer);
      const systemOn = !system || enabledSystems.has(system);
      obj.visible = layerOn && systemOn;
    }
  }

  // Layer toggles
  panel.appendChild(createSection('Layers'));
  const layers = ['surface', 'underground_1', 'underground_2', 'underground_3'];
  for (const layer of layers) {
    const ids = state.groups.layers[layer];
    const count = ids?.length ?? 0;
    const defaultOn = layer === 'surface';
    panel.appendChild(createToggle(
      `${formatLabel(layer)} (${count})`,
      defaultOn,
      (checked) => {
        if (checked) enabledLayers.add(layer);
        else enabledLayers.delete(layer);
        updateVisibility();
        if (layer === 'surface' && options?.ground) {
          options.ground.visible = checked;
        }
      },
    ));
  }

  // System toggles
  panel.appendChild(createSection('Systems'));
  const systems = ['water', 'sewage', 'electrical', 'telecom', 'vehicle', 'pedestrian', 'bicycle', 'shuttle'];
  for (const system of systems) {
    const ids = state.groups.systems[system];
    const count = ids?.length ?? 0;
    panel.appendChild(createToggle(
      `${formatLabel(system)} (${count})`,
      true,
      (checked) => {
        if (checked) enabledSystems.add(system);
        else enabledSystems.delete(system);
        updateVisibility();
      },
    ));
  }

  // Entity count
  const info = document.createElement('div');
  info.style.cssText = 'color:#888;font-size:11px;margin-top:8px;';
  info.textContent = `${state.entityMap.size} entities`;
  panel.appendChild(info);

  container.appendChild(panel);

  // Apply initial visibility (underground layers hidden).
  updateVisibility();
}

function createSection(title: string): HTMLElement {
  const el = document.createElement('div');
  el.style.cssText = 'font-weight:600;font-size:12px;margin-top:10px;margin-bottom:4px;color:#aaa;text-transform:uppercase;letter-spacing:0.5px;';
  el.textContent = title;
  return el;
}

function createToggle(label: string, checked: boolean, onChange: (checked: boolean) => void): HTMLElement {
  const row = document.createElement('label');
  row.style.cssText = 'display:flex;align-items:center;gap:6px;padding:2px 0;cursor:pointer;font-size:13px;';

  const cb = document.createElement('input');
  cb.type = 'checkbox';
  cb.checked = checked;
  cb.addEventListener('change', () => onChange(cb.checked));

  const span = document.createElement('span');
  span.textContent = label;

  row.appendChild(cb);
  row.appendChild(span);
  return row;
}

function formatLabel(s: string): string {
  return s.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}
