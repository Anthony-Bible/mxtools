import axios from 'axios';

const API_BASE = '/api';

export async function runDiagnostic(tool: string, input: string, token?: string) {
  try {
    const response = await axios.post(
      `${API_BASE}/${tool}`,
      { target: input }, // changed from input to target
      token ? { headers: { Authorization: `Bearer ${token}` } } : undefined
    );
    return response.data;
  } catch (error: any) {
    throw error.response?.data?.error || error.message;
  }
}

export async function login(username: string, password: string) {
  try {
    const response = await axios.post(`${API_BASE}/auth/login`, { username, password });
    return response.data;
  } catch (error: any) {
    throw error.response?.data?.error || error.message;
  }
}

export async function getRateLimitStatus(token?: string) {
  try {
    const response = await axios.get(`${API_BASE}/ratelimit`, token ? { headers: { Authorization: `Bearer ${token}` } } : undefined);
    return response.data;
  } catch (error: any) {
    throw error.response?.data?.error || error.message;
  }
}
