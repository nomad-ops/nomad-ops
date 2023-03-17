import * as React from 'react';
import Grid from '@mui/material/Grid';
import { Subscription } from 'rxjs';
import { Source } from '../domain/Source';
import DeleteIcon from '@mui/icons-material/Delete';
import RealTimeAccess from '../services/RealTimeAccess';
import { Card, CardHeader, CardContent, Typography, CardActions, IconButton, Avatar, Divider, List, ListItem, ListItemText, Fab, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Tooltip } from '@mui/material';
import { orange, red, teal } from '@mui/material/colors';
import CheckIcon from '@mui/icons-material/Check';
import ErrorIcon from '@mui/icons-material/Error';
import LoopIcon from '@mui/icons-material/Loop';
import AddIcon from '@mui/icons-material/Add';
import SyncIcon from '@mui/icons-material/Sync';
import NotStartedIcon from '@mui/icons-material/NotStarted';
import HourglassBottomIcon from '@mui/icons-material/HourglassBottom';
import QuestionMarkIcon from '@mui/icons-material/QuestionMark';
import PauseIcon from '@mui/icons-material/Pause';
import { useForm } from "react-hook-form";
import SourceService from '../services/SourceService';
import NotificationService from '../services/NotificationService';
import { Key } from '../domain/Key';
import { FormInputText } from '../components/form-components/FormInputText';
import { FormInputDropdown } from '../components/form-components/FormInputDropdown';
import { FormInputMultiCheckbox } from '../components/form-components/FormInputMultiCheckbox';

interface IFormInput {
    name: string;
    url: string;
    branch: string;
    path: string;
    dataCenter: string;
    namespace: string;
    force: string[];
    region: string;
    deployKey: string;
}

const defaultValues = {
    name: "",
    url: "",
    branch: "",
    path: "",
    dataCenter: "",
    namespace: "",
    region: "",
    deployKey: "__empty__"
};

