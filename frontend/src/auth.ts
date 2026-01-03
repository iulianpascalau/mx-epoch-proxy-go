export const ACCESS_KEY_KEY = 'access_key';
export const USER_INFO_KEY = 'user_info';

export const getAccessKey = () => localStorage.getItem(ACCESS_KEY_KEY);
export const getUserInfo = () => JSON.parse(localStorage.getItem(USER_INFO_KEY) || '{}');

export const clearAuth = () => {
    localStorage.removeItem(ACCESS_KEY_KEY);
    localStorage.removeItem(USER_INFO_KEY);
};

export const setAuth = (token: string, user: any) => {
    localStorage.setItem(ACCESS_KEY_KEY, token);
    localStorage.setItem(USER_INFO_KEY, JSON.stringify(user));
};

export interface User {
    username: string;
    is_admin: boolean;
}
