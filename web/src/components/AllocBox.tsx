import { useConfigStore } from '../store/configStore';
import type { AllocationEvent } from '../types/event';

interface AllocBoxProps {
  alloc: AllocationEvent;
  nextAlloc?: AllocationEvent;
  startTime: number;
}

const KIND_NAMES: Record<number, string> = {
  0: 'Invalid', 1: 'bool', 2: 'int', 3: 'int8', 4: 'int16', 5: 'int32', 6: 'int64',
  7: 'uint', 8: 'uint8', 9: 'uint16', 10: 'uint32', 11: 'uint64', 12: 'uintptr',
  13: 'float32', 14: 'float64', 15: 'complex64', 16: 'complex128',
  17: 'Array', 18: 'Chan', 19: 'Func', 20: 'Interface', 21: 'Map', 22: 'Ptr',
  23: 'Slice', 24: 'String', 25: 'Struct', 26: 'UnsafePtr',
};

export function AllocBox({ alloc, nextAlloc, startTime }: AllocBoxProps) {
  const { nanoseconds_per_pixel, type_colors } = useConfigStore();
  
  // Calculate position and width
  const left = (alloc.timestamp - startTime) / nanoseconds_per_pixel;
  const duration = nextAlloc ? nextAlloc.timestamp - alloc.timestamp : 500000;
  const width = Math.max(duration / nanoseconds_per_pixel, 15); // Minimum 15px width
  
  const color = type_colors[alloc.type] || '#06b6d4';
  const kindName = KIND_NAMES[alloc.typeKind] || `Kind ${alloc.typeKind}`;
  
  let label = '';
  let info = '';
  
  if (alloc.type === 'makeslice') {
    label = `[]${kindName}`;
    info = `len:${alloc.length} cap:${alloc.capacity}`;
  } else if (alloc.type === 'makemap') {
    const elemKind = KIND_NAMES[alloc.typeKind2 || 0] || `Kind ${alloc.typeKind2}`;
    label = `map[${kindName}]${elemKind}`;
    info = `hint:${alloc.hint}`;
  }
  
  return (
    <div
      className="absolute h-6 border-2 border-black flex items-center justify-center overflow-hidden text-[10px] font-mono font-bold"
      style={{
        left: `${left}px`,
        width: `${width}px`,
        backgroundColor: color,
      }}
      title={`${alloc.type}\n${label}\n${info}`}
    >
      {width > 30 ? (
        <div className="flex flex-col items-center leading-tight">
          <span className="truncate">{label}</span>
          {width > 50 && <span className="text-[8px]">{info}</span>}
        </div>
      ) : (
        <span>•</span>
      )}
    </div>
  );
}


