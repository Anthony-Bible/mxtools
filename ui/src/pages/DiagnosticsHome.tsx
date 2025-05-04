import React from 'react';

const DiagnosticsHome: React.FC = () => (
  <div>
    <h1>Diagnostics Dashboard</h1>
    <ul>
      <li><a href="/dns">DNS Diagnostics</a></li>
      <li><a href="/blacklist">Blacklist Check</a></li>
      <li><a href="/smtp">SMTP Diagnostics</a></li>
      <li><a href="/auth">Email Authentication</a></li>
      <li><a href="/network">Network Tools</a></li>
    </ul>
  </div>
);

export default DiagnosticsHome;
