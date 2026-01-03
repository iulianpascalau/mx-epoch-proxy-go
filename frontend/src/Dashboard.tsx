import { useEffect, useState } from 'react';
import { getAccessKey, clearAuth, getUserInfo, parseJwt, type User as AuthUser } from './auth';
import { useNavigate } from 'react-router-dom';
import { LogOut, Key, Users, Copy, Trash2, Shield, Loader, Plus, User, Pencil, RotateCcw } from 'lucide-react';
import axios from 'axios';


interface AccessKeyDetails {
    Key: string;
    MaxRequests: number;
    GlobalCounter: number;
    KeyCounter: number;
    Username: string;
    IsAdmin: boolean;
}

interface UserDetails {
    Username: string;
    MaxRequests: number;
    GlobalCounter: number;
    IsAdmin: boolean;
}

export const Dashboard = () => {
    const navigate = useNavigate();
    const [user, setUser] = useState<AuthUser | null>(null);
    const [keys, setKeys] = useState<Record<string, AccessKeyDetails>>({});
    const [users, setUsers] = useState<Record<string, UserDetails>>({});
    const [loading, setLoading] = useState(true);

    // Key Modal State
    const [showKeyModal, setShowKeyModal] = useState(false);
    const [newKeyVal, setNewKeyVal] = useState('');

    // User Modal State
    const [showUserModal, setShowUserModal] = useState(false);
    const [isEditingUser, setIsEditingUser] = useState(false);
    const [newUserState, setNewUserState] = useState({
        username: '',
        password: '',
        maxRequests: 0,
        isAdmin: false
    });

    useEffect(() => {
        const token = getAccessKey();
        const info = getUserInfo();
        if (!token || !info.username) {
            navigate('/login');
            return;
        }

        const decoded = parseJwt(token);
        let timer: ReturnType<typeof setTimeout>;

        if (decoded && decoded.exp) {
            const expirationTime = decoded.exp * 1000;
            const now = Date.now();

            if (now >= expirationTime) {
                clearAuth();
                navigate('/login');
                return;
            }

            timer = setTimeout(() => {
                clearAuth();
                navigate('/login');
            }, expirationTime - now);
        }

        setUser(info);
        fetchData(info.is_admin);

        return () => {
            if (timer) clearTimeout(timer);
        };
    }, [navigate]);

    const fetchData = async (isAdmin: boolean) => {
        setLoading(true);
        const token = getAccessKey();
        try {
            const headers = { Authorization: `Bearer ${token}` };

            // Fetch Keys (Everyone)
            // Note: Endpoint returns map[string]AccessKeyDetails
            const keysRes = await axios.get('/api/admin-access-keys', { headers });
            setKeys(keysRes.data || {});

            // Fetch Users (Admin Only)
            if (isAdmin) {
                const usersRes = await axios.get('/api/admin-users', { headers });
                setUsers(usersRes.data || {});
            }
        } catch (e: any) {
            if (e.response?.status === 401) {
                handleLogout();
            }
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const handleLogout = () => {
        clearAuth();
        navigate('/login');
    };

    const handleCreateKey = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await axios.post('/api/admin-access-keys', { key: newKeyVal }, {
                headers: { Authorization: `Bearer ${getAccessKey()}` }
            });
            setShowKeyModal(false);
            setNewKeyVal('');
            fetchData(user?.is_admin || false);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to create key';
            alert(msg);
        }
    };

    const handleDeleteKey = async (key: string) => {
        if (!confirm('Revoke this key?')) return;
        try {
            await axios.delete(`/api/admin-access-keys?key=${key}`, {
                headers: { Authorization: `Bearer ${getAccessKey()}` }
            });
            fetchData(user?.is_admin || false);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to delete key';
            alert(msg);
        }
    };

    const handleCreateUser = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (isEditingUser) {
                await axios.put('/api/admin-users', {
                    username: newUserState.username,
                    password: newUserState.password, // Optional in backend
                    max_requests: newUserState.maxRequests,
                    is_admin: newUserState.isAdmin
                }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });
            } else {
                await axios.post('/api/admin-users', {
                    username: newUserState.username,
                    password: newUserState.password,
                    max_requests: newUserState.maxRequests,
                    is_admin: newUserState.isAdmin
                }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });
            }

            setShowUserModal(false);
            setNewUserState({ username: '', password: '', maxRequests: 0, isAdmin: false });
            setIsEditingUser(false);
            fetchData(true);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : (isEditingUser ? 'Failed to update user' : 'Failed to create user');
            alert(msg);
        }
    };

    const handleRefreshUserCounters = async () => {
        await fetchData(user?.is_admin || false);
    };

    const handleDeleteUser = async (username: string) => {
        if (!confirm(`Are you sure you want to delete user "${username}"? This will also revoke all their keys.`)) return;
        try {
            await axios.delete(`/api/admin-users?username=${username}`, {
                headers: { Authorization: `Bearer ${getAccessKey()}` }
            });
            fetchData(user?.is_admin || false);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to delete user';
            alert(msg);
        }
    };

    const handleEditUser = (u: UserDetails) => {
        setNewUserState({
            username: u.Username,
            password: '', // Password not shown
            maxRequests: u.MaxRequests,
            isAdmin: u.IsAdmin
        });
        setIsEditingUser(true);
        setShowUserModal(true);
    };

    const openCreateUserModal = () => {
        setNewUserState({ username: '', password: '', maxRequests: 0, isAdmin: false });
        setIsEditingUser(false);
        setShowUserModal(true);
    };

    if (!user) return null;

    return (
        <div className="min-h-screen p-6 md:p-12 max-w-7xl mx-auto">
            {/* Header */}
            <div className="glass-panel p-6 mb-8 flex justify-between items-center">
                <div className="flex items-center gap-4">
                    <div className="h-10 w-10 rounded-full bg-indigo-500/20 flex items-center justify-center text-indigo-400">
                        {user.is_admin ? <Shield size={20} /> : <User size={20} />}
                    </div>
                    <div>
                        <h1 className="text-xl font-bold">{user.username}</h1>
                        <span className="text-sm text-slate-400">{user.is_admin ? 'Administrator' : 'Standard User'}</span>
                    </div>
                </div>
                <button onClick={handleLogout} className="flex items-center gap-2 px-4 py-2 rounded-lg hover:bg-white/5 transition-colors text-slate-300">
                    <LogOut size={18} />
                    <span>Sign Out</span>
                </button>
            </div>

            {loading ? (
                <div className="flex justify-center mt-20">
                    <Loader className="animate-spin text-indigo-500" size={40} />
                </div>
            ) : (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">

                    {/* Keys Panel */}
                    <div className="glass-panel p-6 flex flex-col h-full col-span-1 lg:col-span-2">
                        <div className="flex justify-between items-center mb-6">
                            <h2 className="text-xl font-semibold flex items-center gap-2">
                                <Key className="text-indigo-400" /> Access Keys
                            </h2>
                            <button
                                onClick={() => setShowKeyModal(true)}
                                className="bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                            >
                                <Plus size={16} /> Generate Key
                            </button>
                        </div>

                        <div className="overflow-x-auto">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="border-b border-white/10 text-slate-400 text-sm uppercase">
                                        <th className="py-3 px-4">Key Value</th>
                                        <th className="py-3 px-4">Requests</th>
                                        <th className="py-3 px-4 text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {Object.entries(keys).map(([k, details]) => (
                                        <tr key={k} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                                            <td className="py-3 px-4 font-mono text-sm text-indigo-200">
                                                <div className="flex items-center gap-2">
                                                    {k}
                                                    <Copy
                                                        size={14}
                                                        className="cursor-pointer text-slate-500 hover:text-white"
                                                        onClick={() => navigator.clipboard.writeText(k)}
                                                    />
                                                </div>
                                            </td>
                                            <td className="py-3 px-4 text-slate-300">
                                                {details.KeyCounter}
                                            </td>
                                            <td className="py-3 px-4 text-right">
                                                <button
                                                    onClick={() => handleDeleteKey(k)}
                                                    className="text-red-400 hover:text-red-300 p-1 rounded hover:bg-red-400/10 transition-colors"
                                                >
                                                    <Trash2 size={16} />
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    {Object.keys(keys).length === 0 && (
                                        <tr>
                                            <td colSpan={3} className="py-8 text-center text-slate-500">
                                                No access keys found. Generate one to get started.
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>

                    {/* Users Panel (Admin Only) */}
                    {user.is_admin && (
                        <div className="glass-panel p-6 col-span-1 lg:col-span-2">
                            <div className="flex justify-between items-center mb-6">
                                <h2 className="text-xl font-semibold flex items-center gap-2">
                                    <Users className="text-emerald-400" /> User Management
                                </h2>
                                <button
                                    onClick={openCreateUserModal}
                                    className="bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                >
                                    <Plus size={16} /> Add User
                                </button>
                            </div>

                            <div className="overflow-x-auto">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="border-b border-white/10 text-slate-400 text-sm uppercase">
                                            <th className="py-3 px-4">Username</th>
                                            <th className="py-3 px-4">Role</th>
                                            <th className="py-3 px-4">Limits (Req)</th>
                                            <th className="py-3 px-4">Current Usage</th>
                                            <th className="py-3 px-4 text-right">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {Object.values(users).map((u) => (
                                            <tr key={u.Username} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                                                <td className="py-3 px-4 text-slate-200 font-medium">{u.Username}</td>
                                                <td className="py-3 px-4">
                                                    <span className={`px-2 py-1 rounded text-xs font-semibold ${u.IsAdmin ? 'bg-indigo-500/20 text-indigo-300' : 'bg-emerald-500/20 text-emerald-300'}`}>
                                                        {u.IsAdmin ? 'ADMIN' : 'USER'}
                                                    </span>
                                                </td>
                                                <td className="py-3 px-4 text-slate-300">
                                                    {u.MaxRequests === 0 ? 'Unlimited' : u.MaxRequests}
                                                </td>
                                                <td className="py-3 px-4 text-slate-300">
                                                    <div className="flex items-center gap-2">
                                                        {u.GlobalCounter}
                                                        <button
                                                            title="Refresh Usage Data"
                                                            onClick={handleRefreshUserCounters}
                                                            className="text-slate-500 hover:text-indigo-400 p-0.5 rounded transition-colors"
                                                        >
                                                            <RotateCcw size={14} />
                                                        </button>
                                                    </div>
                                                </td>
                                                <td className="py-3 px-4 text-right flex justify-end gap-2">
                                                    <button
                                                        onClick={() => handleEditUser(u)}
                                                        className="text-indigo-400 hover:text-indigo-300 p-1 rounded hover:bg-indigo-400/10 transition-colors"
                                                    >
                                                        <Pencil size={16} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDeleteUser(u.Username)}
                                                        className="text-red-400 hover:text-red-300 p-1 rounded hover:bg-red-400/10 transition-colors"
                                                    >
                                                        <Trash2 size={16} />
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {/* Key Modal */}
            {showKeyModal && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
                    <div className="glass-panel w-full max-w-md p-6 animate-in fade-in zoom-in duration-200">
                        <h3 className="text-xl font-bold mb-4">Generate New Key</h3>
                        <form onSubmit={handleCreateKey}>
                            <div className="mb-4">
                                <label className="block text-sm text-slate-400 mb-1">Key Value (Optional)</label>
                                <input
                                    type="text"
                                    className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                    placeholder="Leave empty for random UUID"
                                    value={newKeyVal}
                                    onChange={e => setNewKeyVal(e.target.value)}
                                />
                            </div>
                            <div className="flex justify-end gap-3 mt-6">
                                <button type="button" onClick={() => { setShowKeyModal(false); setNewKeyVal(''); }} className="px-4 py-2 hover:bg-white/5 rounded text-slate-300">Cancel</button>
                                <button type="submit" className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded text-white">Create</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* User Modal */}
            {showUserModal && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
                    <div className="glass-panel w-full max-w-md p-6 animate-in fade-in zoom-in duration-200">
                        <h3 className="text-xl font-bold mb-4">{isEditingUser ? 'Edit User' : 'Add New User'}</h3>
                        <form onSubmit={handleCreateUser}>
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Username</label>
                                    <input
                                        type="text" required
                                        readOnly={isEditingUser}
                                        className={`w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-emerald-500 focus:outline-none ${isEditingUser ? 'opacity-50 cursor-not-allowed' : ''}`}
                                        value={newUserState.username}
                                        onChange={e => setNewUserState({ ...newUserState, username: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Password {isEditingUser && '(Leave empty to keep current)'}</label>
                                    <input
                                        type="password"
                                        required={!isEditingUser}
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                        value={newUserState.password}
                                        onChange={e => setNewUserState({ ...newUserState, password: e.target.value })}
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm text-slate-400 mb-1">Max Requests (0 = Unlimited)</label>
                                    <input
                                        type="number" min="0"
                                        className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                        value={newUserState.maxRequests}
                                        onChange={e => setNewUserState({ ...newUserState, maxRequests: parseInt(e.target.value) })}
                                    />
                                </div>
                                <div className="flex items-center gap-2">
                                    <input
                                        type="checkbox" id="isAdminCheck"
                                        className="w-4 h-4 rounded border-slate-700 bg-slate-800 text-emerald-600 focus:ring-emerald-500"
                                        checked={newUserState.isAdmin}
                                        onChange={e => setNewUserState({ ...newUserState, isAdmin: e.target.checked })}
                                    />
                                    <label htmlFor="isAdminCheck" className="text-sm text-slate-300">Grant Administrator Privileges</label>
                                </div>
                            </div>
                            <div className="flex justify-end gap-3 mt-6">
                                <button type="button" onClick={() => setShowUserModal(false)} className="px-4 py-2 hover:bg-white/5 rounded text-slate-300">Cancel</button>
                                <button type="submit" className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 rounded text-white">{isEditingUser ? 'Update User' : 'Save User'}</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};
