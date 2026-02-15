/**
 * Environment variables utility to provide a single point of access 
 * and consistent defaults across the application.
 */

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080/ws';

export default {
    API_BASE_URL,
    WS_BASE_URL
};
