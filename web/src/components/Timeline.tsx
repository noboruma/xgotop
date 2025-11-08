import { useEffect, useRef } from 'react';
import { useEventStore } from '../store/eventStore';
import { GoroutineRow } from './GoroutineRow';
import { TimelineRuler } from './TimelineRuler';

export function Timeline() {
  const { goroutines, viewport, setViewport } = useEventStore();
  const containerRef = useRef<HTMLDivElement>(null);
  
  // Handle scroll
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    
    const handleScroll = () => {
      setViewport({ scrollX: container.scrollLeft });
    };
    
    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [setViewport]);
  
  // Handle zoom with mouse wheel
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    
    const handleWheel = (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        const delta = e.deltaY > 0 ? 0.9 : 1.1;
        setViewport({ zoom: viewport.zoom * delta });
      }
    };
    
    container.addEventListener('wheel', handleWheel, { passive: false });
    return () => container.removeEventListener('wheel', handleWheel);
  }, [viewport.zoom, setViewport]);
  
  // Convert Map to array and sort by goroutine ID
  const goroutineArray = Array.from(goroutines.values()).sort((a, b) => a.id - b.id);
  
  // Calculate total timeline width
  const timelineWidth = Math.max(
    window.innerWidth * 2,
    ((viewport.timeEnd - viewport.timeStart) / 1000000) * viewport.zoom
  );
  
  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <TimelineRuler />
      
      <div
        ref={containerRef}
        className="flex-1 overflow-auto"
      >
        <div style={{ width: `${timelineWidth}px`, minHeight: '100%' }}>
          {goroutineArray.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <div className="brutalist-box p-8">
                <p className="font-bold text-lg uppercase text-center">
                  No goroutines yet
                </p>
                <p className="text-sm text-center mt-2">
                  Waiting for events...
                </p>
              </div>
            </div>
          ) : (
            goroutineArray.map((goroutine) => (
              <GoroutineRow key={goroutine.id} goroutine={goroutine} />
            ))
          )}
        </div>
      </div>
    </div>
  );
}


