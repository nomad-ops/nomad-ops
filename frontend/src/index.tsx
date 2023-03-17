import React from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import ErrorPage from './pages/ErrorPage';
import { MainNavItems } from './nav/ListItems';
import { ProtectedRoute } from './services/auth/ProtectedRoute';
import SignIn from './pages/SignIn';
import { AuthProvider } from './services/auth/useAuth';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { SnackbarProvider } from 'notistack';
import RealTimeAccess from './services/RealTimeAccess';
import { Key } from './domain/Key';
import { Source } from './domain/Source';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

const theme = createTheme({
  palette: {
    //primary: { main: "#3a34d2" }
    primary: { main: "#009688" }
  }
  // palette: {
  //   mode: 'dark',
  // },
});
root.render(
  <React.StrictMode>
    <SnackbarProvider />
    <BrowserRouter>
      <AuthProvider>
        <ThemeProvider theme={theme}>
          <Routes>
            <Route path="/" errorElement={<ErrorPage />} element={<ProtectedRoute children={<App />}></ProtectedRoute>}>
              {MainNavItems.map((navItem) => {
                return <Route key={navItem.route} path={navItem.route} element={navItem.page}>
                </Route>
              })}
              <Route
                path="/"
                element={
                  <Navigate replace to={"/" + MainNavItems[0].route} />
                }
              />
            </Route>

            <Route path="/login" element={<SignIn />}>


            </Route>
          </Routes>
        </ThemeProvider>
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
