import * as THREE from 'three';
import type { SceneGraph } from '../api';

/**
 * Loads a scene graph into a Three.js scene.
 * Returns the number of entities added.
 */
export function loadSceneGraph(
  _sceneGraph: SceneGraph,
  _scene: THREE.Scene,
): number {
  // TODO: Implement scene graph loading (ADR-008)
  // - Create Three.js objects from entities
  // - Build group containers for layer/system toggling
  // - Use InstancedMesh for repeated geometry
  // - Build octree spatial index for culling

  return 0;
}
