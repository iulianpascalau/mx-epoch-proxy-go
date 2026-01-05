import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './Login';
import { Register } from './Register';
import { Dashboard } from './Dashboard';

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/" element={<Dashboard />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </HashRouter>
  );
}

export default App;
