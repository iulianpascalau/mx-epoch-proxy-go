import { useEffect, useState } from 'react';
import { getAccessKey, clearAuth, getUserInfo, parseJwt, type User as AuthUser } from './auth';
import { useNavigate } from 'react-router-dom';
import { LogOut, Lock, BookOpen, UserCog } from 'lucide-react';
import axios from 'axios';

export const Settings = () => {
    const navigate = useNavigate();
    const [user, setUser] = useState<AuthUser | null>(null);
    const [appInfo, setAppInfo] = useState<{ version: string, backend: string } | null>(null);

    // Security Settings State
    const [passState, setPassState] = useState({ oldPass: '', newPass: '', confirmPass: '' });
    const [emailState, setEmailState] = useState({ oldPass: '', newEmail: '', confirmEmail: '' });

    useEffect(() => {
        const userInfo = getUserInfo();
        if (!userInfo) {
            navigate('/login');
            return;
        }

        const token = getAccessKey();
        if (token) {
            const decoded = parseJwt(token);
            if (decoded) {
                userInfo.is_admin = decoded.is_admin;
            }
        }
        setUser(userInfo);

        // Fetch App Info (for Swagger/Version)
        const fetchAppInfo = async () => {
            const token = getAccessKey();
            if (!token) return;
            try {
                const configRes = await axios.get('/api/config', { headers: { Authorization: `Bearer ${token}` } });
                if (configRes.data) {
                    setAppInfo({
                        version: configRes.data.version || 'v1.0.0',
                        backend: configRes.data.backend_url || window.location.origin
                    });
                }
            } catch (e) {
                console.error("Failed to fetch app config", e);
            }
        };
        fetchAppInfo();

    }, [navigate]);

    const handleLogout = () => {
        clearAuth();
        navigate('/login');
    };

    const handleChangePassword = async (e: React.FormEvent) => {
        e.preventDefault();
        if (passState.newPass !== passState.confirmPass) {
            alert("New passwords do not match!");
            return;
        }
        try {
            await axios.post('/api/change-password', {
                oldPassword: passState.oldPass,
                newPassword: passState.newPass
            }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });

            alert("Password updated successfully.");
            setPassState({ oldPass: '', newPass: '', confirmPass: '' });
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to update password';
            alert(msg);
        }
    };

    const handleChangeEmail = async (e: React.FormEvent) => {
        e.preventDefault();
        if (emailState.newEmail !== emailState.confirmEmail) {
            alert("New email addresses do not match!");
            return;
        }
        try {
            await axios.post('/api/request-email-change', {
                oldPassword: emailState.oldPass,
                newEmail: emailState.newEmail
            }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });

            alert("Confirmation email sent to the new address. Please check your inbox to finalize the change.");
            setEmailState({ oldPass: '', newEmail: '', confirmEmail: '' });
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to request email change';
            alert(msg);
        }
    };

    if (!user) return null;

    return (
        <div className="min-h-screen p-6 md:p-12 max-w-4xl mx-auto">
            {/* Header */}
            <div className="glass-panel p-6 mb-8 flex justify-between items-center">
                <div className="flex items-center gap-4">
                    <div className="h-10 w-10 rounded-full bg-indigo-500/20 flex items-center justify-center text-indigo-400">
                        <UserCog size={20} />
                    </div>
                    <div>
                        <h1 className="text-xl font-bold">User Settings</h1>
                        <span className="text-sm text-slate-400">Manage your security and account details</span>
                    </div>
                </div>
                <div className="flex gap-3">
                    <button onClick={() => navigate('/')} className="px-4 py-2 hover:bg-white/5 rounded-lg transition-colors text-slate-300">
                        Back to Dashboard
                    </button>
                    <button onClick={handleLogout} className="flex items-center gap-2 px-4 py-2 rounded-lg hover:bg-white/5 transition-colors text-slate-300">
                        <LogOut size={18} />
                        <span>Sign Out</span>
                    </button>
                </div>
            </div>

            <div className="flex flex-col gap-8">
                {/* Security Settings Panel */}
                <div className="glass-panel p-6">
                    <h2 className="text-xl font-semibold flex items-center gap-2 mb-6">
                        <Lock className="text-indigo-400" /> Security Settings
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        {/* Change Password */}
                        <div className="bg-white/5 rounded-lg p-6 border border-white/5">
                            <h3 className="text-lg font-medium mb-4 text-slate-200">Change Password</h3>
                            <form onSubmit={handleChangePassword} className="space-y-4">
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Current Password</label>
                                    <input
                                        type="password" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={passState.oldPass}
                                        onChange={e => setPassState({ ...passState, oldPass: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">New Password</label>
                                    <input
                                        type="password" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={passState.newPass}
                                        onChange={e => setPassState({ ...passState, newPass: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Confirm New Password</label>
                                    <input
                                        type="password" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={passState.confirmPass}
                                        onChange={e => setPassState({ ...passState, confirmPass: e.target.value })}
                                    />
                                </div>
                                <div className="pt-2">
                                    <button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors">
                                        Update Password
                                    </button>
                                </div>
                            </form>
                        </div>

                        {/* Change Email */}
                        <div className="bg-white/5 rounded-lg p-6 border border-white/5">
                            <h3 className="text-lg font-medium mb-4 text-slate-200">Change Email Address</h3>
                            <form onSubmit={handleChangeEmail} className="space-y-4">
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Current Password</label>
                                    <input
                                        type="password" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={emailState.oldPass}
                                        onChange={e => setEmailState({ ...emailState, oldPass: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">New Email Address</label>
                                    <input
                                        type="email" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={emailState.newEmail}
                                        onChange={e => setEmailState({ ...emailState, newEmail: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Confirm New Email</label>
                                    <input
                                        type="email" required
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                        value={emailState.confirmEmail}
                                        onChange={e => setEmailState({ ...emailState, confirmEmail: e.target.value })}
                                    />
                                </div>
                                <div className="pt-2">
                                    <button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors">
                                        Request Email Change
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>

                {/* API Info Panel */}
                {appInfo && (
                    <div className="glass-panel p-6">
                        <h2 className="text-xl font-semibold flex items-center gap-2 mb-6">
                            <BookOpen className="text-indigo-400" /> API Documentation
                        </h2>
                        <div className="flex flex-col items-center justify-center gap-6 py-4">
                            <div className="text-center">
                                <p className="text-slate-400 mb-2">Swagger Interface</p>
                                <a
                                    href={`${appInfo.backend}/swagger/`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-xl text-indigo-400 hover:text-indigo-300 underline decoration-indigo-500/50 underline-offset-4 font-medium transition-colors"
                                >
                                    {appInfo.backend}/swagger/
                                </a>
                            </div>
                            <div className="text-center w-full border-t border-white/5 pt-6">
                                <p style={{ fontSize: '0.8rem' }} className="text-slate-500">
                                    Build {appInfo.version} | <a href="https://github.com/iulianpascalau/mx-epoch-proxy-go" className="hover:text-slate-400 underline decoration-slate-600 underline-offset-2" target="_blank" rel="noopener noreferrer">Solution</a>
                                </p>
                            </div>

                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}
