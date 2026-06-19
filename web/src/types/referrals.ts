// src/types/referrals.ts

export const ALL_STATUSES = [
  { id: 'READY_TO_BOOK', label: 'Ready to Book' },
  { id: '1ST_CALL_COMPLETE', label: '1st Call Complete' },
  { id: '2ND_CALL_COMPLETE', label: '2nd Call Complete' },
  { id: '3RD_CALL_COMPLETE', label: '3rd Call Complete' },
  { id: 'BOOKED', label: 'Booked' },
  { id: 'UNABLE_TO_CONTACT', label: 'Unable to Contact' },
  { id: 'PATIENT_TO_CALL_BACK', label: 'Patient to Call Back' },
  { id: 'DECLINED', label: 'Declined' },
  { id: 'SUSPENDED', label: 'Suspended' },
  { id: 'CLOSED', label: 'Closed' }
] as const;

export const ALL_URGENCIES = ['ELECTIVE', 'URGENT', 'ASAP'] as const;
export const ALL_CONSULT_TYPES = ['APP+LE', 'APP+UE', 'APP+SX', 'SX', 'OTHER'] as const;
export const ALL_SOURCES = ['REGULAR', 'FRACTURE_CLINIC', 'OTHER'] as const;
export const ALL_BODY_PARTS = ['SHOULDER', 'KNEE', 'HIP', 'ELBOW', 'WRIST', 'ANKLE', 'FOOT', 'OTHER'] as const;
export const ALL_SIDES = ['LEFT', 'RIGHT', 'BILATERAL', 'OTHER'] as const;

export type ReferralStatus = typeof ALL_STATUSES[number]['id'];
export type ReferralUrgency = typeof ALL_URGENCIES[number];
export type ReferralConsultType = typeof ALL_CONSULT_TYPES[number];
export type ReferralSource = typeof ALL_SOURCES[number];

export interface FrontEndComplaint {
  bodyPart: string;
  side: string;
  details: string;
}

// FIX: Added [key: string] index signature fallback so loose strings can index it safely
export const STATUS_RULES: Record<ReferralStatus, ReferralStatus[]> & { [key: string]: ReferralStatus[] | undefined } = {
  READY_TO_BOOK: ['1ST_CALL_COMPLETE', 'UNABLE_TO_CONTACT', 'DECLINED', 'BOOKED', 'SUSPENDED'],
  '1ST_CALL_COMPLETE': ['2ND_CALL_COMPLETE', 'UNABLE_TO_CONTACT', 'DECLINED', 'BOOKED'],
  '2ND_CALL_COMPLETE': ['3RD_CALL_COMPLETE', 'UNABLE_TO_CONTACT', 'DECLINED', 'BOOKED'],
  '3RD_CALL_COMPLETE': ['UNABLE_TO_CONTACT', 'DECLINED', 'BOOKED'],
  PATIENT_TO_CALL_BACK: ['1ST_CALL_COMPLETE', 'BOOKED', 'DECLINED', 'SUSPENDED', 'CLOSED'],
  UNABLE_TO_CONTACT: ['READY_TO_BOOK'],
  BOOKED: [],
  DECLINED: [],
  SUSPENDED: [],
  CLOSED: [],
};
