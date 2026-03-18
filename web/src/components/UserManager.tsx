import { useState } from 'react';
import axios from 'axios';

export function UserManager() {
  const [newUser, setNewUser] = useState({ username: '', password: '', role: 'BOOKING_TEAM' });

  const handleAddUser = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const token = localStorage.getItem('token');
      await axios.post('/api/v1/users', newUser, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert(`User ${newUser.username} added!`);
      setNewUser({ username: '', password: '', role: 'BOOKING_TEAM' });
    } catch (err) {
      alert("Failed to add user. Ensure you are an Admin.");
    }
  };

  return (
    <section style={{ marginTop: '40px', borderTop: '2px solid #eee', paddingTop: '20px' }}>
      <h3>👥 Staff Management</h3>
      <form onSubmit={handleAddUser} style={{ display: 'flex', gap: '10px' }}>
        <input placeholder="Username" value={newUser.username} onChange={e => setNewUser({...newUser, username: e.target.value})} />
        <input type="password" placeholder="Password" value={newUser.password} onChange={e => setNewUser({...newUser, password: e.target.value})} />
        <select value={newUser.role} onChange={e => setNewUser({...newUser, role: e.target.value})}>
          <option value="BOOKING_TEAM">Booking Team</option>
          <option value="ADMIN">Admin</option>
        </select>
        <button type="submit">Create User</button>
      </form>
    </section>
  );
}
