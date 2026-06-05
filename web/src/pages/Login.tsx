import React, { useState } from 'react';

export default function Login({ onLoginSuccess }: { onLoginSuccess: () => void }) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    const res = await fetch('/api/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    if (res.ok) {
      onLoginSuccess();
    } else {
      alert("Login Failed!");
    }
  };

  return (
    <div className="py-4 sm:py-8 w-80 max-w-full min-h-svh mx-auto flex flex-col justify-start items-center">
      <div className="w-full py-4 grow flex flex-col justify-center items-center">
        
        {/* Header Section */}
        <div className="w-full flex flex-row justify-center items-center mb-6">
          <p className="ml-2 text-5xl text-foreground opacity-80 font-tight">RefTrail</p>
        </div>

        {/* Form Section */}
        <form onSubmit={handleSubmit} className="w-full space-y-4">
          <div className="w-full">
            <input 
              type="text" 
              placeholder="Username"
              className="w-full bg-background border border-slate-200 rounded-xl px-4 py-3 outline-none focus:border-blue-500 transition-all"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>
          <div className="w-full">
            <input 
              type="password" 
              placeholder="Password"
              className="w-full bg-background border border-slate-200 rounded-xl px-4 py-3 outline-none focus:border-blue-500 transition-all"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          
          <button className="w-full bg-blue-600 text-white py-3 rounded-xl font-bold hover:opacity-90 transition-all shadow-sm">
            Sign In
          </button>
        </form>

        {/* Optional Sign Up Link */}
        <p className="w-full mt-4 text-sm text-left">
          <span className="text-muted-foreground">Don't have an account?</span>
          <span className="cursor-pointer ml-2 text-blue-600 hover:underline">Contact the System Admin</span>
        </p>
      </div>
      
    </div>
  );
}
