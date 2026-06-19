import React from 'react';
import {useState, useEffect} from 'react';
import { ReferralPhysicianCard, } from '@/components/ReferralPhysician/ReferralPhysicianCard';
import type {ReferralPhysician} from '@/components/ReferralPhysician/ReferralPhysicianCard';

export const Physicians: React.FC = () => {
  const [physiciansList, setPhysiciansList] = useState<ReferralPhysician[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetch('/api/v1/physicians', {method: 'GET'})
      .then((res) => res.json())
      .then((data: ReferralPhysician[]) => {
        setPhysiciansList(data);
        setIsLoading(false);
      })
      .catch((err) => {
        console.error("Failed to load physicians:", err);
        setIsLoading(false);
      });
  }, []);

  const handleSelect = (physician: ReferralPhysician) => {
    console.log('Selected physician:', physician.id);
  };

  if (isLoading) return <div className="p-6">Loading physicians...</div>;

  return (
    <div className="p-6 bg-slate-50 min-h-screen">
      <h1 className="text-2xl font-bold text-slate-900 mb-6">Physicians</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {physiciansList.map((physician) => (
          <ReferralPhysicianCard 
            key={physician.id} 
            physician={physician} 
            onClick={handleSelect}
          />
        ))}
      </div>
    </div>
  );
};