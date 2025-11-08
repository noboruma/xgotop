import { useEffect, useState, useCallback } from 'react';
import { Controls } from './components/Controls';
import { Timeline } from './components/Timeline';
import { SessionPicker } from './components/SessionPicker';
import { useEventStore } from './store/eventStore';
import { WebSocketClient } from './services/websocket';
import { apiClient } from './services/api';
import { Button } from './components/ui/Button';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

function App() {
  const { addEvent, addEvents, setConnected, clearEvents } = useEventStore();
  const [wsClient] = useState(() => new WebSocketClient(WS_URL));
  const [mode, setMode] = useState<'live' | 'replay'>('live');
  const [replaySessionId, setReplaySessionId] = useState<string | null>(null);
  
  // Setup WebSocket for live mode
  useEffect(() => {
    if (mode === 'live') {
      wsClient.connect();
      
      const unsubscribe = wsClient.onEvent((event) => {
        addEvent(event);
      });
      
      // Check connection status periodically
      const interval = setInterval(() => {
        setConnected(wsClient.isConnected());
      }, 1000);
      
      return () => {
        unsubscribe();
        clearInterval(interval);
        wsClient.disconnect();
        setConnected(false);
      };
    }
  }, [mode, wsClient, addEvent, setConnected]);
  
  // Load replay session
  const loadReplaySession = useCallback(async (sessionId: string) => {
    try {
      setMode('replay');
      setReplaySessionId(sessionId);
      clearEvents();
      
      // Load all events from the session
      const events = await apiClient.getEvents(sessionId, { limit: 100000 });
      addEvents(events);
    } catch (error) {
      console.error('Failed to load replay session:', error);
      alert('Failed to load session: ' + (error instanceof Error ? error.message : 'Unknown error'));
    }
  }, [addEvents, clearEvents]);
  
  const switchToLive = () => {
    clearEvents();
    setMode('live');
    setReplaySessionId(null);
  };
  
  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Header */}
      <header className="border-b-4 border-black p-4 bg-primary text-primary-foreground">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-black uppercase tracking-wider">
            xgotop
          </h1>
          <div className="flex items-center gap-4">
            {mode === 'replay' && (
              <div className="brutalist-box px-4 py-2 bg-yellow-400 text-black">
                <span className="font-bold uppercase text-sm">
                  Replay Mode: {replaySessionId?.slice(0, 8)}...
                </span>
              </div>
            )}
            <Button onClick={switchToLive} disabled={mode === 'live'}>
              Live Mode
            </Button>
            <SessionPicker onSessionSelect={loadReplaySession} />
          </div>
        </div>
      </header>
      
      {/* Controls */}
      <div className="p-4">
        <Controls />
      </div>
      
      {/* Timeline */}
      <Timeline />
      
      {/* Footer */}
      <footer className="border-t-4 border-black p-2 bg-secondary text-center text-xs font-mono">
        <p>
          Use <span className="font-bold">Ctrl+Scroll</span> to zoom • <span className="font-bold">Scroll</span> to pan
        </p>
      </footer>
    </div>
  );
}

export default App;
