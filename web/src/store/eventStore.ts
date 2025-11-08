import { create } from 'zustand';
import type { Event, GoroutineState } from '../types/event';

interface ViewportState {
  scrollX: number;
  zoom: number;
  timeStart: number;
  timeEnd: number;
}

interface EventStoreState {
  events: Event[];
  goroutines: Map<number, GoroutineState>;
  viewport: ViewportState;
  isConnected: boolean;
  
  addEvent: (event: Event) => void;
  addEvents: (events: Event[]) => void;
  clearEvents: () => void;
  setConnected: (connected: boolean) => void;
  setViewport: (viewport: Partial<ViewportState>) => void;
  getGoroutine: (id: number) => GoroutineState | undefined;
}

function processEvent(goroutines: Map<number, GoroutineState>, event: Event): Map<number, GoroutineState> {
  const newMap = new Map(goroutines);
  
  switch (event.event_type) {
    case 0: { // CasGStatus
      const goroutineId = event.attributes[2];
      const oldState = event.attributes[0];
      const newState = event.attributes[1];
      
      let goroutine = newMap.get(goroutineId);
      if (!goroutine) {
        goroutine = {
          id: goroutineId,
          parentId: event.parent_goroutine,
          states: [],
          allocations: [],
          newobjects: [],
          createdAt: event.timestamp,
        };
        newMap.set(goroutineId, goroutine);
      }
      
      // Calculate duration of previous state
      if (goroutine.states.length > 0) {
        const prevState = goroutine.states[goroutine.states.length - 1];
        prevState.duration = event.timestamp - prevState.timestamp;
      }
      
      goroutine.states.push({
        timestamp: event.timestamp,
        oldState,
        newState,
      });
      
      // Check if goroutine is dead
      if (newState === 6) { // G_STATUS_DEAD
        goroutine.exitedAt = event.timestamp;
      }
      break;
    }
    
    case 1: { // MakeSlice
      const goroutineId = event.goroutine;
      let goroutine = newMap.get(goroutineId);
      if (!goroutine) {
        goroutine = {
          id: goroutineId,
          parentId: event.parent_goroutine,
          states: [],
          allocations: [],
          newobjects: [],
          createdAt: event.timestamp,
        };
        newMap.set(goroutineId, goroutine);
      }
      
      goroutine.allocations.push({
        timestamp: event.timestamp,
        type: 'makeslice',
        typeKind: event.attributes[1],
        length: event.attributes[2],
        capacity: event.attributes[3],
      });
      break;
    }
    
    case 2: { // MakeMap
      const goroutineId = event.goroutine;
      let goroutine = newMap.get(goroutineId);
      if (!goroutine) {
        goroutine = {
          id: goroutineId,
          parentId: event.parent_goroutine,
          states: [],
          allocations: [],
          newobjects: [],
          createdAt: event.timestamp,
        };
        newMap.set(goroutineId, goroutine);
      }
      
      goroutine.allocations.push({
        timestamp: event.timestamp,
        type: 'makemap',
        typeKind: event.attributes[1], // key kind
        typeKind2: event.attributes[2], // elem kind
        hint: event.attributes[3],
      });
      break;
    }
    
    case 3: { // NewObject
      const goroutineId = event.goroutine;
      let goroutine = newMap.get(goroutineId);
      if (!goroutine) {
        goroutine = {
          id: goroutineId,
          parentId: event.parent_goroutine,
          states: [],
          allocations: [],
          newobjects: [],
          createdAt: event.timestamp,
        };
        newMap.set(goroutineId, goroutine);
      }
      
      goroutine.newobjects.push({
        timestamp: event.timestamp,
        size: event.attributes[0],
        kind: event.attributes[1],
      });
      break;
    }
    
    case 4: { // NewGoroutine
      const parentId = event.attributes[0];
      const newId = event.attributes[1];
      
      let goroutine = newMap.get(newId);
      if (!goroutine) {
        goroutine = {
          id: newId,
          parentId: parentId,
          states: [],
          allocations: [],
          newobjects: [],
          createdAt: event.timestamp,
        };
        newMap.set(newId, goroutine);
      }
      break;
    }
    
    case 5: { // GoExit
      const goroutineId = event.attributes[0];
      const goroutine = newMap.get(goroutineId);
      if (goroutine) {
        goroutine.exitedAt = event.timestamp;
      }
      break;
    }
  }
  
  return newMap;
}

export const useEventStore = create<EventStoreState>((set, get) => ({
  events: [],
  goroutines: new Map(),
  viewport: {
    scrollX: 0,
    zoom: 1,
    timeStart: 0,
    timeEnd: 0,
  },
  isConnected: false,
  
  addEvent: (event) => {
    set((state) => {
      const events = [...state.events, event];
      const goroutines = processEvent(state.goroutines, event);
      
      // Update time bounds
      const timeEnd = Math.max(state.viewport.timeEnd, event.timestamp);
      const timeStart = state.viewport.timeStart === 0 ? event.timestamp : state.viewport.timeStart;
      
      return {
        events,
        goroutines,
        viewport: {
          ...state.viewport,
          timeStart,
          timeEnd,
        },
      };
    });
  },
  
  addEvents: (newEvents) => {
    set((state) => {
      let goroutines = state.goroutines;
      for (const event of newEvents) {
        goroutines = processEvent(goroutines, event);
      }
      
      const events = [...state.events, ...newEvents];
      
      // Update time bounds
      const timestamps = newEvents.map(e => e.timestamp);
      const timeEnd = Math.max(state.viewport.timeEnd, ...timestamps);
      const timeStart = state.viewport.timeStart === 0 && timestamps.length > 0
        ? Math.min(...timestamps)
        : state.viewport.timeStart;
      
      return {
        events,
        goroutines,
        viewport: {
          ...state.viewport,
          timeStart,
          timeEnd,
        },
      };
    });
  },
  
  clearEvents: () => {
    set({
      events: [],
      goroutines: new Map(),
      viewport: {
        scrollX: 0,
        zoom: 1,
        timeStart: 0,
        timeEnd: 0,
      },
    });
  },
  
  setConnected: (connected) => {
    set({ isConnected: connected });
  },
  
  setViewport: (viewport) => {
    set((state) => ({
      viewport: { ...state.viewport, ...viewport },
    }));
  },
  
  getGoroutine: (id) => {
    return get().goroutines.get(id);
  },
}));

