import { create } from 'zustand';
import type { TimelineConfig } from '../types/event';

interface ConfigStoreState extends TimelineConfig {
  setNanosecondsPerPixel: (value: number) => void;
  setStateColor: (state: string, color: string) => void;
  setTypeColor: (type: string, color: string) => void;
  updateConfig: (config: Partial<TimelineConfig>) => void;
}

const DEFAULT_STATE_COLORS: Record<string, string> = {
  '0': '#22c55e', // Gidle -> green
  '1': '#3b82f6', // Grunnable -> blue
  '2': '#eab308', // Grunning -> yellow
  '3': '#f97316', // Gsyscall -> orange
  '4': '#ef4444', // Gwaiting -> red
  '5': '#a855f7', // Gmoribund_unused -> purple
  '6': '#64748b', // Gdead -> gray
  '7': '#ec4899', // Genqueue_unused -> pink
  '8': '#14b8a6', // Gcopystack -> teal
  '9': '#f59e0b', // Gpreempted -> amber
};

const DEFAULT_TYPE_COLORS: Record<string, string> = {
  makeslice: '#3b82f6',
  makemap: '#8b5cf6',
  newobject: '#06b6d4',
};

export const useConfigStore = create<ConfigStoreState>((set) => ({
  nanoseconds_per_pixel: 1000000, // 1ms per pixel by default
  state_colors: DEFAULT_STATE_COLORS,
  type_colors: DEFAULT_TYPE_COLORS,
  
  setNanosecondsPerPixel: (value) => {
    set({ nanoseconds_per_pixel: value });
  },
  
  setStateColor: (state, color) => {
    set((prev) => ({
      state_colors: { ...prev.state_colors, [state]: color },
    }));
  },
  
  setTypeColor: (type, color) => {
    set((prev) => ({
      type_colors: { ...prev.type_colors, [type]: color },
    }));
  },
  
  updateConfig: (config) => {
    set(config);
  },
}));

