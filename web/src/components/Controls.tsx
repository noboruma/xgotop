import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { useConfigStore } from '../store/configStore';
import { useEventStore } from '../store/eventStore';
import { useState } from 'react';

export function Controls() {
  const { nanoseconds_per_pixel, setNanosecondsPerPixel } = useConfigStore();
  const { viewport, setViewport, isConnected } = useEventStore();
  const [nsPerPixelInput, setNsPerPixelInput] = useState(nanoseconds_per_pixel.toString());
  
  const handleZoomIn = () => {
    setViewport({ zoom: viewport.zoom * 1.5 });
  };
  
  const handleZoomOut = () => {
    setViewport({ zoom: viewport.zoom / 1.5 });
  };
  
  const handleResetZoom = () => {
    setViewport({ zoom: 1, scrollX: 0 });
  };
  
  const handleNsPerPixelChange = () => {
    const value = parseFloat(nsPerPixelInput);
    if (!isNaN(value) && value > 0) {
      setNanosecondsPerPixel(value);
    }
  };
  
  return (
    <div className="brutalist-box p-4 mb-4">
      <div className="flex items-center gap-4 flex-wrap">
        {/* Connection status */}
        <div className="flex items-center gap-2">
          <div className={`w-3 h-3 border-2 border-black ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="font-bold text-sm uppercase">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
        
        {/* Zoom controls */}
        <div className="flex items-center gap-2">
          <span className="font-bold text-sm uppercase">Zoom:</span>
          <Button onClick={handleZoomOut}>-</Button>
          <span className="font-mono font-bold min-w-[60px] text-center">
            {viewport.zoom.toFixed(2)}x
          </span>
          <Button onClick={handleZoomIn}>+</Button>
          <Button onClick={handleResetZoom} variant="secondary">Reset</Button>
        </div>
        
        {/* Nanoseconds per pixel */}
        <div className="flex items-center gap-2">
          <span className="font-bold text-sm uppercase">NS/PX:</span>
          <Input
            type="number"
            value={nsPerPixelInput}
            onChange={(e) => setNsPerPixelInput(e.target.value)}
            onBlur={handleNsPerPixelChange}
            onKeyDown={(e) => e.key === 'Enter' && handleNsPerPixelChange()}
            className="w-32"
          />
        </div>
        
        {/* Stats */}
        <div className="ml-auto flex items-center gap-4 font-mono text-sm">
          <span>
            <span className="font-bold">Scroll:</span> {viewport.scrollX.toFixed(0)}px
          </span>
          <span>
            <span className="font-bold">Time Range:</span>{' '}
            {((viewport.timeEnd - viewport.timeStart) / 1000000).toFixed(2)}ms
          </span>
        </div>
      </div>
    </div>
  );
}


