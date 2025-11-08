import type { GoroutineState } from '../types/event';
import { StateBox } from './StateBox';
import { AllocBox } from './AllocBox';
import { useConfigStore } from '../store/configStore';

interface GoroutineRowProps {
  goroutine: GoroutineState;
}

const KIND_NAMES: Record<number, string> = {
  0: 'Invalid', 1: 'bool', 2: 'int', 3: 'int8', 4: 'int16', 5: 'int32', 6: 'int64',
  7: 'uint', 8: 'uint8', 9: 'uint16', 10: 'uint32', 11: 'uint64', 12: 'uintptr',
  13: 'float32', 14: 'float64', 15: 'complex64', 16: 'complex128',
  17: 'Array', 18: 'Chan', 19: 'Func', 20: 'Interface', 21: 'Map', 22: 'Ptr',
  23: 'Slice', 24: 'String', 25: 'Struct', 26: 'UnsafePtr',
};

export function GoroutineRow({ goroutine }: GoroutineRowProps) {
  const { nanoseconds_per_pixel } = useConfigStore();
  
  return (
    <div className="border-b-3 border-black bg-background">
      {/* Goroutine header */}
      <div className="flex items-center border-b-2 border-black bg-secondary px-3 py-1">
        <span className="font-bold text-sm uppercase tracking-wide">
          Goroutine {goroutine.id}
        </span>
        {goroutine.parentId > 0 && (
          <span className="ml-2 text-xs font-mono">
            (parent: {goroutine.parentId})
          </span>
        )}
        {goroutine.exitedAt && (
          <span className="ml-auto text-xs font-mono text-red-600 font-bold">
            EXITED
          </span>
        )}
      </div>
      
      {/* States row */}
      <div className="relative h-10 border-b-2 border-black overflow-hidden">
        <div className="absolute left-0 top-1 right-0 bottom-1">
          {goroutine.states.map((state, idx) => (
            <StateBox
              key={idx}
              state={state}
              nextState={goroutine.states[idx + 1]}
              startTime={goroutine.createdAt}
            />
          ))}
        </div>
      </div>
      
      {/* Allocations row */}
      <div className="relative h-8 border-b-2 border-black overflow-hidden">
        <div className="absolute left-0 top-1 right-0 bottom-1">
          {goroutine.allocations.map((alloc, idx) => (
            <AllocBox
              key={idx}
              alloc={alloc}
              nextAlloc={goroutine.allocations[idx + 1]}
              startTime={goroutine.createdAt}
            />
          ))}
        </div>
      </div>
      
      {/* NewObject sub-row */}
      {goroutine.newobjects.length > 0 && (
        <div className="relative h-6 border-b-2 border-black bg-gray-50 overflow-hidden">
          <div className="absolute left-0 top-0.5 right-0 bottom-0.5">
            {goroutine.newobjects.map((obj, idx) => {
              const left = (obj.timestamp - goroutine.createdAt) / nanoseconds_per_pixel;
              const kindName = KIND_NAMES[obj.kind] || `Kind ${obj.kind}`;
              return (
                <div
                  key={idx}
                  className="absolute h-5 min-w-[12px] border-2 border-black bg-cyan-400 flex items-center justify-center text-[9px] font-mono font-bold overflow-hidden"
                  style={{ left: `${left}px`, width: `${Math.max(obj.size / 1000, 12)}px` }}
                  title={`newobject\nSize: ${obj.size} bytes\nKind: ${kindName}`}
                >
                  {obj.size > 100 ? `${obj.size}B` : '•'}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}


