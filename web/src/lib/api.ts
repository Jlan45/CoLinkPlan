import axios from 'axios';

// Get base URL from env or use relative path for proxying
const API_URL = import.meta.env.VITE_API_URL || '/api';

export const api = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json',
    },
});

// Intercept requests to add auth token
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Intercept responses to handle 401s
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            // Redirect handled by specific components or hooks
        }
        return Promise.reject(error);
    }
);
