import React, { useState, useEffect } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { 
  pingHost, 
  whoisLookup, 
  PingResponse, 
  TracerouteResponse, 
  WHOISResponse, 
  ErrorResponse, 
  tracerouteHostAsync, 
  getTracerouteResult, 
  TracerouteJob 
} from '../api/diagnostics';

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
  const [traceJobId, setTraceJobId] = useState<string | null>(null);
  const [traceJobStatus, setTraceJobStatus] = useState<string | null>(null);
  const [lastTraceHost, setLastTraceHost] = useState<string>('');
  const [traceJobStartTime, setTraceJobStartTime] = useState<number | null>(null);
  const [timeoutWarning, setTimeoutWarning] = useState<boolean>(false);
  const TIMEOUT_MS = 30000; // 30 seconds

  const handleToolChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTool(e.target.value);
    setResult(null);
    setError(null);
  };

  useEffect(() => {
    if (traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error') {
      if (!traceJobStartTime) setTraceJobStartTime(Date.now());
      const interval = setInterval(async () => {
        try {
          const job: TracerouteJob = await getTracerouteResult(traceJobId);
          setTraceJobStatus(job.status);
          if (job.status === 'complete') {
            setResult(job.result ? {
              ...job.result,
              target: lastTraceHost,
              targetReached: true
            } : null);
            setLoading(false);
            setTraceJobId(null);
            setTraceJobStartTime(null);
            setTimeoutWarning(false);
            clearInterval(interval);
          } else if (job.status === 'error') {
            setError(job.error || 'Traceroute failed');
            setLoading(false);
            setTraceJobId(null);
            setTraceJobStartTime(null);
            setTimeoutWarning(false);
            clearInterval(interval);
          }
        } catch (err: any) {
          setError('Error polling traceroute job');
          setLoading(false);
          setTraceJobId(null);
          setTraceJobStartTime(null);
          setTimeoutWarning(false);
          clearInterval(interval);
        }
      }, 1500);
      return () => clearInterval(interval);
    }
    return undefined;
  }, [traceJobId, traceJobStatus, lastTraceHost, traceJobStartTime]);

  useEffect(() => {
    if (traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && traceJobStartTime) {
      const timeout = setTimeout(() => {
        setTimeoutWarning(true);
        setError('Traceroute is taking longer than expected. Please try again or check your network.');
        setLoading(false);
        setTraceJobId(null);
        setTraceJobStatus(null);
        setTraceJobStartTime(null);
      }, TIMEOUT_MS);
      return () => clearTimeout(timeout);
    } else {
      setTimeoutWarning(false);
    }
  }, [traceJobId, traceJobStatus, traceJobStartTime]);

  const handleSubmit = async (value: string) => {
    if (!value.trim()) {
      setError("Please enter a valid domain or IP address");
      return;
    }
    setLoading(true);
    setError(null);
    setResult(null);
    setTraceJobId(null);
    setTraceJobStatus(null);
    setTraceJobStartTime(null);
    setTimeoutWarning(false);
    if (tool === 'traceroute') setLastTraceHost(value);
    try {
      switch(tool) {
        case 'ping':
          const pingResult = await pingHost(value);
          setResult(pingResult);
          break;
        case 'traceroute':
          const jobResp = await tracerouteHostAsync(value);
          setTraceJobId(jobResp.jobId);
          setTraceJobStatus(jobResp.status);
          setTraceJobStartTime(Date.now());
          break;
        case 'whois':
          const whoisResult = await whoisLookup(value);
          setResult(whoisResult);
          break;
        default:
          throw new Error(`Unsupported network tool: ${tool}`);
      }
    } catch (err: any) {
      setError((err as ErrorResponse).error || err.toString());
      setLoading(false);
    }
  };

  const showLoading = loading || (tool === 'traceroute' && traceJobId && traceJobStatus !== 'complete' && traceJobStatus !== 'error');
  const showStatus = tool === 'traceroute' && traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error';

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
      {showLoading && <LoadingSpinner />}
      {showStatus && (
        <div style={{ margin: '1em 0', color: '#555' }}>
          Traceroute status: <b>{traceJobStatus}</b> {timeoutWarning && <span style={{ color: 'red' }}>(Timeout)</span>}
        </div>
      )}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Network;
