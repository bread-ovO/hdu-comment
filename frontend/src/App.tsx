import { Suspense, lazy } from 'react';
import { Route, Routes } from 'react-router-dom';
import AppHeader from './components/Layout/Header';
import ProtectedRoute from './components/ProtectedRoute';
import './styles/global.css';
import './styles/home.css';
import './styles/responsive.css';
import './styles/lazy-image.css';
import './styles/polish.css';

const NewHome = lazy(() => import('./pages/NewHome'));
const Login = lazy(() => import('./pages/Login'));
const Register = lazy(() => import('./pages/Register'));
const SubmitReview = lazy(() => import('./pages/SubmitReview'));
const MyReviews = lazy(() => import('./pages/MyReviews'));
const ReviewDetail = lazy(() => import('./pages/ReviewDetail'));
const AdminPending = lazy(() => import('./pages/AdminPending'));
const AdminUsers = lazy(() => import('./pages/AdminUsers'));
const NotFound = lazy(() => import('./pages/NotFound'));

const App = () => {
  return (
    <div className="app-container">
      <AppHeader />
      <main className="app-main">
        <Suspense fallback={<div className="loading-container">页面加载中...</div>}>
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
              <Route path="/admin/users" element={<AdminUsers />} />
            </Route>

            <Route path="*" element={<NotFound />} />
          </Routes>
        </Suspense>
      </main>
    </div>
  );
};

export default App;
