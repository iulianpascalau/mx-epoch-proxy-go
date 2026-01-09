import { useEffect, useState } from 'react';
import { getAccessKey, clearAuth, getUserInfo, parseJwt, type User as AuthUser } from './auth';
import { useNavigate } from 'react-router-dom';
import { LogOut, Key, Users, Copy, Trash2, Shield, Loader, Plus, User, Pencil, RotateCcw, ChevronLeft, ChevronRight, ArrowUpDown, ArrowUp, ArrowDown, Check, X as XIcon, BookOpen, Lock } from 'lucide-react';
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
    AccountType: string;
    IsActive: boolean;
}

const copyToClipboard = async (text: string) => {
    try {
        if (navigator.clipboard) {
            await navigator.clipboard.writeText(text);
        } else {
            // Fallback for older browsers or non-secure contexts
            const textArea = document.createElement("textarea");
            textArea.value = text;
            textArea.style.position = "fixed";
            textArea.style.left = "-9999px";
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
        }
    } catch (err) {
        console.error('Failed to copy: ', err);
        alert('Failed to copy to clipboard');
    }
};

export const Dashboard = () => {
    const navigate = useNavigate();
    const [user, setUser] = useState<AuthUser | null>(null);
    const [keys, setKeys] = useState<Record<string, AccessKeyDetails>>({});
    const [users, setUsers] = useState<Record<string, UserDetails>>({});
    const [performanceMetrics, setPerformanceMetrics] = useState<Record<string, number>>({});
    const [performanceLabels, setPerformanceLabels] = useState<string[]>([]);
    const [appInfo, setAppInfo] = useState<{ version: string, backend: string } | null>(null);
    const [loading, setLoading] = useState(true);

    // Key Modal State
    const [showKeyModal, setShowKeyModal] = useState(false);
    const [newKeyVal, setNewKeyVal] = useState('');
    const [managingKeysForUser, setManagingKeysForUser] = useState<UserDetails | null>(null);

    // User Modal State
    const [showUserModal, setShowUserModal] = useState(false);
    const [isEditingUser, setIsEditingUser] = useState(false);
    const [newUserState, setNewUserState] = useState({
        username: '',
        password: '',
        maxRequests: 0,
        isAdmin: false,
        accountType: 'free'
    });

    // Pagination & Sorting State
    const ITEMS_PER_PAGE = 10;
    const [keysPage, setKeysPage] = useState(1);
    const [usersPage, setUsersPage] = useState(1);
    const [sortConfig, setSortConfig] = useState<{
        type: 'keys' | 'users' | null;
        key: string | null;
        direction: 'asc' | 'desc';
    }>({ type: null, key: null, direction: 'asc' });

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

        // Fetch App Info
        fetch('/api/app-info')
            .then(res => res.json())
            .then(data => setAppInfo(data))
            .catch(err => console.error('Failed to fetch app info:', err));

        return () => {
            if (timer) clearTimeout(timer);
        };
    }, [navigate]);

    const fetchData = async (isAdmin: boolean) => {
        setLoading(true);
        const token = getAccessKey();
        const userInfo = getUserInfo();
        try {
            const headers = { Authorization: `Bearer ${token}` };

            // Fetch Keys (Everyone)
            // Note: Endpoint returns map[string]AccessKeyDetails
            const keysRes = await axios.get('/api/admin-access-keys', { headers });
            setKeys(keysRes.data || {});

            // Fetch Users
            if (isAdmin) {
                const usersRes = await axios.get('/api/admin-users', { headers });
                setUsers(usersRes.data || {});
            } else if (userInfo.username) {
                const usersRes = await axios.get(`/api/admin-users?username=${userInfo.username}`, { headers });
                setUsers(usersRes.data || {});
            }

            // Fetch Performance Metrics (Admin Only)
            if (isAdmin) {
                try {
                    const perfRes = await axios.get('/api/performance', { headers });
                    if (perfRes.data) {
                        setPerformanceMetrics(perfRes.data.metrics || {});
                        setPerformanceLabels(perfRes.data.labels || []);
                    }
                } catch (e) {
                    console.error("Failed to fetch performance metrics", e);
                }
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

    const refreshPerformance = async () => {
        const token = getAccessKey();
        const userInfo = getUserInfo();
        const isAdmin = userInfo?.is_admin;
        if (!isAdmin) return;

        try {
            const headers = { Authorization: `Bearer ${token}` };
            const perfRes = await axios.get('/api/performance', { headers });
            if (perfRes.data) {
                setPerformanceMetrics(perfRes.data.metrics || {});
                setPerformanceLabels(perfRes.data.labels || []);
            }
        } catch (e) {
            console.error("Failed to refresh performance metrics", e);
        }
    };

    const handleLogout = () => {
        clearAuth();
        navigate('/login');
    };

    const handleCreateKey = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const payload: any = { key: newKeyVal };
            if (managingKeysForUser) {
                payload.username = managingKeysForUser.Username;
            }
            await axios.post('/api/admin-access-keys', payload, {
                headers: { Authorization: `Bearer ${getAccessKey()}` }
            });
            if (!managingKeysForUser) {
                setShowKeyModal(false);
            }
            setNewKeyVal('');
            fetchData(user?.is_admin || false);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : 'Failed to create key';
            alert(msg);
        }
    };

    const handleDeleteKey = async (key: string, username?: string) => {
        if (!confirm('Revoke this key?')) return;
        try {
            let url = `/api/admin-access-keys?key=${key}`;
            if (username) {
                url += `&username=${username}`;
            }
            await axios.delete(url, {
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
                    is_admin: newUserState.isAdmin,
                    account_type: newUserState.accountType
                }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });
            } else {
                await axios.post('/api/admin-users', {
                    username: newUserState.username,
                    password: newUserState.password,
                    max_requests: newUserState.maxRequests,
                    is_admin: newUserState.isAdmin,
                    account_type: newUserState.accountType
                }, { headers: { Authorization: `Bearer ${getAccessKey()}` } });
            }

            setShowUserModal(false);
            setShowUserModal(false);
            setNewUserState({ username: '', password: '', maxRequests: 0, isAdmin: false, accountType: 'free' });
            setIsEditingUser(false);
            fetchData(true);
        } catch (e: any) {
            const msg = e.response?.data ? String(e.response.data).trim() : (isEditingUser ? 'Failed to update user' : 'Failed to create user');
            alert(msg);
        }
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
            isAdmin: u.IsAdmin,
            accountType: u.AccountType || 'free'
        });
        setIsEditingUser(true);
        setShowUserModal(true);
    };

    const openCreateUserModal = () => {
        setNewUserState({ username: '', password: '', maxRequests: 0, isAdmin: false, accountType: 'free' });
        setIsEditingUser(false);
        setShowUserModal(true);
    };


    // Security Settings State
    const [passState, setPassState] = useState({ oldPass: '', newPass: '', confirmPass: '' });
    const [emailState, setEmailState] = useState({ oldPass: '', newEmail: '', confirmEmail: '' });

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
    const handleSort = (type: 'keys' | 'users', key: string) => {
        let direction: 'asc' | 'desc' = 'asc';
        if (sortConfig.type === type && sortConfig.key === key && sortConfig.direction === 'asc') {
            direction = 'desc';
        }
        setSortConfig({ type, key, direction });
    };

    const getSortIcon = (type: 'keys' | 'users', key: string) => {
        if (sortConfig.type !== type || sortConfig.key !== key) {
            return <ArrowUpDown size={14} className="text-slate-600" />;
        }
        return sortConfig.direction === 'asc' ? <ArrowUp size={14} className="text-indigo-400" /> : <ArrowDown size={14} className="text-indigo-400" />;
    };

    const sortedKeys = Object.entries(keys)
        .map(([k, details]) => ({ ...details, ActualKey: k })) // Ensure we have the key string
        .sort((a, b) => {
            if (sortConfig.type !== 'keys' || !sortConfig.key) return 0;
            const valA = String((a as any)[sortConfig.key]).toLowerCase();
            const valB = String((b as any)[sortConfig.key]).toLowerCase();
            if (valA < valB) return sortConfig.direction === 'asc' ? -1 : 1;
            if (valA > valB) return sortConfig.direction === 'asc' ? 1 : -1;
            return 0;
        });

    const paginatedKeys = sortedKeys.slice((keysPage - 1) * ITEMS_PER_PAGE, keysPage * ITEMS_PER_PAGE);
    const totalKeysPages = Math.ceil(sortedKeys.length / ITEMS_PER_PAGE);

    const sortedUsers = Object.values(users).sort((a, b) => {
        if (sortConfig.type !== 'users' || !sortConfig.key) return 0;
        const valA = String((a as any)[sortConfig.key]).toLowerCase();
        const valB = String((b as any)[sortConfig.key]).toLowerCase();
        if (valA < valB) return sortConfig.direction === 'asc' ? -1 : 1;
        if (valA > valB) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
    });

    const paginatedUsers = sortedUsers.slice((usersPage - 1) * ITEMS_PER_PAGE, usersPage * ITEMS_PER_PAGE);
    const totalUsersPages = Math.ceil(sortedUsers.length / ITEMS_PER_PAGE);

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

                    {/* Account Status Panel (Standard Users) */}
                    {!user.is_admin && users[user.username.toLowerCase()] && (
                        <div className="glass-panel p-6 col-span-1 lg:col-span-2">
                            <div className="flex justify-between items-center mb-6">
                                <h2 className="text-xl font-semibold flex items-center gap-2">
                                    <Shield className="text-emerald-400" /> Account Status
                                </h2>
                                <button
                                    onClick={() => fetchData(user.is_admin)}
                                    className="bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                >
                                    <RotateCcw size={16} /> Refresh Data
                                </button>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                                <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                                    <div className="text-slate-400 text-sm mb-2">Account Type</div>
                                    <span className={`px-3 py-1 rounded text-lg font-bold ${users[user.username.toLowerCase()].AccountType === 'premium' ? 'bg-amber-500/20 text-amber-300' : 'bg-slate-500/20 text-slate-300'}`}>
                                        {(users[user.username.toLowerCase()].AccountType || 'free').toUpperCase()}
                                    </span>
                                </div>
                                <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                                    <div className="text-slate-400 text-sm mb-1">Max Requests</div>
                                    <div className="text-2xl font-bold text-slate-200">
                                        {users[user.username.toLowerCase()].MaxRequests === 0 ? 'Unlimited' : users[user.username.toLowerCase()].MaxRequests}
                                    </div>
                                </div>
                                <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                                    <div className="text-slate-400 text-sm mb-1">Total Usage (Global)</div>
                                    <div className="text-2xl font-bold text-indigo-400">
                                        {users[user.username.toLowerCase()].GlobalCounter}
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Keys Panel */}
                    <div className="glass-panel p-6 flex flex-col h-full col-span-1 lg:col-span-2">
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4 md:gap-0">
                            <h2 className="text-xl font-semibold flex items-center gap-2">
                                <Key className="text-indigo-400" /> Access Keys
                            </h2>
                            <div className="flex gap-3 w-full md:w-auto">
                                <button
                                    onClick={() => window.location.reload()}
                                    className="flex-1 md:flex-none justify-center bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                >
                                    <RotateCcw size={16} /> Refresh Data
                                </button>
                                <button
                                    onClick={() => setShowKeyModal(true)}
                                    className="flex-1 md:flex-none justify-center bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                >
                                    <Plus size={16} /> Generate Key
                                </button>
                            </div>
                        </div>

                        <div className="overflow-x-auto min-h-[300px]">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="border-b border-white/10 text-slate-400 text-sm uppercase">
                                        <th
                                            className="py-3 px-4 cursor-pointer hover:text-white transition-colors group"
                                            onClick={() => handleSort('keys', 'ActualKey')}
                                        >
                                            <div className="flex items-center gap-1">
                                                Key Value {getSortIcon('keys', 'ActualKey')}
                                            </div>
                                        </th>
                                        {user.is_admin && (
                                            <th
                                                className="py-3 px-4 cursor-pointer hover:text-white transition-colors group"
                                                onClick={() => handleSort('keys', 'Username')}
                                            >
                                                <div className="flex items-center gap-1">
                                                    Owner {getSortIcon('keys', 'Username')}
                                                </div>
                                            </th>
                                        )}
                                        <th className="py-3 px-4">Requests</th>
                                        <th className="py-3 px-4 text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {paginatedKeys.map((details) => (
                                        <tr key={details.ActualKey} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                                            <td className="py-3 px-4 font-mono text-sm text-indigo-200">
                                                <div className="flex items-center gap-2">
                                                    {details.ActualKey}
                                                    <Copy
                                                        size={14}
                                                        className="cursor-pointer text-slate-500 hover:text-white"
                                                        onClick={() => copyToClipboard(details.ActualKey)}
                                                    />
                                                </div>
                                            </td>
                                            {user.is_admin && (
                                                <td className="py-3 px-4 text-slate-300">
                                                    {details.Username}
                                                </td>
                                            )}
                                            <td className="py-3 px-4 text-slate-300">
                                                {details.KeyCounter}
                                            </td>
                                            <td className="py-3 px-4 text-right">
                                                <button
                                                    onClick={() => handleDeleteKey(details.ActualKey, details.Username)}
                                                    className="text-red-400 hover:text-red-300 p-1 rounded hover:bg-red-400/10 transition-colors"
                                                >
                                                    <Trash2 size={16} />
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    {paginatedKeys.length === 0 && (
                                        <tr>
                                            <td colSpan={user.is_admin ? 4 : 3} className="py-8 text-center text-slate-500">
                                                No access keys found.
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>

                        {/* Pagination for Keys */}
                        {totalKeysPages > 1 && (
                            <div className="flex justify-between items-center mt-4 pt-4 border-t border-white/5">
                                <button
                                    onClick={() => setKeysPage(p => Math.max(1, p - 1))}
                                    disabled={keysPage === 1}
                                    className="p-1 rounded hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-slate-400 hover:text-white"
                                >
                                    <ChevronLeft size={20} />
                                </button>
                                <span className="text-sm text-slate-400">
                                    Page {keysPage} of {totalKeysPages}
                                </span>
                                <button
                                    onClick={() => setKeysPage(p => Math.min(totalKeysPages, p + 1))}
                                    disabled={keysPage === totalKeysPages}
                                    className="p-1 rounded hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-slate-400 hover:text-white"
                                >
                                    <ChevronRight size={20} />
                                </button>
                            </div>
                        )}
                    </div>

                    {/* Users Panel (Admin Only) */}
                    {user.is_admin && (
                        <div className="glass-panel p-6 col-span-1 lg:col-span-2">
                            <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4 md:gap-0">
                                <h2 className="text-xl font-semibold flex items-center gap-2">
                                    <Users className="text-emerald-400" /> User Management
                                </h2>
                                <div className="flex gap-3 w-full md:w-auto">
                                    <button
                                        onClick={() => window.location.reload()}
                                        className="flex-1 md:flex-none justify-center bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                    >
                                        <RotateCcw size={16} /> Refresh Data
                                    </button>
                                    <button
                                        onClick={openCreateUserModal}
                                        className="flex-1 md:flex-none justify-center bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all"
                                    >
                                        <Plus size={16} /> Add User
                                    </button>
                                </div>
                            </div>

                            <div className="overflow-x-auto min-h-[300px]">
                                <table className="w-full text-left border-collapse">
                                    <thead>
                                        <tr className="border-b border-white/10 text-slate-400 text-sm uppercase">
                                            <th
                                                className="py-3 px-4 cursor-pointer hover:text-white transition-colors group"
                                                onClick={() => handleSort('users', 'Username')}
                                            >
                                                <div className="flex items-center gap-1">
                                                    Username {getSortIcon('users', 'Username')}
                                                </div>
                                            </th>
                                            <th className="py-3 px-4">Verified</th>
                                            <th className="py-3 px-4">Role</th>
                                            <th className="py-3 px-4">Account Type</th>
                                            <th className="py-3 px-4">Limits (Req)</th>
                                            <th className="py-3 px-4"></th>
                                            <th className="py-3 px-4">Current Usage</th>
                                            <th className="py-3 px-4 text-right">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {paginatedUsers.map((u) => (
                                            <tr key={u.Username} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                                                <td className="py-3 px-4 text-slate-200 font-medium">{u.Username}</td>
                                                <td className="py-3 px-4">
                                                    {u.IsActive ? (
                                                        <Check className="text-emerald-400" size={18} />
                                                    ) : (
                                                        <XIcon className="text-red-400" size={18} />
                                                    )}
                                                </td>
                                                <td className="py-3 px-4">
                                                    <span className={`px-2 py-1 rounded text-xs font-semibold ${u.IsAdmin ? 'bg-indigo-500/20 text-indigo-300' : 'bg-emerald-500/20 text-emerald-300'}`}>
                                                        {u.IsAdmin ? 'ADMIN' : 'USER'}
                                                    </span>
                                                </td>
                                                <td className="py-3 px-4">
                                                    <span className={`px-2 py-1 rounded text-xs font-semibold ${u.AccountType === 'premium' ? 'bg-amber-500/20 text-amber-300' : 'bg-slate-500/20 text-slate-300'}`}>
                                                        {(u.AccountType || 'free').toUpperCase()}
                                                    </span>
                                                </td>
                                                <td className="py-3 px-4 text-slate-300">
                                                    {u.MaxRequests === 0 ? 'Unlimited' : u.MaxRequests}
                                                </td>
                                                <td className="py-3 px-4">
                                                    <button
                                                        onClick={() => setManagingKeysForUser(u)}
                                                        className="flex items-center gap-1 text-xs bg-slate-700 hover:bg-slate-600 px-2 py-1 rounded transition-colors text-slate-300"
                                                    >
                                                        <Key size={12} />
                                                        Manage Keys
                                                    </button>
                                                </td>
                                                <td className="py-3 px-4 text-slate-300">
                                                    <div className="flex items-center gap-2">
                                                        {u.GlobalCounter}
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

                            {/* Pagination for Users */}
                            {totalUsersPages > 1 && (
                                <div className="flex justify-between items-center mt-4 pt-4 border-t border-white/5">
                                    <button
                                        onClick={() => setUsersPage(p => Math.max(1, p - 1))}
                                        disabled={usersPage === 1}
                                        className="p-1 rounded hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-slate-400 hover:text-white"
                                    >
                                        <ChevronLeft size={20} />
                                    </button>
                                    <span className="text-sm text-slate-400">
                                        Page {usersPage} of {totalUsersPages}
                                    </span>
                                    <button
                                        onClick={() => setUsersPage(p => Math.min(totalUsersPages, p + 1))}
                                        disabled={usersPage === totalUsersPages}
                                        className="p-1 rounded hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-slate-400 hover:text-white"
                                    >
                                        <ChevronRight size={20} />
                                    </button>
                                </div>
                            )}
                        </div>
                    )}

                </div>
            )}

            {/* Security Settings Panel */}
            <div className="glass-panel p-6 col-span-1 lg:col-span-2">
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
                                    type="password" required minLength={8}
                                    className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                                    value={passState.newPass}
                                    onChange={e => setPassState({ ...passState, newPass: e.target.value })}
                                />
                            </div>
                            <div>
                                <label className="block text-sm text-slate-400 mb-1">Confirm New Password</label>
                                <input
                                    type="password" required minLength={8}
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
                        <h3 className="text-lg font-medium mb-4 text-slate-200">Change Registered Email</h3>
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
            {user.is_admin && (
                <div className="glass-panel p-6 col-span-1 lg:col-span-2">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-xl font-semibold flex items-center gap-2">
                            <ArrowUp className="text-indigo-400" /> Response Time Distribution
                        </h2>
                        <button
                            onClick={refreshPerformance}
                            className="bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm font-normal transition-all"
                        >
                            <RotateCcw size={16} /> Refresh Graph
                        </button>
                    </div>
                    <div className="h-80 flex items-end gap-2 pt-6 pb-24 px-4 bg-black/20 rounded-lg">
                        {(() => {
                            const labelsToUse = performanceLabels.length > 0 ? performanceLabels : Object.keys(performanceMetrics).sort();
                            const maxPerf = Math.max(...Object.values(performanceMetrics).concat([1]));
                            return labelsToUse.map(label => {
                                const val = performanceMetrics[label] || 0;
                                const height = (val / maxPerf) * 100;
                                return (
                                    <div key={label} className="flex-1 flex flex-col items-center gap-1 group h-full justify-end relative">
                                        <div className="text-xs text-slate-400 opacity-0 group-hover:opacity-100 transition-opacity mb-1 font-mono absolute -top-5">{val}</div>
                                        <div
                                            className="w-full bg-indigo-500/50 hover:bg-indigo-400 rounded-t transition-all relative min-h-[1px]"
                                            style={{ height: `${height}%` }}
                                        >
                                        </div>
                                        <div className="absolute -bottom-2 w-0 h-0 flex justify-center items-center">
                                            <div className="text-[10px] text-slate-500 -rotate-90 origin-right translate-y-[50%] translate-x-[-50%] whitespace-nowrap w-24 text-right pr-2">
                                                {label}
                                            </div>
                                        </div>
                                    </div>
                                )
                            });
                        })()}
                    </div>
                </div>
            )}

            {/* API Info Panel */}
            {appInfo && (
                <div className="glass-panel p-6 col-span-1 lg:col-span-2">
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
    )
}

{/* Key Modal */ }
{
    showKeyModal && (
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
    )
}

{/* User Modal */ }
{
    showUserModal && (
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
                        <div>
                            <label className="block text-sm text-slate-400 mb-1">Account Type</label>
                            <select
                                className="w-full bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                value={newUserState.accountType}
                                onChange={e => setNewUserState({ ...newUserState, accountType: e.target.value })}
                            >
                                <option value="free">Free</option>
                                <option value="premium">Premium</option>
                            </select>
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
    )
}

{/* Manage User Keys Modal */ }
{
    managingKeysForUser && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
            <div className="glass-panel w-full max-w-2xl p-6 animate-in fade-in zoom-in duration-200">
                <div className="flex justify-between items-center mb-6">
                    <h3 className="text-xl font-bold flex items-center gap-2">
                        <Key className="text-indigo-400" />
                        Keys for {managingKeysForUser.Username}
                    </h3>
                    <button
                        onClick={() => setManagingKeysForUser(null)}
                        className="text-slate-400 hover:text-white"
                    >
                        ✕
                    </button>
                </div>

                {/* Add Key Form */}
                <form onSubmit={handleCreateKey} className="mb-6 flex gap-2">
                    <input
                        type="text"
                        className="flex-1 bg-slate-800 border border-slate-700 rounded p-2 text-slate-200 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                        placeholder="New Key Value (Optional)"
                        value={newKeyVal}
                        onChange={e => setNewKeyVal(e.target.value)}
                    />
                    <button type="submit" className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded text-white whitespace-nowrap">
                        Add Key
                    </button>
                </form>

                {/* Keys List */}
                <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="border-b border-white/10 text-slate-400 text-sm uppercase">
                                <th className="py-2 px-4">Key Value</th>
                                <th className="py-2 px-4">Requests</th>
                                <th className="py-2 px-4 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {Object.entries(keys)
                                .filter(([_, details]) => details.Username === managingKeysForUser.Username)
                                .map(([k, details]) => (
                                    <tr key={k} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                                        <td className="py-2 px-4 font-mono text-sm text-indigo-200">
                                            <div className="flex items-center gap-2">
                                                {k}
                                                <Copy
                                                    size={12}
                                                    className="cursor-pointer text-slate-500 hover:text-white"
                                                    onClick={() => copyToClipboard(k)}
                                                />
                                            </div>
                                        </td>
                                        <td className="py-2 px-4 text-slate-300">
                                            {details.KeyCounter}
                                        </td>
                                        <td className="py-2 px-4 text-right">
                                            <button
                                                onClick={() => handleDeleteKey(k, managingKeysForUser.Username)}
                                                className="text-red-400 hover:text-red-300 p-1 rounded hover:bg-red-400/10 transition-colors"
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            {Object.values(keys).filter(k => k.Username === managingKeysForUser.Username).length === 0 && (
                                <tr>
                                    <td colSpan={3} className="py-4 text-center text-slate-500">
                                        No specific keys for this user.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    )
}
        </div >
    );
};
