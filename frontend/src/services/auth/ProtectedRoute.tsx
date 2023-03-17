import { Navigate } from "react-router-dom";
import { useAuth } from "./useAuth";

export const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
    const { user } = useAuth();
    if (!user) {
        // user is not authenticated
        return <Navigate to="/login" />;
    }
    return children;
};