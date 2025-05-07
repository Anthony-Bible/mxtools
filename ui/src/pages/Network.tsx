import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { pingHost, tracerouteHost, whoisLookup, PingResponse, TracerouteResponse, WHOISResponse, ErrorResponse } from '../api/diagnostics';

const NETWORK_TOOLS = [
  { label: 'Ping', value: 'ping', placeholder: 'Enter domain or IP for Ping' },
  { label: 'Traceroute', value: 'traceroute', placeholder: 'Enter domain or IP for Traceroute' },
  { label: 'WHOIS', value: 'whois', placeholder: 'Enter domain for WHOIS' },
];

type NetworkResult = PingResponse | TracerouteResponse | WHOISResponse | null;

const Network: React.FC = () => {
  const [tool, setTool] = useState(NETWORK_TOOLS[0].value);
  const [result, setResult] = useState<NetworkResult>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleToolChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTool(e.target.value);
    setResult(null);
    setError(null);
  };

  const handleSubmit = async (value: string) => {
    if (!value.trim()) {
      setError("Please enter a valid domain or IP address");
      return;
    }
    
    setLoading(true);
    setError(null);
    setResult(null);
    
    try {
      // Use the specific API function based on the selected tool
      switch(tool) {
        case 'ping':
          const pingResult = await pingHost(value);
          console.log("Ping result:", pingResult);
          setResult(pingResult);
          break;
        case 'traceroute':
          const traceResult = await tracerouteHost(value);
          console.log("Traceroute result:", traceResult);
          setResult(traceResult);
          break;
        case 'whois':
          const whoisResult = await whoisLookup(value);
          console.log("WHOIS result:", whoisResult);
          setResult(whoisResult);
          break;
        default:
          throw new Error(`Unsupported network tool: ${tool}`);
      }
    } catch (err: any) {
      console.error("API error:", err);
      const errorResponse = err as ErrorResponse;
      setError(errorResponse.error || err.toString());
    } finally {
      setLoading(false);
    }
  };

  const currentTool = NETWORK_TOOLS.find(t => t.value === tool);

  return (
    <div>
      <h2>Network Tools</h2>
      <label htmlFor="network-tool-select">Select Tool:</label>
      <select id="network-tool-select" value={tool} onChange={handleToolChange} style={{ marginBottom: '1rem' }}>
        {NETWORK_TOOLS.map(t => (
          <option key={t.value} value={t.value}>{t.label}</option>
        ))}
      </select>
      <DiagnosticForm onSubmit={handleSubmit} placeholder={currentTool?.placeholder || 'Enter domain or IP'} />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Network;
