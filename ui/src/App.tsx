import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import DiagnosticsHome from './pages/DiagnosticsHome';
import DNS from './pages/DNS';
import Blacklist from './pages/Blacklist';
import SMTP from './pages/SMTP';
import Auth from './pages/Auth';
import Network from './pages/Network';

import './App.css';

const App: React.FC = () => (
  <Router>
    <nav>
      <Link to="/">Home</Link> |{' '}
      <Link to="/dns">DNS</Link> |{' '}
      <Link to="/blacklist">Blacklist</Link> |{' '}
      <Link to="/smtp">SMTP</Link> |{' '}
      <Link to="/auth">Auth</Link> |{' '}
      <Link to="/network">Network</Link>
    </nav>
    <Routes>
      <Route path="/" element={<DiagnosticsHome />} />
      <Route path="/dns" element={<DNS />} />
      <Route path="/blacklist" element={<Blacklist />} />
      <Route path="/smtp" element={<SMTP />} />
      <Route path="/auth" element={<Auth />} />
      <Route path="/network" element={<Network />} />
    </Routes>
  </Router>
);

export default App;
