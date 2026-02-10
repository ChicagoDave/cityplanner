import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

export type CameraMode = 'first_person' | 'orbit' | 'bird_eye';

export class CameraModeManager {
  private mode: CameraMode = 'bird_eye';
  private camera: THREE.PerspectiveCamera;
  private controls: OrbitControls;

  constructor(camera: THREE.PerspectiveCamera, domElement: HTMLElement) {
    this.camera = camera;
    this.controls = new OrbitControls(camera, domElement);
    this.controls.enableDamping = true;
    this.controls.dampingFactor = 0.1;
    this.controls.maxPolarAngle = Math.PI / 2 - 0.05;
    this.setDefaultBirdEye();
  }

  getMode(): CameraMode {
    return this.mode;
  }

  setMode(_mode: CameraMode): void {
    // TODO: Implement camera mode switching (ADR-012)
  }

  fitToBounds(bounds: { min: { x: number; y: number; z: number }; max: { x: number; y: number; z: number } }): void {
    const cx = (bounds.min.x + bounds.max.x) / 2;
    const cz = (bounds.min.z + bounds.max.z) / 2;
    const dx = bounds.max.x - bounds.min.x;
    const dz = bounds.max.z - bounds.min.z;
    const span = Math.max(dx, dz);

    this.controls.target.set(cx, 0, cz);
    this.camera.position.set(cx, span * 0.8, cz + span * 0.8);
    this.camera.lookAt(cx, 0, cz);
    this.controls.update();
  }

  private setDefaultBirdEye(): void {
    this.camera.position.set(0, 1500, 1500);
    this.camera.lookAt(0, 0, 0);
    this.controls.target.set(0, 0, 0);
  }

  update(_delta: number): void {
    this.controls.update();
  }

  dispose(): void {
    this.controls.dispose();
  }
}
