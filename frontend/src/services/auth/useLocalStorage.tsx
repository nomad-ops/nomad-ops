import { useState } from "react";

export const useLocalStorage = (keyName: string, defaultValue: string | null) => {
    const [storedValue, setStoredValue] = useState(() => {
        try {
            const value = window.localStorage.getItem(keyName);
            if (value) {
                return JSON.parse(value);
            } else {
                window.localStorage.setItem(keyName, JSON.stringify(defaultValue));
                return defaultValue;
            }
        } catch (err) {
            return defaultValue;
        }
    });
    const setValue = (newValue: any) => {
        try {
            window.localStorage.setItem(keyName, JSON.stringify(newValue));
        } catch (err) { }
        setStoredValue(newValue);
    };
    return [storedValue, setValue];
};