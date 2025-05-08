import React from 'react';
import { SPFResponse, DMARCResponse, DKIMResponse } from '../api/diagnostics';

interface AuthResult {
  spf?: SPFResponse;
  dmarc?: DMARCResponse;
  dkim?: DKIMResponse;
  [key: string]: any;
}

interface ResultViewProps {
  result: AuthResult;
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
                            {result.spf.mechanisms.map((mechanism, i) => (
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
              <h5>DKIM Record {renderStatus(result.dkim.isValid, result.dkim.hasRecords)}</h5>
            </div>
            <div className="card-body">
              {result.dkim.error ? (
                <div className="alert alert-danger">{result.dkim.error}</div>
              ) : (
                <>
                  <p><strong>Domain:</strong> {result.dkim.domain}</p>
                  <p><strong>Selector:</strong> {result.dkim.selector}</p>
                  {result.dkim.hasRecords && result.dkim.records && (
                    <div>
                      <strong>Records:</strong>
                      <div className="records-container">
                        {Object.entries(result.dkim.records).map(([key, value], i) => (
                          <div key={i} className="mb-2">
                            <strong>{key}:</strong> <pre className="mt-2">{value}</pre>
                          </div>
                        ))}
                      </div>
                    </div>
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

  // For other types of results, fallback to the original JSON view
  return (
    <div className="result-container">
      <div className="card">
        <div className="card-header">
          <h5>Result</h5>
        </div>
        <div className="card-body">
          <pre>{JSON.stringify(result, null, 2)}</pre>
        </div>
      </div>
    </div>
  );
};

export default ResultView;
