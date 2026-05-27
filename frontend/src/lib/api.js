import axios from "axios";
import { supabase } from "./supabase";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080",
});

api.interceptors.request.use(async (config) => {
  const { data } = await supabase.auth.getSession();
  const token = data?.session?.access_token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const getCurrentUser = () => api.get("/api/users/me");
export const getFriends = () => api.get("/api/friends");
export const getOrCreateDirectRoom = (friendId) => api.get(`/api/friends/${friendId}/room`);
export const getPendingRequestsCount = () => api.get("/api/friends/requests/count");

export default api;