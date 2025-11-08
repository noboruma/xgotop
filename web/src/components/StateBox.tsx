import { useConfigStore } from '../store/configStore';
import type { StateEvent } from '../types/event';

interface StateBoxProps {
  state: StateEvent;
  nextState?: StateEvent;
  startTime: number;
}

const STATE_NAMES: Record<number, string> = {
  0: 'Idle',
  1: 'Runnable',
  2: 'Running',
  3: 'Syscall',
  4: 'Waiting',
  5: 'Moribund',
  6: 'Dead',
  7: 'Enqueue',
  8: 'CopyStack',
  9: 'Preempted',
};

export function StateBox({ state, nextState, startTime }: StateBoxProps) {
  const { nanoseconds_per_pixel, state_colors } = useConfigStore();
  
  // Calculate position and width
  const left = (state.timestamp - startTime) / nanoseconds_per_pixel;
  const duration = state.duration || (nextState ? nextState.timestamp - state.timestamp : 1000000);
  const width = Math.max(duration / nanoseconds_per_pixel, 20); // Minimum 20px width
  
  const color = state_colors[state.newState.toString()] || '#94a3b8';
  const stateName = STATE_NAMES[state.newState] || `State ${state.newState}`;
  
  return (
    <div
      className="absolute h-8 border-3 border-black flex items-center justify-center overflow-hidden group"
      style={{
        left: `${left}px`,
        width: `${width}px`,
        backgroundColor: color,
      }}
      title={`${stateName}\nDuration: ${(duration / 1000000).toFixed(2)}ms`}
    >
      <span className="text-xs font-bold uppercase truncate px-1">
        {stateName}
      </span>
      {width > 50 && (
        <span className="absolute bottom-0 right-1 text-[10px] font-mono">
          {(duration / 1000000).toFixed(1)}ms
        </span>
      )}
    </div>
  );
}


