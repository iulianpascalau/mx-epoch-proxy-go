import { useEffect, useState } from 'react';
import { getAccessKey, clearAuth, getUserInfo, parseJwt, type User as AuthUser } from './auth';
import { useNavigate } from 'react-router-dom';
import { LogOut, Key, Users, Copy, Trash2, Shield, Loader, Plus, User, Pencil, RotateCcw, ChevronLeft, ChevronRight, ArrowUpDown, ArrowUp, ArrowDown, Check, X as XIcon, UserCog, BookOpen, ExternalLink, Zap, AlertTriangle, CreditCard, RefreshCw, Wallet } from 'lucide-react';
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
    PaymentID?: number; // Added for crypto payment
}

interface CryptoPaymentState {
    isServiceAvailable: boolean;
    isPaused: boolean;
    requestsPerEGLD: number;
    walletURL: string;
    explorerURL: string;
    contractAddress: string;

    paymentId: number | null;
    depositAddress: string | null;
    numberOfRequests: number;

    isLoading: boolean;
    error: string | null;
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

    // Crypto Payment State
    const [cryptoState, setCryptoState] = useState<CryptoPaymentState>({
        isServiceAvailable: true,
        isPaused: false,
        requestsPerEGLD: 10000,
        walletURL: 'https://devnet-wallet.multiversx.com',
        explorerURL: 'https://devnet-explorer.multiversx.com',
        contractAddress: 'erd1qqqqqqqqqqqqqpgqc6u0p4kfkr5ekcrae86m6knx46gr36khrqqqhf96zw',
        paymentId: null,
        depositAddress: null,
        numberOfRequests: 0,
        isLoading: false,
        error: null
    });

    // Mock Crypto API
    const fetchCryptoData = async () => {
        setCryptoState(prev => ({ ...prev, isLoading: true, error: null }));
        // Simulate API delay
        await new Promise(resolve => setTimeout(resolve, 800));

        // Mock Configuration
        const mockConfig = {
            isAvailable: true,
            isPaused: false, // Set to true to test paused state
            requestsPerEGLD: 500000,
            walletURL: "https://devnet-wallet.multiversx.com",
            explorerURL: "https://devnet-explorer.multiversx.com",
            contractAddress: "erd1qqqqqqqqqqqqqpgqc6u0p4kfkr5ekcrae86m6knx46gr36khrqqqhf96zw"
        };

        // Mock Account (dependent on user state, simplified for demo)
        // In reality we would check if user has PaymentID
        const mockAccount = {
            paymentId: user?.is_admin ? 12345 : null, // Admin gets dummy payment ID for demo
            depositAddress: user?.is_admin ? "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruq0gwzgd" : null,
            numberOfRequests: user?.is_admin ? 8500 : 0
        };

        setCryptoState(prev => ({
            ...prev,
            isServiceAvailable: mockConfig.isAvailable,
            isPaused: mockConfig.isPaused,
            requestsPerEGLD: mockConfig.requestsPerEGLD,
            walletURL: mockConfig.walletURL,
            explorerURL: mockConfig.explorerURL,
            contractAddress: mockConfig.contractAddress,
            paymentId: mockAccount.paymentId,
            depositAddress: mockAccount.depositAddress,
            numberOfRequests: mockAccount.numberOfRequests,
            isLoading: false
        }));
    };

    const handleRequestAddress = async () => {
        if (cryptoState.isPaused) return;
        setCryptoState(prev => ({ ...prev, isLoading: true }));

        // Simulate API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        const newPaymentId = Math.floor(Math.random() * 10000) + 1;
        const newAddress = "erd1mocknewaddress" + Math.random().toString(36).substring(7);

        setCryptoState(prev => ({
            ...prev,
            paymentId: newPaymentId,
            depositAddress: newAddress,
            numberOfRequests: 0,
            isLoading: false
        }));
    };

    const handleRefreshBalance = async () => {
        setCryptoState(prev => ({ ...prev, isLoading: true }));
        await new Promise(resolve => setTimeout(resolve, 800));

        // Simulate balance increase if address exists
        if (cryptoState.paymentId) {
            setCryptoState(prev => ({
                ...prev,
                numberOfRequests: prev.numberOfRequests + 100, // Simulate incoming payment
                isLoading: false
            }));
        } else {
            setCryptoState(prev => ({ ...prev, isLoading: false }));
        }
    };

    useEffect(() => {
        if (user) {
            fetchCryptoData();
        }
    }, [user]);

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

