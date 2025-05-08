import axios from 'axios';

const API_BASE = '/api/v1';  // Updated to use versioned API

// Request and Response Types
export interface DiagnosticRequest {
  target: string;
  [key: string]: any;
}

// Common Types
export interface ErrorResponse {
  error: string;
  code?: number;
  type?: string;
  details?: string;
  validations?: Record<string, string>;
}

// DNS Types
export interface DNSResponse {
  records: Record<string, string[]>;
  timing?: string;
  error?: string;
}

// Blacklist Types
export interface BlacklistResponse {
  ip: string;
  listedOn: Record<string, string>;
  error?: string;
}

// SMTP Types
export interface SMTPResponse {
  connected: boolean;
  supportsStartTLS?: boolean;
  error?: string;
}

export interface SMTPConnectionResponse {
  host: string;
  port: number;
  connected: boolean;
  latency?: string;
  supportsStartTLS?: boolean;
  authMethods?: string[];
  banner?: string;
  error?: string;
}

export interface SMTPRelayTestRequest {
  host: string;
  port?: number;
  fromAddress: string;
  toAddress: string;
  timeout?: number;
  authentication?: boolean;
  username?: string;
  password?: string;
}

export interface SMTPRelayTestResponse {
  host: string;
  port: number;
  isOpenRelay: boolean;
  authRequired?: boolean;
  responseCode?: number;
  responseText?: string;
  testDetails?: string;
  error?: string;
}

// Email Auth Types
export interface SPFResponse {
  domain: string;
  hasRecord: boolean;
  record?: string;
  isValid: boolean;
  mechanisms?: string[];
  error?: string;
}

export interface DKIMResponse {
  domain: string;
  selector: string;
  hasRecords: boolean;
  records?: Record<string, string>;
  isValid: boolean;
  error?: string;
}

export interface DMARCResponse {
  domain: string;
  hasRecord: boolean;
  record?: string;
  isValid: boolean;
  policy?: string;
  subdomainPolicy?: string;
  percentage?: number;
  error?: string;
}

// Network Tools Types
export interface PingResponse {
  target: string;
  resolvedIP?: string;
  success: boolean;
  rtts?: string[];
  avgRTT?: string;
  minRTT?: string;
  maxRTT?: string;
  packetsSent?: number;
  packetsReceived?: number;
  packetLoss?: number;
  error?: string;
  rawOutput?: string;
}

export interface TracerouteHopResponse {
  number: number;
  ip: string;
  hostname?: string;
  rtt?: string;
  error?: string;
}

export interface TracerouteResponse {
  target: string;
  resolvedIP?: string;
  hops: TracerouteHopResponse[];
  targetReached: boolean;
  error?: string;
  rawOutput?: string;
}

export interface WHOISResponse {
  target: string;
  registrar?: string;
  createdDate?: string;
  expirationDate?: string;
  nameServers?: string[];
  rawData?: string;
  error?: string;
}

// API Documentation Type
export interface APIDocs {
  version: string;
  description: string;
  baseUrl: string;
  groups: APIGroupDoc[];
}

export interface APIGroupDoc {
  name: string;
  description: string;
  endpoints: APIEndpointDoc[];
}

export interface APIEndpointDoc {
  path: string;
  method: string;
  description: string;
  parameters?: APIParameterDoc[];
  request?: any;
  response?: any;
  example?: string;
}

export interface APIParameterDoc {
  name: string;
  in: string;
  description: string;
  required: boolean;
  type: string;
  example?: any;
}

// Helper function for error handling
function handleAxiosError(error: any): never {
  if (error.response?.data) {
    throw error.response.data;
  }
  throw new Error(error.message || 'An unknown error occurred');
}

// Legacy API function - kept for backward compatibility
export async function runDiagnostic(tool: string, input: string, token?: string) {
  try {
    const response = await axios.post(
      `/api/${tool}`,  // Using legacy non-versioned API
      { target: input },
      token ? { headers: { Authorization: `Bearer ${token}` } } : undefined
    );
    return response.data;
  } catch (error: any) {
    throw error.response?.data?.error || error.message;
  }
}

