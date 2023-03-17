import PocketBase from 'pocketbase';
import NotificationService from './NotificationService';

const pb = new PocketBase(process.env.NODE_ENV === "development" ? "http://127.0.0.1:8090" : undefined);

pb.beforeSend = function (url: string, options: {
    [key: string]: any;
}): {
    [key: string]: any;
    url?: string | undefined;
    options?: {
        [key: string]: any;
    } | undefined;
} {
    console.log(url, options);

    if (url.includes("/api/realtime") === true) {
        pb.collection("users").authRefresh();
    }

    return {}
}

pb.afterSend = function (response: Response, data: any) {
    console.log(response);
    if (response.status > 299) {
        NotificationService.notifyError(`Could not do request: ${response.url} - ${response.status}`);
    }
    if (response.status === 401) {
        //const auth = useAuth();
        //auth.logout();
        window.location.href = "/login";
    }

    return data;
};

export default pb;