import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Transfers from './pages/Transfers';
import TokenAnalysis from './pages/TokenAnalysis';
import Notifications from './pages/Notifications';
import ApiDocs from './pages/ApiDocs';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="transfers" element={<Transfers />} />
          <Route path="tokens" element={<TokenAnalysis />} />
          <Route path="notifications" element={<Notifications />} />
          <Route path="api-docs" element={<ApiDocs />} />
          {/* Catch all redirect to dashboard */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
