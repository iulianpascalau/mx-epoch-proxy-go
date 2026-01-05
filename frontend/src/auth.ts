export const ACCESS_KEY_KEY = 'access_key';
export const USER_INFO_KEY = 'user_info';

export const getAccessKey = () => sessionStorage.getItem(ACCESS_KEY_KEY);
export const getUserInfo = () => JSON.parse(sessionStorage.getItem(USER_INFO_KEY) || '{}');

export const clearAuth = () => {
    sessionStorage.removeItem(ACCESS_KEY_KEY);
    sessionStorage.removeItem(USER_INFO_KEY);
};

export const setAuth = (token: string, user: any) => {
    sessionStorage.setItem(ACCESS_KEY_KEY, token);
    sessionStorage.setItem(USER_INFO_KEY, JSON.stringify(user));
};

export interface User {
    username: string;
    is_admin: boolean;
}

export const parseJwt = (token: string) => {
    try {
        return JSON.parse(atob(token.split('.')[1]));
    } catch (e) {
        return null;
    }
};
