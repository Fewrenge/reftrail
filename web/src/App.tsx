import { useEffect, useState } from 'react';
import axios from 'axios';
import type { WLEntry} from './types'; 
import { UserManager } from './components/UserManager';
import { Timeline } from './components/Timeline';

function App() {
  // --- Navigation & Auth State ---
  const [view, setView] = useState<'waitlist' | 'users'>('waitlist');
  const [token, setToken] = useState(localStorage.getItem('token') || '');
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('password123');
  
  // --- Data State ---
  const [entries, setEntries] = useState<WLEntry[]>([]);
  const [showLogs, setShowLogs] = useState<number | null>(null); // Track which row shows history

  const [newPatient, setNewPatient] = useState({
    patientName: '',
    patientDob: '',
    urgency: 'Elective',
    complaint: '',
    referringPhysician: '',
    state: 'READY_TO_BOOK'
  });

  // 1. CREATE Logic
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const currentToken = localStorage.getItem('token');
      await axios.post('/api/v1/waitlist', newPatient, {
        headers: { Authorization: `Bearer ${currentToken}` }
      });
      setNewPatient({ 
        patientName: '', patientDob: '', urgency: 'Elective', 
        complaint: '', referringPhysician: '', state: 'READY_TO_BOOK' 
      });
      fetchWaitlist(currentToken!); 
      alert("Patient added successfully!");
    } catch (err) {
      alert("Failed to add patient. Check the console.");
    }
  };

  // 2. LOGIN Logic
  const handleLogin = async () => {
    try {
      const res = await axios.post('/api/v1/login', { username, password });
      localStorage.setItem('token', res.data.token);
      setToken(res.data.token);
      fetchWaitlist(res.data.token);
    } catch (err) {
      alert("Login failed! Check your credentials.");
    }
  };

  // 3. UPDATE Logic (With Force/Hierarchy check)
  const handleStateChange = async (id: number, newState: string, isForced: boolean = false) => {
    try {
      const currentToken = localStorage.getItem('token');
      await axios.patch(`/api/v1/waitlist/${id}`, 
        { state: newState, force: isForced }, 
        { headers: { Authorization: `Bearer ${currentToken}` } }
      );
      fetchWaitlist(currentToken!);
    } catch (err: any) {
      // The "Brain" (Go) sends this error if hierarchy is violated
      if (err.response?.data.includes("force flag")) {
        if (window.confirm("This is a backward move. Force it? (Admin only)")) {
          handleStateChange(id, newState, true); 
        }
      } else {
        alert(err.response?.data || "Update failed");
      }
    }
  };

  // 4. FETCH Logic
  const fetchWaitlist = async (authToken: string) => {
    try {
      const res = await axios.get('/api/v1/waitlist', {
        headers: { 'Authorization': `Bearer ${authToken}` }
      });
      setEntries(res.data || []);
    } catch (err: any) {
      if (err.response?.status === 401) {
        localStorage.removeItem('token');
        setToken('');
      }
    }
  };

  useEffect(() => {
    if (token) fetchWaitlist(token);
  }, [token]);

  // --- RENDER LOGIN ---
  if (!token) {
    return (
      <div style={{ padding: '50px', textAlign: 'center' }}>
        <h2>🔐 Staff Login</h2>
        <input value={username} onChange={e => setUsername(e.target.value)} placeholder="Username" /><br/><br/>
        <input type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="Password" /><br/><br/>
        <button onClick={handleLogin}>Sign In</button>
      </div>
    );
  }

  // --- RENDER DASHBOARD ---
  return (
    <div style={{ padding: '20px', fontFamily: 'sans-serif', maxWidth: '1200px', margin: '0 auto' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '2px solid #eee', paddingBottom: '10px', marginBottom: '20px' }}>
        <h1>🏥 Waitlist Management</h1>
        <nav style={{ display: 'flex', gap: '20px' }}>
          <button onClick={() => setView('waitlist')} style={{ fontWeight: view === 'waitlist' ? 'bold' : 'normal', background: 'none', border: 'none', cursor: 'pointer' }}>📋 Waitlist</button>
          <button onClick={() => setView('users')} style={{ fontWeight: view === 'users' ? 'bold' : 'normal', background: 'none', border: 'none', cursor: 'pointer' }}>👥 Staff</button>
          <button onClick={() => { localStorage.removeItem('token'); setToken(''); }}>Logout</button>
        </nav>
      </header>

      {view === 'waitlist' ? (
        <>
          {/* New Patient Form Section */}
          <section style={{ background: '#f9f9f9', padding: '20px', borderRadius: '8px', marginBottom: '30px', border: '1px solid #ddd' }}>
            <h3>➕ Add New Referral</h3>
            <form onSubmit={handleCreate} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px' }}>
              <input placeholder="Patient Name" value={newPatient.patientName} onChange={e => setNewPatient({...newPatient, patientName: e.target.value})} required />
              <input type="date" value={newPatient.patientDob} onChange={e => setNewPatient({...newPatient, patientDob: e.target.value})} required />
              <select value={newPatient.urgency} onChange={e => setNewPatient({...newPatient, urgency: e.target.value})}>
                <option value="Elective">Elective</option>
                <option value="Urgent">Urgent</option>
                <option value="ASAP">ASAP</option>
              </select>
              <input placeholder="Referring Physician" value={newPatient.referringPhysician} onChange={e => setNewPatient({...newPatient, referringPhysician: e.target.value})} />
              <textarea style={{ gridColumn: 'span 2' }} placeholder="Clinical Complaint (e.g. Left Knee)" value={newPatient.complaint} onChange={e => setNewPatient({...newPatient, complaint: e.target.value})} />
              <button type="submit" style={{ gridColumn: 'span 2', background: '#007bff', color: 'white', padding: '10px', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
                Add to Waitlist
              </button>
            </form>
          </section>

          {/* Waitlist Table Section */}
          <table border={1} cellPadding={10} style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: '#f4f4f4' }}>
                <th>Patient</th>
                <th>Urgency</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {entries.length > 0 ? entries.map(e => (
                <tr key={e.id}>
                  <td>
                    <strong>{e.patientName}</strong><br/>
                    <small>{e.patientDob} | {e.referringPhysician}</small>
                  </td>
                  <td>
                    <span style={{ color: e.urgency === 'ASAP' ? 'red' : 'orange', fontWeight: 'bold' }}>{e.urgency}</span>
                  </td>
                  <td>
                    <select 
                      value={e.state} 
                      onChange={(opt) => handleStateChange(e.id, opt.target.value)}
                      style={{ padding: '4px', borderRadius: '4px', border: '1px solid #ccc' }}
                    >
                      <option value="READY_TO_BOOK">Ready to Book</option>
                      <option value="1ST_CALL">1st Call Attempted</option>
                      <option value="2ND_CALL">2nd Call Attempted</option>
                      <option value="3RD_CALL">3rd Call Attempted</option>
                      <option value="BOOKED">✅ Booked</option>
                    </select>
                  </td>
                  <td>
                    <button onClick={() => setShowLogs(showLogs === e.id ? null : e.id)}>
                        {showLogs === e.id ? 'Close' : 'History'}
                        </button>
                         {/* Wrap the timeline so it stays "Inside" the action column */}
                         {showLogs === e.id && (
                          <div style={{ position: 'absolute', zIndex: 10, background: 'white', boxShadow: '0 4px 8px rgba(0,0,0,0.1)', minWidth: '250px' }}>
                            <Timeline entryId={e.id} />
                          </div>)}
                  </td>
                </tr>
              )) : <tr><td colSpan={4}>No patients found.</td></tr>}
            </tbody>
          </table>
        </>
      ) : (
        <UserManager />
      )}
    </div>
  );
}

export default App;
