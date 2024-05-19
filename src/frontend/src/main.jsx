import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import App from './components/App.jsx'
import Help from './components/Help.jsx';
import 'bootstrap/dist/css/bootstrap.css';


const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <BrowserRouter>
    <Routes>
      <Route path="/" element={<App />} />
      <Route path="/help" element={<Help />} />
    </Routes>
  </BrowserRouter>
);