export default function Sources() {
    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = (ev?: any | undefined, reason?: string | undefined) => {
        if (reason && reason === "backdropClick")
            return;
        setOpen(false);
    };

    const methods = useForm<IFormInput>({ defaultValues: defaultValues });
    const { handleSubmit, reset, control, setValue } = methods;
    const onSubmit = (data: IFormInput) => {
        // TODO validate

        console.log(data);

        SourceService.createSource({
            name: data.name,
            url: data.url,
            branch: data.branch,
            path: data.path,
            dataCenter: data.dataCenter,
            force: (data.force && data.force.length > 0 && data.force[0] === "true"),
            namespace: data.namespace,
            region: data.region,
            deployKey: data.deployKey && data.deployKey !== "__empty__" ? data.deployKey : undefined
        })
            .then(() => {
                NotificationService.notifySuccess(`Watching ${data.url}...`);
                setOpen(false);
                reset();
            });

    };

    const [sources, setSources] = React.useState<Source[] | undefined>(undefined);

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<Source>("sources").then((s) => {
            sub = s.subscribe((sources) => {
                var objArray: Source[] = [];
                for (const key in sources) {
                    if (Object.prototype.hasOwnProperty.call(sources, key)) {
                        const element = sources[key];
                        objArray.push(element);
                    }
                }
                objArray.sort((a, b) => {
                    return a.url.localeCompare(b.url);
                });
                setSources(objArray);
            });
        })
        return () => {
            sub?.unsubscribe();
        };
    }, []);

    const [keys, setKeys] = React.useState<Key[] | undefined>(undefined);

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<Key>("keys").then((s) => {
            sub = s.subscribe((keys) => {
                var objArray: Key[] = [];
                for (const key in keys) {
                    if (Object.prototype.hasOwnProperty.call(keys, key)) {
                        const element = keys[key];
                        objArray.push(element);
                    }
                }
                objArray.sort((a, b) => {
                    return a.name.localeCompare(b.name);
                });
                objArray.unshift({
                    id: "__empty__",
                    name: "No key",
                    value: "",
                    created: ""
                })
                setKeys(objArray);
            });
        })
        return () => {
            sub?.unsubscribe();
        };
    }, []);

    return <div>
        <Grid container spacing={3}>
            {sources ? sources.map((k) => {
                let avatar = <Avatar sx={{ bgcolor: teal[500] }} aria-label="recipe">
                    <QuestionMarkIcon />
                </Avatar>;
                switch (k.status?.status) {
                    case "synced":
                        avatar = <Avatar sx={{ bgcolor: teal[500] }} aria-label="recipe">
                            <CheckIcon />
                        </Avatar>;
                        break;
                    case "syncedwitherror":
                        avatar = <Avatar sx={{ bgcolor: orange[500] }} aria-label="recipe">
                            <ErrorIcon />
                        </Avatar>;
                        break;
                    case "init":
                        avatar = <Avatar sx={{ bgcolor: teal[500] }} aria-label="recipe">
                            <HourglassBottomIcon />
                        </Avatar>;
                        break;
                    case "syncing":
                        avatar = <Avatar sx={{ bgcolor: teal[500] }} aria-label="recipe">
                            <LoopIcon />
                        </Avatar>;
                        break;
                    case "paused":
                        avatar = <Avatar sx={{ bgcolor: orange[500] }} aria-label="recipe">
                            <PauseIcon />
                        </Avatar>;
                        break;
                    case "error":
                        avatar = <Avatar sx={{ bgcolor: red[500] }} aria-label="recipe">
                            <ErrorIcon />
                        </Avatar>;
                        break;

                    default:
                        break;
                }

                let deployKey = "";
                if (keys) {
                    let found = false;
                    for (let index = 0; index < keys.length; index++) {
                        const element = keys[index];
                        if (k.deployKey && (element.id === k.deployKey || (k.deployKey.length && k.deployKey.length > 0 && k.deployKey[0] === element.id))) {
                            deployKey = element.name;
                            found = true;
                            break;
                        }
                    }
                    if (found === false && (k.deployKey && k.deployKey.length && k.deployKey.length > 0)) {
                        // We expected a key...probably deleted
                        deployKey = "Key was not found. Please fix"
                    }
                }

                return <Grid key={k.id} item xs={12} md={6} lg={4}>


                    <Card>
                        <CardHeader
                            avatar={
                                avatar
                            }
                            title={k.name}
                            subheader={k.created ? new Date(k.created).toLocaleString() : ""}
                        />
                        <CardContent>
                            <Grid container spacing={3}>
                                <Grid item xs={12} md={6} lg={6}>
                                    <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }} disablePadding={true} dense={true}>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="URL:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.url}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Branch:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.branch}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Path:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.path}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Deploy key:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {deployKey === "" ? "No deploy key set" : deployKey}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Force update on commit:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.force ? "true" : "false"}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                    </List>
                                </Grid>
                                <Grid item xs={12} md={6} lg={6}>
                                    <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }} disablePadding={true} dense={true}>

                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Last check:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.status?.lastCheckTime ? new Date(k.status?.lastCheckTime).toLocaleString() : "No check performed"}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Data Center:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.dataCenter}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Namespace:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.namespace ? k.namespace : "No namespace set"}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                        <ListItem alignItems="flex-start">
                                            <ListItemText
                                                primary="Region:"
                                                secondary={
                                                    <React.Fragment>
                                                        <Typography
                                                            sx={{ display: 'inline' }}
                                                            component="span"
                                                            variant="body2"
                                                            color="text.primary"
                                                        >
                                                            {k.region ? k.region : "No region set"}
                                                        </Typography>
                                                    </React.Fragment>
                                                }
                                            />
                                        </ListItem>
                                    </List>
                                </Grid>
                            </Grid>
                            <Divider variant="fullWidth" />
                            <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }} disablePadding={true} dense={true}>
                                <ListItem alignItems="flex-start">
                                    <ListItemText
                                        primary="Message:"
                                        secondary={
                                            <React.Fragment>
                                                <Typography
                                                    sx={{ display: 'inline' }}
                                                    component="span"
                                                    variant="body2"
                                                    color="text.primary"
                                                >
                                                    {k.status && k.status.message ? k.status.message : "No message available"}
                                                </Typography>
                                            </React.Fragment>
                                        }
                                    />
                                </ListItem>
                            </List>
                        </CardContent>
                        <CardActions disableSpacing>
                            <span style={{ flexGrow: "1" }}></span>

                            {k.status && k.status.status !== "paused" ? <Tooltip title="Pause"><IconButton aria-label="pause" color='primary' onClick={() => {
                                if (!k.id) {
                                    return;
                                }
                                SourceService.pauseSource(k.id, true)
                                    .then(() => {
                                        NotificationService.notifySuccess(`Paused ${k.url} ...`);
                                    });
                            }}>
                                <PauseIcon />
                            </IconButton></Tooltip> : undefined}
                            {k.status && k.status.status === "paused" ? <Tooltip title="Resume"><IconButton aria-label="resume" color='primary' onClick={() => {
                                if (!k.id) {
                                    return;
                                }
                                SourceService.pauseSource(k.id, false)
                                    .then(() => {
                                        NotificationService.notifySuccess(`Resumed watch on ${k.url} ...`);
                                    });
                            }}>
                                <NotStartedIcon />
                            </IconButton></Tooltip> : undefined}
                            <Tooltip title="Sync">
                                <IconButton aria-label="sync" color='primary' onClick={() => {
                                    if (!k.id) {
                                        return;
                                    }
                                    SourceService.syncSource(k.id)
                                        .then(() => {
                                            NotificationService.notifySuccess(`Syncing ${k.url} ...`);
                                        });
                                }}>
                                    <SyncIcon />
                                </IconButton>
                            </Tooltip>
                            <Tooltip title="Delete">
                                <IconButton aria-label="delete" color='primary' onClick={() => {
                                    if (window.confirm(`Do you really want to delete ${k.url}?`) === true) {
                                        if (!k.id) {
                                            return;
                                        }
                                        SourceService.deleteSource(k.id)
                                            .then(() => {
                                                NotificationService.notifySuccess(`Removed watcher on ${k.url}`);
                                            });
                                    }
                                }}>
                                    <DeleteIcon />
                                </IconButton>
                            </Tooltip>
                        </CardActions>
                    </Card>
                </Grid>
            }) : undefined}
        </Grid>
        <Fab color="primary" aria-label="add" sx={{
            position: "fixed",
            right: "30px",
            bottom: "30px"
        }} onClick={handleClickOpen}>
            <AddIcon />
        </Fab>
        <Dialog open={open} onClose={handleClose}>
            <DialogTitle>Add new source</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Fill in the form to watch a repository for changes.
                </DialogContentText>
                <FormInputText
                    name="name"
                    control={control}
                    required={true}
                    autoFocus={true}
                    label="Name" />
                <FormInputText
                    name="url"
                    control={control}
                    required={true}
                    label="Repository URL" />
                <FormInputText
                    name="branch"
                    control={control}
                    required={true}
                    label="Branch" />
                <FormInputText
                    name="path"
                    control={control}
                    required={true}
                    label="Path" />
                <FormInputText
                    name="dataCenter"
                    control={control}
                    required={true}
                    label="Data Center" />
                <FormInputText
                    name="namespace"
                    control={control}
                    required={false}
                    label="Namespace" />
                <FormInputDropdown
                    name="deployKey"
                    control={control}
                    required={false}
                    label="Deploy key"
                    options={keys ? keys.filter((key) => {
                        return key.id !== undefined
                    }).map((key) => {
                        return {
                            label: key.name,
                            value: key.id as string
                        }
                    }) : []} />
                <FormInputMultiCheckbox
                    name="force"
                    control={control}
                    required={false}
                    label="Force update on commit?"
                    setValue={setValue}
                    options={[{
                        label: "Yes",
                        value: "true"
                    }]} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleClose() }}>Cancel</Button>
                <Button onClick={handleSubmit(onSubmit)}>Save</Button>
            </DialogActions>
        </Dialog>
    </div >;
}
