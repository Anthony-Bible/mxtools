import React from 'react';
import { SPFResponse, DMARCResponse, DKIMResponse } from '../api/diagnostics';

interface AuthResult {
  spf?: SPFResponse;
  dmarc?: DMARCResponse;
  dkim?: DKIMResponse;
  ping?: PingResult;
  whois?: WhoisResult;
  [key: string]: any;
}

interface ResultViewProps {
  result: AuthResult;
}

// Define some type interfaces to help with type checking
interface TracerouteHop {
  hop?: number;
  Hop?: number;
  number?: number;
  Number?: number;
  address?: string;
  Address?: string;
  ip?: string;
  IP?: string;
  rtt?: string;
  RTT?: string;
  error?: string;
  Error?: string;
}

interface PingResult {
  target: string;
  resolvedIP?: string;
  success: boolean;
  rtts: string[];
  avgRTT?: string;
  minRTT?: string;
  maxRTT?: string;
  packetsSent?: number;
  packetsReceived?: number;
  packetLoss?: number;
  error?: string;
}

interface WhoisResult {
  target: string;
  registrar?: string;
  createdDate?: string;
  expirationDate?: string;
  nameServers?: string[];
  rawData?: string;
  error?: string;
}

const ResultView: React.FC<ResultViewProps> = ({ result }) => {
  // Helper function to render validation status
  const renderStatus = (isValid: boolean, hasRecord: boolean) => {
    if (!hasRecord) {
      return <span className="badge bg-secondary">No Record</span>;
    }
    return isValid ? 
      <span className="badge bg-success">Valid</span> : 
      <span className="badge bg-danger">Invalid</span>;
  };

  // Check if we're displaying email auth results
  const isAuthResult = result.spf || result.dmarc || result.dkim;

  if (isAuthResult) {
    return (
      <div className="result-container">
        {/* SPF Results */}
        {result.spf && (
          <div className="card mb-3">
            <div className="card-header d-flex justify-content-between align-items-center">
              <h5>SPF Record {renderStatus(result.spf.isValid, result.spf.hasRecord)}</h5>
            </div>
            <div className="card-body">
              {result.spf.error ? (
                <div className="alert alert-danger">{result.spf.error}</div>
              ) : (
                <>
                  <p><strong>Domain:</strong> {result.spf.domain}</p>
                  {result.spf.hasRecord && (
                    <>
                      <p><strong>Record:</strong> <code>{result.spf.record}</code></p>
                      {result.spf.mechanisms && (
                        <div>
                          <strong>Mechanisms:</strong>
                          <ul>
                            {result.spf.mechanisms.map((mechanism: string, i: number) => (
                              <li key={i}><code>{mechanism}</code></li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </>
                  )}
                </>
              )}
            </div>
          </div>
        )}

        {/* DKIM Results */}
        {result.dkim && (
          <div className="card mb-3">
            <div className="card-header d-flex justify-content-between align-items-center">
              <h5>DKIM Record {renderStatus(result.dkim.isValid || false, (result.dkim.hasRecords || (result.dkim.results && result.dkim.results.length > 0)) || false)}</h5>
            </div>
            <div className="card-body">
              {result.dkim.error ? (
                <div className="alert alert-danger">{result.dkim.error}</div>
              ) : (
                <>
                  <p><strong>Domain:</strong> {result.dkim.domain}</p>
                  
                  {/* Handle combined results with multiple selectors */}
                  {result.dkim.results && result.dkim.results.length > 0 ? (
                    <>
                      <p><strong>Selectors:</strong> {result.dkim.selectors?.join(', ')}</p>
                      {result.dkim.results.map((dkimResult: any, index: number) => (
                        <div key={index} className="mb-3 p-3 border rounded">
                          <h6>Selector: {dkimResult.selector}</h6>
                          {dkimResult.hasRecords && dkimResult.records && (
                            <div>
                              <strong>Records:</strong>
                              <div className="records-container">
                                {Object.entries(dkimResult.records).map((entry: [string, unknown], i: number) => {
                                  const [key, value] = entry;
                                  return (
                                    <div key={i} className="mb-2">
                                      <strong>{key}:</strong> <pre className="mt-2">{String(value)}</pre>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>
                          )}
                          <div className="mt-2">
                            <strong>Status:</strong> {renderStatus(dkimResult.isValid || false, dkimResult.hasRecords || false)}
                          </div>
                        </div>
                      ))}
                    </>
                  ) : (
                    <>
                      <p><strong>Selector:</strong> {result.dkim.selector}</p>
                      {result.dkim.hasRecords && result.dkim.records && (
                        <div>
                          <strong>Records:</strong>
                          <div className="records-container">
                            {Object.entries(result.dkim.records).map((entry: [string, unknown], i: number) => {
                              const [key, value] = entry;
                              return (
                                <div key={i} className="mb-2">
                                  <strong>{key}:</strong> <pre className="mt-2">{String(value)}</pre>
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      )}
                    </>
                  )}
                </>
              )}
            </div>
          </div>
        )}

        {/* DMARC Results */}
        {result.dmarc && (
          <div className="card mb-3">
            <div className="card-header d-flex justify-content-between align-items-center">
              <h5>DMARC Record {renderStatus(result.dmarc.isValid, result.dmarc.hasRecord)}</h5>
            </div>
            <div className="card-body">
              {result.dmarc.error ? (
                <div className="alert alert-danger">{result.dmarc.error}</div>
              ) : (
                <>
                  <p><strong>Domain:</strong> {result.dmarc.domain}</p>
                  {result.dmarc.hasRecord && (
                    <>
                      <p><strong>Record:</strong> <code>{result.dmarc.record}</code></p>
                      {result.dmarc.policy && (
                        <p><strong>Policy:</strong> <code>{result.dmarc.policy}</code></p>
                      )}
                      {result.dmarc.subdomainPolicy && (
                        <p><strong>Subdomain Policy:</strong> <code>{result.dmarc.subdomainPolicy}</code></p>
                      )}
                      {result.dmarc.percentage !== undefined && (
                        <p><strong>Percentage:</strong> {result.dmarc.percentage}%</p>
                      )}
                    </>
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </div>
    );
  }

  // For other types of results, display in appropriate table format or fallback to JSON
  // @ts-ignore
  return (
    <div className="result-container">
      <div className="card">
        <div className="card-header">
          <h5>Result</h5>
        </div>
        <div className="card-body">
          {/* Special handling for traceroute results */}
          {result && ((result.hops && Array.isArray(result.hops)) || 
                     (result.Hops && Array.isArray(result.Hops))) ? (
            <div>
              <h4>Traceroute Results</h4>
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
                  {(result.hops || result.Hops)
                    .map((hop: TracerouteHop, i: number) => {
                      const hopNumber = hop.hop || hop.Hop || hop.number || hop.Number || i+1;
                      const address = hop.address || hop.Address || hop.ip || hop.IP || '*';
                      const rtt = hop.rtt || hop.RTT || '*';
                      const error = hop.error || hop.Error || '';
                      
                      return (
                        <tr key={i}>
                          <td style={{ padding: '8px 12px' }}>{hopNumber}</td>
                          <td style={{ padding: '8px 12px' }}>{address}</td>
                          <td style={{ padding: '8px 12px' }}>{rtt}</td>
                          <td style={{ padding: '8px 12px' }}>{error}</td>
                        </tr>
                      );
                    })}
                </tbody>
              </table>
            </div>
          ) : /* Special handling for ping results */
          result && result.rtts && Array.isArray(result.rtts) ? (
            <div>
              <h4>Ping Results</h4>
              <table className="table table-sm table-bordered" style={{ 
                background: '#fafbfc', 
                width: '100%', 
                maxWidth: '800px'
              }}>
                <tbody>
                  <tr>
                    <th style={{ width: '30%' }}>Target</th>
                    <td>{(result as PingResult).target}</td>
                  </tr>
                  {(result as PingResult).resolvedIP && (
                    <tr>
                      <th>Resolved IP</th>
                      <td>{(result as PingResult).resolvedIP}</td>
                    </tr>
                  )}
                  <tr>
                    <th>Status</th>
                    <td>
                      {(result as PingResult).success ? (
                        <span className="badge bg-success">Success</span>
                      ) : (
                        <span className="badge bg-danger">Failed</span>
                      )}
                    </td>
                  </tr>
                  {(result as PingResult).packetsSent && (
                    <tr>
                      <th>Packets</th>
                      <td>{(result as PingResult).packetsReceived} received / {(result as PingResult).packetsSent} sent ({(result as PingResult).packetLoss}% loss)</td>
                    </tr>
                  )}
                  {(result as PingResult).avgRTT && (
                    <tr>
                      <th>Average RTT</th>
                      <td>{(result as PingResult).avgRTT}</td>
                    </tr>
                  )}
                  {(result as PingResult).minRTT && (
                    <tr>
                      <th>Min RTT</th>
                      <td>{(result as PingResult).minRTT}</td>
                    </tr>
                  )}
                  {(result as PingResult).maxRTT && (
                    <tr>
                      <th>Max RTT</th>
                      <td>{(result as PingResult).maxRTT}</td>
                    </tr>
                  )}
                </tbody>
              </table>
              
              {(result as PingResult).rtts.length > 0 && (
                <div className="mt-3">
                  <h5>Individual Ping Results</h5>
                  <table className="table table-sm table-bordered" style={{ 
                    background: '#fafbfc', 
                    width: '100%', 
                    maxWidth: '800px'
                  }}>
                    <thead>
                      <tr>
                        <th style={{ width: '20%' }}>Sequence</th>
                        <th style={{ width: '80%' }}>RTT</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(result as PingResult).rtts.map((rtt: string, i: number) => (
                        <tr key={i}>
                          <td>{i + 1}</td>
                          <td>{rtt}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
              
              {(result as PingResult).error && (
                <div className="alert alert-danger mt-3">
                  {(result as PingResult).error}
                </div>
              )}
            </div>
          ) : /* Special handling for WHOIS results */
          result && (result as WhoisResult).target && typeof (result as WhoisResult).target === 'string' && ((result as WhoisResult).registrar || (result as WhoisResult).nameServers || (result as WhoisResult).rawData) ? (
            <div>
              <h4>WHOIS Results</h4>
              <table className="table table-sm table-bordered" style={{
                background: '#fafbfc',
                width: '100%',
                maxWidth: '800px'
              }}>
                <tbody>
                  <tr>
                    <th style={{ width: '30%' }}>Domain</th>
                    <td>{(result as WhoisResult).target}</td>
                  </tr>
                  {(result as WhoisResult).registrar && (
                    <tr>
                      <th>Registrar</th>
                      <td>{(result as WhoisResult).registrar}</td>
                    </tr>
                  )}
                  {(() => {
                    const nameServers = (result as WhoisResult).nameServers || [];
                    return nameServers.length > 0 && (
                      <tr>
                        <th>Name Servers</th>
                        <td>
                          <ul className="list-unstyled mb-0">
                            {nameServers.map((ns: string, i: number) => (
                              <li key={i}>{ns}</li>
                            ))}
                          </ul>
                        </td>
                      </tr>
                    );
                  })()}
                </tbody>
              </table>

              {(result as WhoisResult).error && (
                <div className="alert alert-danger mt-3">
                  {(result as WhoisResult).error}
                </div>
              )}

              {(result as WhoisResult).rawData && (
                <div className="mt-3">
                  <h5>Raw WHOIS Data</h5>
                  <div className="border rounded p-3" style={{ maxHeight: '300px', overflow: 'auto' }}>
                    <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{(result as WhoisResult).rawData}</pre>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <pre>{JSON.stringify(result, null, 2)}</pre>
          )}
        </div>
      </div>
    </div>
  );
};

export default ResultView;
