import PocketBase from 'pocketbase';
import NotificationService from './NotificationService';

const pb = new PocketBase(process.env.NODE_ENV === "development" ? "http://127.0.0.1:8090" : undefined);


pb.afterSend = function (response: Response, data: any) {
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