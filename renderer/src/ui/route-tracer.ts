import * as THREE from 'three';
import type { SceneState } from '../scene/loader';
import { getHighlightMaterial } from '../scene/materials';

/**
 * Click-to-highlight route tracing for transport infrastructure.
 * Click a transport segment (lane, pedway, bike_tunnel, pipe) to highlight
 * the full connected route via BFS on the connected_to graph.
 * Click empty space to clear the highlight.
 */
export class RouteTracer {
  private raycaster = new THREE.Raycaster();
  private mouse = new THREE.Vector2();
  private camera: THREE.Camera;
  private state: SceneState;
  private highlighted = new Map<string, THREE.Material | THREE.Material[]>();
  private infoEl: HTMLElement | null = null;

  constructor(camera: THREE.Camera, state: SceneState, domElement: HTMLElement) {
    this.camera = camera;
    this.state = state;
    domElement.addEventListener('click', (e) => this.onClick(e));

    // Info display for highlighted route
    this.infoEl = document.createElement('div');
    this.infoEl.id = 'route-info';
    this.infoEl.style.cssText =
      'position:fixed;bottom:12px;left:12px;background:rgba(0,0,0,0.75);color:#0fa;' +
      'padding:6px 12px;border-radius:4px;font-size:13px;display:none;z-index:100;';
    document.body.appendChild(this.infoEl);
  }

  private onClick(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    // Ignore clicks on UI controls
    if (target.closest('#controls-panel')) return;

    const rect = target.getBoundingClientRect();
    this.mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    this.mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    this.raycaster.setFromCamera(this.mouse, this.camera);

    // Collect visible meshes
    const meshes: THREE.Mesh[] = [];
    for (const obj of this.state.entityMap.values()) {
      if (obj.visible && obj instanceof THREE.Mesh) {
        meshes.push(obj);
      }
    }

    const intersects = this.raycaster.intersectObjects(meshes, false);
    if (intersects.length === 0) {
      this.clearHighlight();
      return;
    }

    const hit = intersects[0].object;
    const entityId = hit.name;

    // Only trace infrastructure entities with connectivity
    const meta = this.state.entityMetadata.get(entityId);
    if (!meta?.connected_to) {
      this.clearHighlight();
      return;
    }

    const route = this.traceRoute(entityId);
    this.highlightRoute(route);
  }

  private traceRoute(startId: string): Set<string> {
    const visited = new Set<string>();
    const queue = [startId];

    // Filter to same network type
    const startMeta = this.state.entityMetadata.get(startId);
    const startNetwork = startMeta?.network as string | undefined;

    while (queue.length > 0) {
      const id = queue.shift()!;
      if (visited.has(id)) continue;
      visited.add(id);

      const meta = this.state.entityMetadata.get(id);
      if (!meta) continue;

      // Only trace within the same network
      if (startNetwork && meta.network !== startNetwork) continue;

      const connectedTo = meta.connected_to as string[] | undefined;
      if (connectedTo) {
        for (const nextId of connectedTo) {
          if (!visited.has(nextId)) {
            queue.push(nextId);
          }
        }
      }
    }
    return visited;
  }

  private highlightRoute(route: Set<string>): void {
    this.clearHighlight();
    const hlMat = getHighlightMaterial();

    for (const id of route) {
      const obj = this.state.entityMap.get(id);
      if (obj && obj instanceof THREE.Mesh) {
        this.highlighted.set(id, obj.material);
        obj.material = hlMat;
      }
    }

    if (this.infoEl) {
      // Determine network from first entity
      const firstId = route.values().next().value;
      const meta = firstId ? this.state.entityMetadata.get(firstId) : undefined;
      const network = (meta?.network as string) ?? 'unknown';
      this.infoEl.textContent = `${network} network: ${route.size} segments`;
      this.infoEl.style.display = 'block';
    }
  }

  private clearHighlight(): void {
    for (const [id, originalMat] of this.highlighted) {
      const obj = this.state.entityMap.get(id);
      if (obj && obj instanceof THREE.Mesh) {
        obj.material = originalMat as THREE.Material;
      }
    }
    this.highlighted.clear();
    if (this.infoEl) {
      this.infoEl.style.display = 'none';
    }
  }
}