        // Fetch App Info (for Swagger/Version)
        const fetchAppInfo = async () => {
            const token = getAccessKey();
            if (!token) return;
            try {
                const configRes = await axios.get('/api/app-info', { headers: { Authorization: `Bearer ${token}` } });
                if (configRes.data) {
                    setAppInfo({
                        version: configRes.data.version || 'v1.0.0',
                        backend: configRes.data.backend || window.location.origin
                    });
                }
            } catch (e) {
                console.error("Failed to fetch app config", e);
            }
        };
        fetchAppInfo();


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
            <div className="glass-panel p-6 flex flex-col md:flex-row justify-between items-center gap-4 md:gap-0">
                <div className="flex items-center gap-4 w-full md:w-auto">
                    <div className="h-10 w-10 rounded-full bg-indigo-500/20 flex items-center justify-center text-indigo-400 shrink-0">
                        {user.is_admin ? <Shield size={20} /> : <User size={20} />}
                    </div>
                    <div>
                        <h1 className="text-xl font-bold truncate max-w-[200px] md:max-w-none">{user.username}</h1>
                        <span className="text-sm text-slate-400">{user.is_admin ? 'Administrator' : 'Standard User'}</span>
                    </div>
                </div>
                <div className="flex gap-3 w-full md:w-auto md:justify-end items-center">
                    <button onClick={() => navigate('/settings')} className="flex items-center gap-2 px-4 py-2 rounded-lg hover:bg-white/5 transition-colors text-slate-300 whitespace-nowrap">
                        <UserCog size={18} />
                        <span>Settings</span>
                    </button>
                    <button onClick={handleLogout} className="flex items-center gap-2 px-4 py-2 rounded-lg hover:bg-white/5 transition-colors text-slate-300 whitespace-nowrap">
                        <LogOut size={18} />
                        <span>Sign Out</span>
                    </button>
                </div>
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
                            <div className="flex justify-between items-center mb-6 flex-wrap gap-4">
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


