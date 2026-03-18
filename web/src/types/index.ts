export interface WLEntry {
  id: number;
  patientName: string;
  patientDob: string;
  urgency: 'Elective' | 'Urgent' | 'ASAP';
  state: string;
  referringPhysician: string;
  complaint: string;
  createdTs: number;
}

export interface LoginResponse {
  user: {
    id: number;
    username: string;
    role: string;
  };
  token: string;
}

export interface WLLog {
  id: number;
  entryId: number;
  userId: number;
  oldState: string;
  newState: string;
  note: string;
  createdTs: number;
}

export interface User {
  id: number;
  username: string;
  role: 'ADMIN' | 'BOOKING_TEAM';
}