import React, { useState, useEffect } from 'react';
import { setAuth } from './auth';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { Lock, User, ChevronUp } from 'lucide-react';

export const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');
    const [loading, setLoading] = useState(false);
    const [version, setVersion] = useState('');
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        const params = new URLSearchParams(location.search);
        if (params.get('activated') === 'true') {
            setMessage('Account activated successfully! You can now log in.');
        }
    }, [location]);

    useEffect(() => {
        fetch('/api/app-info')
            .then(res => res.json())
            .then(data => setVersion(data.version))
            .catch(err => console.error('Failed to fetch version:', err));
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setMessage('');
        setLoading(true);

        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            if (!res.ok) {
                const text = await res.text();
                throw new Error(text.replace(/\n/g, '') || 'Invalid credentials');
            }

            const data = await res.json();
            setAuth(data.token, { username: data.username, is_admin: data.is_admin });
            navigate('/');
        } catch (err: any) {
            setError(err.message);
            setLoading(false);
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen">
            <div className="glass-panel p-8 w-full max-w-md">
                <h1 className="text-3xl font-bold mb-2 text-center bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-purple-400">
                    Deep History on MultiversX
                </h1>
                <p className="text-center text-slate-400 text-sm mb-8 font-light italic">
                    Uncover the past, analyze the future
                </p>

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
                                autoCapitalize="none"
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
                        disabled={loading}
                        className={`w-full bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-2.5 rounded-lg transition-all transform active:scale-[0.98] ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
                    >
                        {loading ? 'Signing In...' : 'Sign In'}
                    </button>

                    <div className="text-center mt-4">
                        <span className="text-slate-400 text-sm">Don't have an account? </span>
                        <Link to="/register" className="text-indigo-400 hover:text-indigo-300 text-sm font-medium transition-colors">
                            Register
                        </Link>
                    </div>
                </form>

                <div className="mt-6 text-center">
                    <p style={{ fontSize: '0.8rem' }} className="text-slate-500 flex items-center justify-center gap-1">
                        Build {version} |
                        <div className="relative group inline-block ml-1">
                            <button className="hover:text-slate-300 underline decoration-slate-600 underline-offset-2 flex items-center gap-1 transition-colors">
                                Source Code <ChevronUp size={12} />
                            </button>
                            <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-56 bg-slate-900/95 border border-slate-700/50 backdrop-blur-xl rounded-lg shadow-2xl opacity-0 invisible group-hover:visible group-hover:opacity-100 transition-all duration-200 transform origin-bottom z-50 flex flex-col">
                                <div className="px-4 py-2 border-b border-white/5 text-[10px] uppercase font-bold text-slate-500 tracking-wider">
                                    Repositories
                                </div>
                                <a href="https://github.com/iulianpascalau/mx-epoch-proxy-go" target="_blank" rel="noopener noreferrer" className="px-4 py-3 hover:bg-indigo-500/10 hover:text-indigo-300 text-slate-300 text-xs text-left transition-colors flex items-center gap-2">
                                    <span className="w-1.5 h-1.5 rounded-full bg-indigo-500"></span> Epoch Proxy (Go)
                                </a>
                                <a href="https://github.com/iulianpascalau/mx-crypto-payments" target="_blank" rel="noopener noreferrer" className="px-4 py-3 hover:bg-emerald-500/10 hover:text-emerald-300 text-slate-300 text-xs text-left transition-colors flex items-center gap-2">
                                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> Crypto Payments (Go)
                                </a>
                                <a href="https://github.com/iulianpascalau/mx-credits-contract-rs" target="_blank" rel="noopener noreferrer" className="px-4 py-3 hover:bg-amber-500/10 hover:text-amber-300 text-slate-300 text-xs text-left transition-colors flex items-center gap-2 rounded-b-lg">
                                    <span className="w-1.5 h-1.5 rounded-full bg-amber-500"></span> Credits Contract (Rust)
                                </a>
                                <div className="absolute -bottom-1 left-1/2 -translate-x-1/2 w-2 h-2 bg-slate-900 border-r border-b border-slate-700/50 transform rotate-45"></div>
                            </div>
                        </div>
                    </p>
                </div>
            </div>
        </div>
    );
};
