import { enqueueSnackbar } from 'notistack';

const NotificationService = {
    notifySuccess: (t: string) => {
        enqueueSnackbar<"success">(t);
    },
    notifyError: (t: string) => {
        enqueueSnackbar<"error">(t, {
            preventDuplicate: true
        });
    },
}
export default NotificationService;
