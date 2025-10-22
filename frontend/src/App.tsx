import { Route, Routes } from 'react-router-dom';
import AppHeader from './components/Layout/Header';
import ProtectedRoute from './components/ProtectedRoute';
import NewHome from './pages/NewHome';
import Login from './pages/Login';
import Register from './pages/Register';
import SubmitReview from './pages/SubmitReview';
import MyReviews from './pages/MyReviews';
import ReviewDetail from './pages/ReviewDetail';
import AdminPending from './pages/AdminPending';
import NotFound from './pages/NotFound';
import './styles/global.css';
import './styles/home.css';
import './styles/responsive.css';
import './styles/lazy-image.css';

const App = () => {
  return (
    <div className="app-container">
      <AppHeader />
      <main className="app-main">
        <Routes>
          <Route path="/" element={<NewHome />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/reviews/:id" element={<ReviewDetail />} />

          <Route element={<ProtectedRoute />}>
            <Route path="/submit" element={<SubmitReview />} />
            <Route path="/my" element={<MyReviews />} />
          </Route>

          <Route element={<ProtectedRoute requireAdmin />}>
            <Route path="/admin/reviews" element={<AdminPending />} />
           </Route>
           {/* Fallback route for 404 Not Found */}
           <Route path="*" element={<NotFound />} />
          </Routes>
         </main>
        </div>
       );
      };

export default App;
