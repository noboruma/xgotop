import { useEffect, useState } from 'react';
import { Button } from './ui/Button';
import { apiClient } from '../services/api';
import type { Session } from '../types/event';
import { useEventStore } from '../store/eventStore';

interface SessionPickerProps {
  onSessionSelect: (sessionId: string) => void;
}

export function SessionPicker({ onSessionSelect }: SessionPickerProps) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const { clearEvents } = useEventStore();
  
  useEffect(() => {
    if (isOpen) {
      loadSessions();
    }
  }, [isOpen]);
  
  const loadSessions = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiClient.getSessions();
      setSessions(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sessions');
    } finally {
      setLoading(false);
    }
  };
  
  const handleSessionClick = async (sessionId: string) => {
    clearEvents();
    onSessionSelect(sessionId);
    setIsOpen(false);
  };
  
  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };
  
  if (!isOpen) {
    return (
      <Button onClick={() => setIsOpen(true)} variant="secondary">
        Load Session
      </Button>
    );
  }
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="brutalist-box bg-background p-6 max-w-2xl w-full max-h-[80vh] overflow-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold uppercase">Sessions</h2>
          <Button onClick={() => setIsOpen(false)}>Close</Button>
        </div>
        
        {loading && (
          <div className="text-center py-8">
            <p className="font-bold">Loading sessions...</p>
          </div>
        )}
        
        {error && (
          <div className="border-3 border-red-500 bg-red-100 p-4 mb-4">
            <p className="font-bold text-red-700">{error}</p>
          </div>
        )}
        
        {!loading && !error && sessions.length === 0 && (
          <div className="text-center py-8">
            <p className="font-bold">No sessions found</p>
          </div>
        )}
        
        {!loading && sessions.length > 0 && (
          <div className="space-y-2">
            {sessions.map((session) => (
              <div
                key={session.id}
                className="brutalist-box p-4 cursor-pointer hover:translate-x-1 hover:translate-y-1 transition-transform"
                onClick={() => handleSessionClick(session.id)}
              >
                <div className="flex items-start justify-between">
                  <div>
                    <p className="font-bold text-sm font-mono">{session.id}</p>
                    <p className="text-xs mt-1">
                      <span className="font-bold">Started:</span> {formatDate(session.start_time)}
                    </p>
                    {session.end_time && (
                      <p className="text-xs">
                        <span className="font-bold">Ended:</span> {formatDate(session.end_time)}
                      </p>
                    )}
                    <p className="text-xs">
                      <span className="font-bold">Binary:</span> {session.binary_path}
                    </p>
                    {session.pid && (
                      <p className="text-xs">
                        <span className="font-bold">PID:</span> {session.pid}
                      </p>
                    )}
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-lg">{session.event_count}</p>
                    <p className="text-xs uppercase">events</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}


