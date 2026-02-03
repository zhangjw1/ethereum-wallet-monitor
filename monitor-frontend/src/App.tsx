import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Liquidations from './pages/Liquidations';
import Transfers from './pages/Transfers';
import ApiDocs from './pages/ApiDocs';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="liquidations" element={<Liquidations />} />
          <Route path="transfers" element={<Transfers />} />
          <Route path="api-docs" element={<ApiDocs />} />
          {/* Catch all redirect to dashboard */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
