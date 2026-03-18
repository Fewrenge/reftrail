// web/src/components/Timeline.tsx
import { useEffect, useState } from 'react';
import axios from 'axios';
import type { WLLog } from '../types';

export function Timeline({ entryId }: { entryId: number }) {
  const [logs, setLogs] = useState<WLLog[]>([]);

  useEffect(() => {
    const fetchLogs = async () => {
      const token = localStorage.getItem('token');
      const res = await axios.get(`/api/v1/waitlist/${entryId}/logs`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setLogs(res.data || []); // Ensure we have an array
    };
    fetchLogs();
  }, [entryId]);

  return (
    <div style={{ padding: '10px', fontSize: '0.85rem' }}>
      <h5 style={{ margin: '0 0 8px 0' }}>📜 Audit History</h5>
      {logs.length === 0 ? (
        <p style={{ color: '#888' }}>No history found. Try changing the status!</p>
      ) : (
        <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
          {logs.map((log) => (
            <li key={log.id} style={{ marginBottom: '6px', borderBottom: '1px solid #eee', paddingBottom: '4px' }}>
              <span style={{ fontWeight: 'bold' }}>{log.oldState} → {log.newState}</span>
              <br />
              <small style={{ color: '#666' }}>
                User #{log.userId} at {new Date(log.createdTs * 1000).toLocaleTimeString()}
              </small>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
