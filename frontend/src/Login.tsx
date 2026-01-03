import React, { useState, useEffect } from 'react';
import { setAuth } from './auth';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { Lock, User } from 'lucide-react';

export const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        const params = new URLSearchParams(location.search);
        if (params.get('activated') === 'true') {
            setMessage('Account activated successfully! You can now log in.');
        }
    }, [location]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');

        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            if (!res.ok) throw new Error('Invalid credentials');

            const data = await res.json();
            setAuth(data.token, { username: data.username, is_admin: data.is_admin });
            navigate('/');
        } catch (err: any) {
            setError(err.message);
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen">
            <div className="glass-panel p-8 w-full max-w-md">
                <h1 className="text-3xl font-bold mb-8 text-center bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-purple-400">
                    Epoch Proxy
                </h1>

                {message && (
                    <div className="mb-6 p-3 bg-green-500/20 border border-green-500/50 rounded-lg text-green-200 text-sm text-center">
                        {message}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Username</label>
                        <div className="relative">
                            <User className="absolute left-3 top-3 h-5 w-5 text-slate-500" />
                            <input
                                type="text"
                                value={username}
                                onChange={e => setUsername(e.target.value)}
                                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2.5 pl-10 pr-4 text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
                                placeholder="Enter username"
                                required
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Password</label>
                        <div className="relative">
                            <Lock className="absolute left-3 top-3 h-5 w-5 text-slate-500" />
                            <input
                                type="password"
                                value={password}
                                onChange={e => setPassword(e.target.value)}
                                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2.5 pl-10 pr-4 text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
                                placeholder="••••••••"
                                required
                            />
                        </div>
                    </div>

                    {error && <div className="text-red-400 text-sm text-center">{error}</div>}

                    <button
                        type="submit"
                        className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-2.5 rounded-lg transition-all transform active:scale-[0.98]"
                    >
                        Sign In
                    </button>

                    <div className="text-center mt-4">
                        <span className="text-slate-400 text-sm">Don't have an account? </span>
                        <Link to="/register" className="text-indigo-400 hover:text-indigo-300 text-sm font-medium transition-colors">
                            Register
                        </Link>
                    </div>
                </form>
            </div>
        </div>
    );
};
