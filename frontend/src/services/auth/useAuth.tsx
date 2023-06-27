import { createContext, useContext, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { User } from "../../domain/User";
import pb from "../PocketBase";
import { useLocalStorage } from "./useLocalStorage";
const AuthContext = createContext<{
    user: User | null,
    login: (username: string, password: string) => Promise<void>,
    loginWithOauth2: (provider: string) => Promise<void>,
    logout: () => Promise<void>
}>({
    user: null,
    login: () => { return Promise.reject() },
    loginWithOauth2: () => { return Promise.reject() },
    logout: () => { return Promise.reject() }
});

export const AuthProvider = ({ children }: { children: JSX.Element }) => {
    const [user, setUser] = useLocalStorage("user", null);
    const navigate = useNavigate();

    // call this function when you want to authenticate the user
    const login = async (username: string, password: string) => {
        return pb.collection("users").authWithPassword(username, password)
            .then((resp) => {
                setUser({
                    id: resp.record.id,
                    username: username,
                });
                navigate("/");
            });
    };

    const loginWithOauth2 = async (provider: string) => {
        return pb.collection('users').authWithOAuth2({ provider: provider })
            .then((resp) => {
                setUser({
                    id: resp.record.id,
                    username: resp.record["username"],
                });
                navigate("/");
            });
    };

    // call this function to sign out logged in user
    const logout = () => {
        setUser(null);
        navigate("/login", { replace: true });
        return Promise.resolve();
    };

    const value = useMemo(
        () => ({
            user,
            login,
            loginWithOauth2,
            logout
        }),
        [user]
    );
    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
    return useContext(AuthContext);
};