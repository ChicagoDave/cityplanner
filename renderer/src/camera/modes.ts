import * as THREE from 'three';

export type CameraMode = 'first_person' | 'orbit' | 'bird_eye';

/**
 * Manages camera modes and transitions between them.
 */
export class CameraModeManager {
  private mode: CameraMode = 'bird_eye';
  private camera: THREE.PerspectiveCamera;

  constructor(camera: THREE.PerspectiveCamera) {
    this.camera = camera;
    this.setDefaultBirdEye();
  }

  getMode(): CameraMode {
    return this.mode;
  }

  setMode(_mode: CameraMode): void {
    // TODO: Implement camera mode switching (ADR-012)
    // - First person: WASD + mouse look, collision detection
    // - Orbit: click to set center, drag to orbit
    // - Bird's eye: top-down pan and zoom
    // - Smooth transitions between modes
  }

  private setDefaultBirdEye(): void {
    this.camera.position.set(0, 1500, 1500);
    this.camera.lookAt(0, 0, 0);
  }

  update(_delta: number): void {
    // TODO: Per-frame update for active camera mode
  }
}