// DNS API Functions
export async function dnsLookup(target: string): Promise<DNSResponse> {
  try {
    const response = await axios.post(`${API_BASE}/dns/${encodeURIComponent(target)}`);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// Blacklist API Functions
export async function blacklistCheck(target: string): Promise<BlacklistResponse> {
  try {
    const response = await axios.post(`${API_BASE}/blacklist/${encodeURIComponent(target)}` );
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// SMTP API Functions
export async function smtpCheck(target: string): Promise<SMTPResponse> {
  try {
    const response = await axios.post(`${API_BASE}/smtp`, { target });
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function smtpConnect(host: string, port?: number, timeout?: string): Promise<SMTPConnectionResponse> {
  try {
    // Build the query parameters
    const params = new URLSearchParams();
    if (port !== undefined) params.append('port', port.toString());
    if (timeout !== undefined) params.append('timeout', timeout);
    
    const url = `${API_BASE}/smtp/connect/${encodeURIComponent(host)}`;
    const fullUrl = params.toString() ? `${url}?${params.toString()}` : url;
    
    const response = await axios.post(fullUrl);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function smtpStartTLS(host: string, timeout?: string): Promise<Record<string, any>> {
  try {
    // Build the query parameters
    const params = new URLSearchParams();
    if (timeout !== undefined) params.append('timeout', timeout);
    
    const url = `${API_BASE}/smtp/starttls/${encodeURIComponent(host)}`;
    const fullUrl = params.toString() ? `${url}?${params.toString()}` : url;
    
    const response = await axios.post(fullUrl);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function smtpRelayTest(request: SMTPRelayTestRequest): Promise<SMTPRelayTestResponse> {
  try {
    const response = await axios.post(`${API_BASE}/smtp/relay-test`, request);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// Email Authentication API Functions
export async function spfCheck(domain: string): Promise<SPFResponse> {
  try {
    const response = await axios.post(`${API_BASE}/auth/spf/${encodeURIComponent(domain)}`);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function dkimCheck(domain: string, selector?: string): Promise<DKIMResponse> {
  try {
    // If selector is provided, include it in the URL, otherwise the backend will use defaults
    const url = selector 
      ? `${API_BASE}/auth/dkim/${encodeURIComponent(domain)}/${encodeURIComponent(selector)}`
      : `${API_BASE}/auth/dkim/${encodeURIComponent(domain)}`;
    
    const response = await axios.post(url);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function dmarcCheck(domain: string): Promise<DMARCResponse> {
  try {
    const response = await axios.post(`${API_BASE}/auth/dmarc/${encodeURIComponent(domain)}`);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// Network Tools API Functions
export async function pingHost(host: string, count?: number, timeout?: number): Promise<PingResponse> {
  try {
    const params = new URLSearchParams();
    params.append('host', host); // Add host as query param too
    if (count !== undefined) params.append('count', count.toString());
    if (timeout !== undefined) params.append('timeout', timeout.toString());
    
    const url = `${API_BASE}/network/ping/${encodeURIComponent(host)}`;
    const fullUrl = `${url}?${params.toString()}`;
    
    const response = await axios.post(fullUrl);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function tracerouteHost(host: string, maxHops?: number, timeout?: number): Promise<TracerouteResponse> {
  try {
    const params = new URLSearchParams();
    params.append('host', host); // Add host as query param too
    if (maxHops !== undefined) params.append('maxHops', maxHops.toString());
    if (timeout !== undefined) params.append('timeout', timeout.toString());
    
    const url = `${API_BASE}/network/traceroute/${encodeURIComponent(host)}`;
    const fullUrl = `${url}?${params.toString()}`;
    
    const response = await axios.post(fullUrl);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function whoisLookup(domain: string): Promise<WHOISResponse> {
  try {
    const params = new URLSearchParams();
    params.append('domain', domain); // Add domain as query param too
    
    const url = `${API_BASE}/network/whois/${encodeURIComponent(domain)}`;
    const fullUrl = `${url}?${params.toString()}`;
    
    const response = await axios.post(fullUrl);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// API Documentation
export async function getApiDocs(group?: string): Promise<APIDocs | APIGroupDoc> {
  try {
    const url = group 
      ? `${API_BASE}/docs?group=${encodeURIComponent(group)}`
      : `${API_BASE}/docs`;
    const response = await axios.get(url);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// Auth Functions
export async function login(username: string, password: string) {
  try {
    const response = await axios.post(`${API_BASE}/auth/login`, { username, password });
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

export async function getRateLimitStatus(token?: string) {
  try {
    const response = await axios.get(
      `${API_BASE}/ratelimit`, 
      token ? { headers: { Authorization: `Bearer ${token}` } } : undefined
    );
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}

// Health Check
export async function getHealth() {
  try {
    const response = await axios.get(`${API_BASE}/health`);
    return response.data;
  } catch (error: any) {
    return handleAxiosError(error);
  }
}
