import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Lock, Mail, ArrowLeft, RefreshCw, Key } from 'lucide-react';

export const Register = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [captchaId, setCaptchaId] = useState('');
    const [captchaSolution, setCaptchaSolution] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const [loading, setLoading] = useState(false);

    const fetchCaptcha = async () => {
        try {
            const res = await fetch('/api/captcha');
            if (res.ok) {
                const data = await res.json();
                setCaptchaId(data.captchaId);
            }
        } catch (e) {
            console.error("Failed to load captcha", e);
        }
    };

    useEffect(() => {
        fetchCaptcha();
    }, []);

    const refreshCaptcha = () => {
        setCaptchaSolution('');
        fetchCaptcha();
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const res = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password, captchaId, captchaSolution })
            });

            if (!res.ok) {
                const text = await res.text();
                throw new Error(text || 'Registration failed');
            }

            setSuccess(true);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    if (success) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div className="glass-panel p-8 w-full max-w-md text-center">
                    <h1 className="text-3xl font-bold mb-4 bg-clip-text text-transparent bg-gradient-to-r from-green-400 to-emerald-400">
                        Registration Successful!
                    </h1>
                    <p className="text-slate-300 mb-8">
                        Please check your email ({username}) to activate your account.
                    </p>
                    <Link
                        to="/login"
                        className="inline-block bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-2.5 px-6 rounded-lg transition-all"
                    >
                        Back to Login
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="flex items-center justify-center min-h-screen">
            <div className="glass-panel p-8 w-full max-w-md">
                <div className="flex items-center mb-8">
                    <Link to="/login" className="text-slate-400 hover:text-white transition-colors mr-4">
                        <ArrowLeft className="h-6 w-6" />
                    </Link>
                    <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-purple-400">
                        Create Account
                    </h1>
                </div>

                <form onSubmit={handleSubmit} className="space-y-6">
                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Email Address</label>
                        <div className="relative">
                            <Mail className="absolute left-3 top-3 h-5 w-5 text-slate-500" />
                            <input
                                type="email"
                                value={username}
                                onChange={e => setUsername(e.target.value)}
                                className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2.5 pl-10 pr-4 text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
                                placeholder="name@example.com"
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
                                placeholder="At least 8 characters"
                                minLength={8}
                                required
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-400 mb-2">Security Code</label>
                        <div className="flex gap-4">
                            <div className="relative flex-1">
                                <Key className="absolute left-3 top-3 h-5 w-5 text-slate-500" />
                                <input
                                    type="text"
                                    value={captchaSolution}
                                    onChange={e => setCaptchaSolution(e.target.value)}
                                    className="w-full bg-slate-800/50 border border-slate-700 rounded-lg py-2.5 pl-10 pr-4 text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all uppercase"
                                    placeholder="Enter code"
                                    required
                                />
                            </div>
                            <div className="flex items-center gap-2">
                                {captchaId && (
                                    <img
                                        src={`/api/captcha/${captchaId}.png`}
                                        alt="Captcha"
                                        className="h-11 rounded border border-slate-700 bg-white"
                                    />
                                )}
                                <button
                                    type="button"
                                    onClick={refreshCaptcha}
                                    className="p-2.5 bg-slate-800 hover:bg-slate-700 rounded-lg border border-slate-700 text-slate-400 hover:text-white transition-colors"
                                >
                                    <RefreshCw size={20} />
                                </button>
                            </div>
                        </div>
                    </div>

                    {error && <div className="text-red-400 text-sm">{error}</div>}

                    <button
                        type="submit"
                        disabled={loading}
                        className={`w-full bg-indigo-600 hover:bg-indigo-500 text-white font-medium py-2.5 rounded-lg transition-all transform active:scale-[0.98] ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
                    >
                        {loading ? 'Signing Up...' : 'Sign Up'}
                    </button>

                    <div className="text-center mt-4">
                        <span className="text-slate-400 text-sm">Already have an account? </span>
                        <Link to="/login" className="text-indigo-400 hover:text-indigo-300 text-sm font-medium transition-colors">
                            Sign In
                        </Link>
                    </div>
                </form>
            </div>
        </div>
    );
};
