'use client';

/**
 * The arena's backdrop, drawn with WebGL: two soft lights in the brand's Go
 * blue and chocolate drifting under a fine grid, as light with its own alpha
 * over the card, so it suits either theme. One fragment shader on one
 * triangle, at most 30 frames a second. It holds still (one frame) under
 * reduced motion, stops while the tab is hidden, and draws nothing where WebGL
 * is not to be had: the screen's own background is under it.
 */
import { useEffect, useRef } from 'react';
import { useReducedMotionConfig } from 'motion/react';

const VERTEX = `attribute vec2 p; void main() { gl_Position = vec4(p, 0.0, 1.0); }`;

const FRAGMENT = `
precision mediump float;
uniform vec2 size;
uniform float time;
void main() {
  vec2 uv = gl_FragCoord.xy / size;
  vec2 a = vec2(0.25 + 0.15 * sin(time * 0.21), 0.7 + 0.12 * cos(time * 0.17));
  vec2 b = vec2(0.78 + 0.12 * cos(time * 0.19), 0.3 + 0.15 * sin(time * 0.23));
  float aspect = size.x / size.y;
  float la = 0.22 / (0.06 + length((uv - a) * vec2(aspect, 1.0)));
  float lb = 0.16 / (0.06 + length((uv - b) * vec2(aspect, 1.0)));
  vec3 blue = vec3(0.0, 0.678, 0.847);
  vec3 brown = vec3(0.482, 0.247, 0.0);
  vec2 cell = fract(gl_FragCoord.xy / 24.0);
  float line = step(cell.x, 0.04) + step(cell.y, 0.04);
  // Light, not paint: the color comes with its own alpha, so the lights glow
  // on the card in either theme and the grid is a faint tint of the blue.
  float alpha = clamp(la * 0.16 + lb * 0.14 + line * 0.035 * (0.5 + la), 0.0, 0.5);
  vec3 color = (blue * la + brown * lb * 1.2) / max(la + lb * 1.2, 0.001);
  gl_FragColor = vec4(color, alpha);
}`;

function compile(gl: WebGLRenderingContext, type: number, source: string) {
  const shader = gl.createShader(type);
  if (!shader) return null;
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  return gl.getShaderParameter(shader, gl.COMPILE_STATUS) ? shader : null;
}

export function ArenaBackdrop() {
  const canvas = useRef<HTMLCanvasElement>(null);
  const reduce = useReducedMotionConfig() ?? false;

  useEffect(() => {
    const c = canvas.current;
    const gl = c?.getContext('webgl', { antialias: false, alpha: true, premultipliedAlpha: false });
    if (!c || !gl) return;
    const vs = compile(gl, gl.VERTEX_SHADER, VERTEX);
    const fs = compile(gl, gl.FRAGMENT_SHADER, FRAGMENT);
    const program = gl.createProgram();
    if (!vs || !fs || !program) return;
    gl.attachShader(program, vs);
    gl.attachShader(program, fs);
    gl.linkProgram(program);
    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) return;
    gl.useProgram(program);
    // One triangle that covers the screen.
    const buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
    const p = gl.getAttribLocation(program, 'p');
    gl.enableVertexAttribArray(p);
    gl.vertexAttribPointer(p, 2, gl.FLOAT, false, 0, 0);
    const size = gl.getUniformLocation(program, 'size');
    const time = gl.getUniformLocation(program, 'time');

    let frame = 0;
    let last = 0;
    const started = performance.now();
    const draw = (now: number) => {
      const dpr = Math.min(window.devicePixelRatio || 1, 1.5);
      const w = Math.round(c.clientWidth * dpr);
      const h = Math.round(c.clientHeight * dpr);
      if (c.width !== w || c.height !== h) {
        c.width = w;
        c.height = h;
      }
      gl.viewport(0, 0, w, h);
      gl.uniform2f(size, w, h);
      gl.uniform1f(time, reduce ? 0 : (now - started) / 1000);
      gl.clearColor(0, 0, 0, 0);
      gl.clear(gl.COLOR_BUFFER_BIT);
      gl.drawArrays(gl.TRIANGLES, 0, 3);
    };
    const loop = (now: number) => {
      frame = requestAnimationFrame(loop);
      if (document.hidden || now - last < 33) return;
      last = now;
      draw(now);
    };
    if (reduce) draw(performance.now());
    else frame = requestAnimationFrame(loop);
    return () => {
      cancelAnimationFrame(frame);
      gl.deleteProgram(program);
      gl.deleteShader(vs);
      gl.deleteShader(fs);
      gl.deleteBuffer(buffer);
    };
  }, [reduce]);

  return (
    <canvas
      ref={canvas}
      aria-hidden
      className="pointer-events-none absolute inset-0 -z-10 size-full"
    />
  );
}