                    {/* Crypto Payment / Premium Management Section */}
                    {/* Only show for non-admin users or admins who want to see the UI */}
                    {!user.is_admin && users[user.username.toLowerCase()] && (
                        <div className="glass-panel p-6 col-span-1 lg:col-span-2 relative overflow-hidden">
                            {/* Background accent for premium feel */}
                            {users[user.username.toLowerCase()].AccountType === 'premium' && (
                                <div className="absolute top-0 right-0 p-32 bg-amber-500/10 blur-3xl rounded-full -mr-16 -mt-16 pointer-events-none"></div>
                            )}

                            <div className="flex justify-between items-center mb-6 flex-wrap gap-4 relative z-10">
                                <h2 className="text-xl font-semibold flex items-center gap-2">
                                    <Zap className={users[user.username.toLowerCase()].AccountType === 'premium' ? "text-amber-400" : "text-indigo-400"} />
                                    {users[user.username.toLowerCase()].AccountType === 'premium' ? "Premium Account Management" : "Upgrade to Premium"}
                                </h2>
                                {/* Service Status Indicator */}
                                <div className="flex items-center gap-2 text-sm bg-black/20 px-3 py-1 rounded-full">
                                    <div className={`w-2 h-2 rounded-full ${!cryptoState.isServiceAvailable ? 'bg-red-500' : cryptoState.isPaused ? 'bg-amber-500' : 'bg-emerald-500'}`}></div>
                                    <span className="text-slate-400">
                                        {!cryptoState.isServiceAvailable ? 'Crypto Service Unavailable' : cryptoState.isPaused ? 'Payments Paused' : 'Crypto Service Online'}
                                    </span>
                                </div>
                            </div>

                            {cryptoState.error && (
                                <div className="bg-red-500/10 border border-red-500/20 text-red-200 p-4 rounded-lg mb-6 flex items-center gap-3">
                                    <AlertTriangle size={20} />
                                    {cryptoState.error}
                                </div>
                            )}

                            {/* State: Loading */}
                            {cryptoState.isLoading && !cryptoState.depositAddress && (
                                <div className="flex justify-center p-8">
                                    <Loader className="animate-spin text-indigo-500" size={30} />
                                </div>
                            )}

                            {/* State: Free User - No Payment ID */}
                            {!cryptoState.paymentId && users[user.username.toLowerCase()].AccountType !== 'premium' && !cryptoState.isLoading && (
                                <div className="space-y-6 relative z-10">
                                    <div className="bg-indigo-500/10 border border-indigo-500/20 rounded-lg p-6">
                                        <div className="flex items-start gap-4">
                                            <div className="bg-indigo-500/20 p-3 rounded-full text-indigo-400 hidden sm:block">
                                                <CreditCard size={24} />
                                            </div>
                                            <div>
                                                <h3 className="text-lg font-medium text-white mb-2">Unlock Unlimited Requests</h3>
                                                <p className="text-slate-400 mb-4 text-sm leading-relaxed max-w-2xl">
                                                    Upgrade your account to Premium by making a secure crypto payment.
                                                    You are paying directly to the smart contract using eGLD.
                                                    Zero gas fees for deposit transaction relay.
                                                </p>
                                                <div className="flex flex-wrap gap-2 mb-6">
                                                    <span className="bg-white/5 px-2 py-1 rounded text-xs text-slate-300">Rate: {cryptoState.requestsPerEGLD.toLocaleString()} req / 1 eGLD</span>
                                                    <span className="bg-white/5 px-2 py-1 rounded text-xs text-slate-300">Instant Activation</span>
                                                </div>
                                                <button
                                                    onClick={handleRequestAddress}
                                                    disabled={cryptoState.isPaused || !cryptoState.isServiceAvailable}
                                                    className="bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-700 disabled:cursor-not-allowed text-white px-6 py-2 rounded-lg font-medium transition-all shadow-lg shadow-indigo-500/20 hover:shadow-indigo-500/40 flex items-center gap-2"
                                                >
                                                    {cryptoState.isLoading ? <Loader className="animate-spin" size={18} /> : <Wallet size={18} />}
                                                    Request Payment Address
                                                </button>
                                                {cryptoState.isPaused && (
                                                    <p className="text-amber-400 text-xs mt-3 flex items-center gap-1">
                                                        <AlertTriangle size={12} /> Service is currently paused. Payments are disabled.
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* State: Has Payment ID / Premium User */}
                            {cryptoState.paymentId && (
                                <div className="space-y-6 relative z-10">
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        {/* Left: Payment Info */}
                                        <div className="bg-white/5 rounded-lg p-5 border border-white/10">
                                            <h3 className="text-sm font-medium text-slate-400 mb-4 uppercase tracking-wider">Deposit Details</h3>

                                            <div className="space-y-4">
                                                <div>
                                                    <label className="text-xs text-slate-500 block mb-1">Your Unique Deposit Address</label>
                                                    <div className="flex items-center gap-2 bg-black/20 p-2 rounded border border-white/5">
                                                        <code className="text-xs text-indigo-300 break-all font-mono">{cryptoState.depositAddress}</code>
                                                        <button
                                                            onClick={() => copyToClipboard(cryptoState.depositAddress || "")}
                                                            className="p-1 hover:bg-white/10 rounded text-slate-400 hover:text-white transition-colors"
                                                            title="Copy Address"
                                                        >
                                                            <Copy size={14} />
                                                        </button>
                                                    </div>
                                                </div>

                                                <div className="grid grid-cols-2 gap-4">
                                                    <div>
                                                        <label className="text-xs text-slate-500 block mb-1">Payment ID</label>
                                                        <div className="text-slate-200 font-mono">#{cryptoState.paymentId}</div>
                                                    </div>
                                                    <div>
                                                        <label className="text-xs text-slate-500 block mb-1">Current Rate</label>
                                                        <div className="text-slate-200">{cryptoState.requestsPerEGLD.toLocaleString()} req/eGLD</div>
                                                    </div>
                                                </div>

                                                <div className="pt-2 flex gap-2">
                                                    <a
                                                        href={`${cryptoState.walletURL}`}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex-1 bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 text-xs py-2 rounded flex items-center justify-center gap-2 transition-colors border border-indigo-500/20"
                                                    >
                                                        <Wallet size={14} /> Open Web Wallet
                                                    </a>
                                                    <a
                                                        href={`${cryptoState.explorerURL}/accounts/${cryptoState.depositAddress}`}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex-1 bg-slate-700/50 hover:bg-slate-700 text-slate-300 text-xs py-2 rounded flex items-center justify-center gap-2 transition-colors"
                                                    >
                                                        <ExternalLink size={14} /> Explorer
                                                    </a>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Right: Balance & Status */}
                                        <div className="bg-white/5 rounded-lg p-5 border border-white/10 flex flex-col justify-between">
                                            <div>
                                                <h3 className="text-sm font-medium text-slate-400 mb-4 uppercase tracking-wider">Requests Balance</h3>

                                                <div className="mb-2">
                                                    <span className="text-3xl font-bold text-white">{cryptoState.numberOfRequests.toLocaleString()}</span>
                                                    <span className="text-slate-500 text-sm ml-2">available credits</span>
                                                </div>

                                                {/* Visual Bar */}
                                                <div className="w-full bg-black/30 h-2 rounded-full mb-4 overflow-hidden">
                                                    <div
                                                        className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500"
                                                        style={{ width: users[user.username.toLowerCase()].AccountType === 'premium' ? '100%' : '5%' }}
                                                    ></div>
                                                </div>

                                                {users[user.username.toLowerCase()].AccountType === 'premium' ? (
                                                    <p className="text-xs text-emerald-400 flex items-center gap-1">
                                                        <Check size={12} /> Account is Premium active
                                                    </p>
                                                ) : (
                                                    <p className="text-xs text-amber-400 flex items-center gap-1">
                                                        <Loader className="animate-spin" size={12} /> Waiting for deposit...
                                                    </p>
                                                )}
                                            </div>

                                            <div className="mt-6 pt-4 border-t border-white/5">
                                                <button
                                                    onClick={handleRefreshBalance}
                                                    className="w-full bg-white/5 hover:bg-white/10 text-slate-300 py-2 rounded text-xs flex items-center justify-center gap-2 transition-colors"
                                                >
                                                    <RefreshCw size={14} className={cryptoState.isLoading ? "animate-spin" : ""} />
                                                    Refresh Contract Balance
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    {cryptoState.isPaused && (
                                        <div className="bg-amber-500/10 border border-amber-500/20 text-amber-200 p-3 rounded text-sm text-center">
                                            ⚠️ The smart contract is currently paused. Deposits made now will be processed when it resumes.
                                        </div>
                                    )}
                                </div>
                            )}

                        </div>
                    )}

                    {/* Keys Panel */}
                    <div className="glass-panel p-6 flex flex-col h-full col-span-1 lg:col-span-2">
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4 md:gap-0">
                            <h2 className="text-xl font-semibold flex items-center gap-2">
                                <Key className="text-indigo-400" /> Access Keys
                            </h2>
                            <div className="flex gap-3 w-full md:w-auto flex-wrap md:flex-nowrap">
                                <button
                                    onClick={() => window.location.reload()}
                                    className="flex-1 md:flex-none justify-center bg-slate-700 hover:bg-slate-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all whitespace-nowrap"
                                >
                                    <RotateCcw size={16} /> Refresh Data
                                </button>
                                <button
                                    onClick={() => setShowKeyModal(true)}
                                    className="flex-1 md:flex-none justify-center bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg flex items-center gap-2 text-sm transition-all whitespace-nowrap"
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
                                            <th className="py-3 px-4">Limits</th>
                                            <th className="py-3 px-4">Current Usage</th>
                                            <th className="py-3 px-4"></th>
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
                                                    {u.MaxRequests}
                                                </td>
                                                <td className="py-3 px-4 text-slate-300">
                                                    <div className="flex items-center gap-2">
                                                        {u.GlobalCounter}
                                                    </div>
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

            {user.is_admin && (
                <div className="glass-panel p-6 col-span-1 lg:col-span-2">
                    <div className="flex justify-between items-center mb-6 flex-wrap gap-4">
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
                <div className="glass-panel p-6 mt-8">
                    <h2 className="text-xl font-semibold flex items-center gap-2 mb-6">
                        <BookOpen className="text-indigo-400" /> API Documentation
                    </h2>
                    <div className="flex flex-col items-center justify-center gap-4 py-4 w-full max-w-md mx-auto">
                        <a
                            href={`${appInfo.backend}/swagger/`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center justify-between w-full md:w-80 bg-indigo-600 hover:bg-indigo-500 text-white px-6 py-3 rounded-lg font-medium transition-all shadow-lg shadow-indigo-500/20 hover:shadow-indigo-500/40"
                        >
                            <span>Go to Swagger Interface</span>
                            <ExternalLink size={18} />
                        </a>
                        <a
                            href="https://docs.multiversx.com/integrators/deep-history-squad/"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center justify-between w-full md:w-80 bg-slate-700 hover:bg-slate-600 text-white px-6 py-3 rounded-lg font-medium transition-all shadow-lg shadow-slate-500/10 hover:shadow-slate-500/20"
                        >
                            <span>Official Deep History Docs</span>
                            <ExternalLink size={18} />
                        </a>
                        <div className="text-center w-full border-t border-white/5 pt-6">
                            <p style={{ fontSize: '0.8rem' }} className="text-slate-500">
                                Build {appInfo.version} | <a href="https://github.com/iulianpascalau/mx-epoch-proxy-go" className="hover:text-slate-400 underline decoration-slate-600 underline-offset-2" target="_blank" rel="noopener noreferrer">Solution</a>
                            </p>
                        </div>

                    </div>
                </div>
            )}

            {/* Key Modal */}
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

            {/* User Modal */}
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
                                            autoCapitalize="none"
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

            {/* Manage User Keys Modal */}
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
                            <form onSubmit={handleCreateKey} className="mb-6 flex flex-col md:flex-row gap-2">
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
