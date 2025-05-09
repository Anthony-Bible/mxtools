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
  const [cancelled, setCancelled] = useState(false);
  const [traceHops, setTraceHops] = useState<{ hop: number; address: string; rtt: string; error?: string }[]>([]);
  const TIMEOUT_MS = 30000; // 30 seconds

  const handleToolChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTool(e.target.value);
    setResult(null);
    setError(null);
  };

  useEffect(() => {
    if (traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && !cancelled) {
      if (!traceJobStartTime) setTraceJobStartTime(Date.now());
      const interval = setInterval(async () => {
        try {
          const job: TracerouteJob = await getTracerouteResult(traceJobId);
          setTraceJobStatus(job.status);
          // Debug: log backend result and hops
          console.log('Traceroute job.result:', job.result);
          if (job.result && Array.isArray((job.result as any).Hops) && (job.result as any).Hops.length > 0) {
            const hopsRaw = (job.result as any).Hops;
            const hops = hopsRaw.map((h: any) => ({
              hop: h.Hop ?? h.HopNumber ?? h.Number ?? h.hop ?? h.number,
              address: h.Address ?? h.IP,
              rtt: formatRTT(h.RTT),
              error: h.Error,
            })) as { hop: number; address: string; rtt: string; error?: string }[];
            setTraceHops(hops);
            console.log('traceHops state:', hops);
          }
          if (job.status === 'complete') {
            // Extract hops from the final result if available
            if (job.result && Array.isArray((job.result as any).Hops) && (job.result as any).Hops.length > 0) {
              const hopsRaw = (job.result as any).Hops;
              const hops = hopsRaw.map((h: any) => ({
                hop: h.Hop ?? h.HopNumber ?? h.Number ?? h.hop ?? h.number,
                address: h.Address ?? h.IP,
                rtt: formatRTT(h.RTT),
                error: h.Error,
              })) as { hop: number; address: string; rtt: string; error?: string }[];
              setTraceHops(hops);
            }
            
            setResult(job.result ? {
              ...job.result,
              target: lastTraceHost,
              targetReached: true,
              // Ensure hops are properly formatted for ResultView
              hops: (job.result as any).Hops?.map((h: any) => ({
                hop: h.Hop ?? h.HopNumber ?? h.Number ?? h.hop ?? h.number,
                address: h.Address ?? h.IP ?? h.address ?? '',
                rtt: formatRTT(h.RTT ?? h.rtt),
                error: h.Error ?? h.error ?? ''
              })) ?? []
            } : null);
            setLoading(false);
            setTraceJobId(null);
            setTraceJobStartTime(null);
            setTimeoutWarning(false);
            setCancelled(false);
            // Don't clear traceHops here to keep displaying the table
            clearInterval(interval);
          } else if (job.status === 'error') {
            setError(job.error || 'Traceroute failed');
            setLoading(false);
            setTraceJobId(null);
            setTraceJobStartTime(null);
            setTimeoutWarning(false);
            setCancelled(false);
            clearInterval(interval);
          }
        } catch (err: any) {
          setError('Error polling traceroute job');
          setLoading(false);
          setTraceJobId(null);
          setTraceJobStartTime(null);
          setTimeoutWarning(false);
          setCancelled(false);
          clearInterval(interval);
        }
      }, 1500);
      return () => clearInterval(interval);
    }
    return undefined;
  }, [traceJobId, traceJobStatus, lastTraceHost, traceJobStartTime, cancelled]);

  useEffect(() => {
    // Only clear hops when explicitly cancelled, not on job completion or error
    if (cancelled) setTraceHops([]);
  }, [cancelled]);

  useEffect(() => {
    if (traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && traceJobStartTime && !cancelled) {
      const timeout = setTimeout(() => {
        setTimeoutWarning(true);
        setError('Traceroute timed out. You can try again.');
        setLoading(false);
        setTraceJobId(null);
        setTraceJobStatus(null);
        setTraceJobStartTime(null);
        setCancelled(false);
      }, TIMEOUT_MS);
      return () => clearTimeout(timeout);
    } else {
      setTimeoutWarning(false);
    }
  }, [traceJobId, traceJobStatus, traceJobStartTime, cancelled]);

  const handleCancel = () => {
    setCancelled(true);
    setLoading(false);
    setTraceJobId(null);
    setTraceJobStatus(null);
    setTraceJobStartTime(null);
    setTimeoutWarning(false);
    setError('Traceroute cancelled.');
  };

  const handleRetry = () => {
    setResult(null);
    setError(null);
    setCancelled(false);
    setTimeoutWarning(false);
    setTraceJobId(null);
    setTraceJobStatus(null);
    setTraceJobStartTime(null);
    setTraceHops([]);
    // User must resubmit via form
  };

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
    setCancelled(false);
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

  const showLoading = loading || (tool === 'traceroute' && traceJobId && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && !cancelled);
  const showStatus = tool === 'traceroute' && traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && !cancelled;
  const showCancel = tool === 'traceroute' && traceJobId && traceJobStatus && traceJobStatus !== 'complete' && traceJobStatus !== 'error' && !cancelled;
  const showTraceTable = tool === 'traceroute' && traceHops.length > 0;
  const showResult = (tool !== 'traceroute' || !showTraceTable) && result !== null;

  const currentTool = NETWORK_TOOLS.find(t => t.value === tool);

  // Format RTT to milliseconds with consistent precision
  const formatRTT = (rtt: string | undefined | null): string => {
    if (!rtt) return '*';
    
    try {
      // If already in ms format (e.g., "12.345ms")
      if (typeof rtt === 'string' && rtt.endsWith('ms')) {
        // Extract number and format to 2 decimal places
        const ms = parseFloat(rtt.replace('ms', ''));
        return `${ms.toFixed(2)}ms`;
      }
      
      // If in µs format (e.g., "12345µs")
      if (typeof rtt === 'string' && rtt.includes('µs')) {
        // Convert microseconds to milliseconds
        const us = parseFloat(rtt.replace('µs', ''));
        return `${(us / 1000).toFixed(2)}ms`;
      }
      
      // If in ns format (e.g., "12345678ns")
      if (typeof rtt === 'string' && rtt.endsWith('ns')) {
        // Convert nanoseconds to milliseconds
        const ns = parseFloat(rtt.replace('ns', ''));
        return `${(ns / 1000000).toFixed(2)}ms`;
      }
      
      // If it's just a number (likely nanoseconds from Go's time.Duration)
      if (!isNaN(Number(rtt))) {
        const value = Number(rtt);
        // If it's a large number (>100000), assume it's in nanoseconds
        if (value > 100000) {
          return `${(value / 1000000).toFixed(4)}ms`;
        }
        // Otherwise assume it's already in milliseconds
        return `${value.toFixed(2)}ms`;
      }
      
      // If format is unknown, return as is
      return String(rtt);
    } catch (e) {
      console.error('Error formatting RTT:', e, 'RTT value:', rtt);
      return '*';
    }
  };

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
          {showCancel && (
            <button style={{ marginLeft: '1em' }} onClick={handleCancel}>Cancel</button>
          )}
        </div>
      )}
      {/* Live hops table for progressive traceroute */}
      {tool === 'traceroute' && traceHops.length > 0 && (
        (() => { console.log('traceHops (before render):', traceHops); return null; })()
      )}
      {showTraceTable && (
        <div style={{ margin: '1em 0' }}>
          <h4>{traceJobStatus === 'complete' ? 'Traceroute Results' : 'Traceroute Progress'}</h4>
          {showLoading && <LoadingSpinner />}
          <table className="table table-sm table-bordered" style={{ 
            background: '#fafbfc', 
            width: '100%', 
            maxWidth: '800px',
            tableLayout: 'fixed'
          }}>
            <thead>
              <tr>
                <th style={{ width: '10%' }}>Hop</th>
                <th style={{ width: '40%' }}>Address</th>
                <th style={{ width: '25%' }}>RTT (ms)</th>
                <th style={{ width: '25%' }}>Error</th>
              </tr>
            </thead>
            <tbody>
              {traceHops
                .filter(hop => hop.hop !== undefined && hop.hop !== null)
                .map((hop, i) => (
                  <tr key={i}>
                    <td style={{ padding: '8px 12px' }}>{hop.hop}</td>
                    <td style={{ padding: '8px 12px' }}>{hop.address || '*'}</td>
                    <td style={{ padding: '8px 12px' }}>{hop.rtt || '*'}</td>
                    <td style={{ padding: '8px 12px' }}>{hop.error || ''}</td>
                  </tr>
                ))}
            </tbody>
          </table>
          {traceJobStatus === 'complete' && (
            <div className="alert alert-success">
              <strong>Traceroute complete!</strong> Traced route to {lastTraceHost}.
            </div>
          )}
        </div>
      )}
      {error && (
        <div style={{ margin: '1em 0' }}>
          <ErrorAlert message={error} />
          {(timeoutWarning || cancelled) && (
            <button onClick={handleRetry} style={{ marginTop: 8 }}>Try Again</button>
          )}
        </div>
      )}
      {showResult && <ResultView result={result} />}
    </div>
  );
};

export default Network;
