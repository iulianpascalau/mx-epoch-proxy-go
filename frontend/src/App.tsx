import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './Login';
import { Dashboard } from './Dashboard';

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Dashboard />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </HashRouter>
  );
}

export default App;
