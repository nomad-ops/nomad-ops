import * as React from 'react';
import Grid from '@mui/material/Grid';
import { Subscription } from 'rxjs';
import { Source } from '../domain/Source';
import DeleteIcon from '@mui/icons-material/Delete';
import RealTimeAccess from '../services/RealTimeAccess';
import { Card, CardHeader, CardContent, Typography, CardActions, IconButton, Avatar, Divider, List, ListItem, ListItemText, Fab, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Tooltip, Container, Skeleton, Chip, Paper, Stack, TextField, Box, Drawer, ListItemButton, ListItemIcon, useMediaQuery, Toolbar } from '@mui/material';
import { orange, red, teal } from '@mui/material/colors';
import CheckIcon from '@mui/icons-material/Check';
import ErrorIcon from '@mui/icons-material/Error';
import LoopIcon from '@mui/icons-material/Loop';
import AddIcon from '@mui/icons-material/Add';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import SyncIcon from '@mui/icons-material/Sync';
import NotStartedIcon from '@mui/icons-material/NotStarted';
import HourglassBottomIcon from '@mui/icons-material/HourglassBottom';
import QuestionMarkIcon from '@mui/icons-material/QuestionMark';
import PauseIcon from '@mui/icons-material/Pause';
import InfoIcon from '@mui/icons-material/Info';
import PublishedWithChangesIcon from '@mui/icons-material/PublishedWithChanges';
import { useForm } from "react-hook-form";
import SourceService from '../services/SourceService';
import NotificationService from '../services/NotificationService';
import { Key } from '../domain/Key';
import { FormInputText } from '../components/form-components/FormInputText';
import { FormInputDropdown } from '../components/form-components/FormInputDropdown';
import { FormInputMultiCheckbox } from '../components/form-components/FormInputMultiCheckbox';
import { Team } from '../domain/Team';
import TeamFilter from '../components/TeamFilter';
import { useAuth } from '../services/auth/useAuth';
import SourceDetailDrawer from '../components/SourceDetailDrawer';

interface IFormInput {
    name: string;
    url: string;
    branch: string;
    path: string;
    dataCenter: string;
    namespace: string;
    force: string[];
    teams?: string[];
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

interface IEditTeamsFormInput {
    id: string;
    teams?: string[];
}

const defaultEditTeamsValues = {
    teams: []
};

export default function Sources() {
    const auth = useAuth();

    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = (_ev?: any | undefined, reason?: string | undefined) => {
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
            teams: data.teams,
            region: data.region,

            deployKey: data.deployKey && data.deployKey !== "__empty__" ? data.deployKey : undefined
        })
            .then(() => {
                NotificationService.notifySuccess(`Watching ${data.url}...`);
                setOpen(false);
                reset();
            })
            .catch((err) => {
                console.log(err);
            });

    };

    const [openEditTeams, setOpenEditTeams] = React.useState<IEditTeamsFormInput | undefined>(undefined);

    const handleClickOpenEditTeams = (data: IEditTeamsFormInput) => {
        setOpenEditTeams(data);
    };

    const handleCloseEditTeams = (_ev?: any | undefined, reason?: string | undefined) => {
        if (reason && reason === "backdropClick")
            return;
        setOpenEditTeams(undefined);
    };

    const methodsEditTeams = useForm<IEditTeamsFormInput>({ defaultValues: defaultEditTeamsValues });
    const handleSubmitEditTeams = methodsEditTeams.handleSubmit;
    const resetEditTeams = methodsEditTeams.reset;
    const controlEditTeams = methodsEditTeams.control;
    const setValueEditTeams = methodsEditTeams.setValue;
    const onSubmitEditTeams = (data: IEditTeamsFormInput) => {
        // TODO validate

        console.log(data);
        if (openEditTeams === undefined) {
            return;
        }

        SourceService.updateAssignedTeams(openEditTeams.id, data.teams)
            .then(() => {
                NotificationService.notifySuccess(`Updated assigned teams...`);
                setOpenEditTeams(undefined);
                resetEditTeams();
            })
            .catch((err: any) => {
                console.log(err);
            });

    };

    const [sources, setSources] = React.useState<Source[] | undefined>(undefined);

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<Source>("sources").then((s) => {
            sub = s.subscribe((sources) => {
                if (sources === undefined) {
                    setSources(undefined);
                    return;
                }
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
                if (keys === undefined) {
                    setKeys(undefined);
                    return;
                }
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

    const [teams, setTeams] = React.useState<Team[] | undefined>(undefined);
    const [teamFilter, setTeamFilter] = React.useState<{
        [id: string]: Team
    }>({});

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<Team>("teams").then((s) => {
            sub = s.subscribe((teams) => {
                if (teams === undefined) {
                    setTeams(undefined);
                    return;
                }
                var objArray: Team[] = [];
                for (const key in teams) {
                    if (Object.prototype.hasOwnProperty.call(teams, key)) {
                        const element = teams[key];
                        objArray.push(element);
                    }
                }
                objArray.sort((a, b) => {
                    return a.name.localeCompare(b.name);
                });

                setTeams(objArray);
            });
        })
        return () => {
            sub?.unsubscribe();
        };
    }, []);

    const [searchTerm, setSearchTerm] = React.useState<string>('');

    const [detailDrawerState, setDetailDrawerState] = React.useState<{
        open: boolean,
        source: Source | undefined
    }>({
        open: false,
        source: undefined
    });

    const toggleDrawer =
        (open: boolean, source: Source | undefined) =>
            (event: React.KeyboardEvent | React.MouseEvent) => {
                if (
                    event.type === 'keydown' &&
                    ((event as React.KeyboardEvent).key === 'Tab' ||
                        (event as React.KeyboardEvent).key === 'Shift')
                ) {
                    return;
                }

                setDetailDrawerState({ ...detailDrawerState, open: open, source: source });
            };

    return <div>
        <Paper>
            <List component={Stack} direction="row" sx={{ paddingLeft: "4px" }}>
                <TextField
                    name="search"
                    autoFocus={true}
                    size="small"
                    onChange={(ev: any) => { setSearchTerm(ev.target.value) }}
                    type={"text"}
                    value={searchTerm}
                    label={"Search"}
                    variant="outlined"
                    margin="dense"
                />
                <TeamFilter teams={teams} userID={auth.user?.id} onChange={(selectedTeams) => {
                    const tf: {
                        [id: string]: Team
                    } = {};
                    for (let i = 0; i < selectedTeams.length; i++) {
                        const element = selectedTeams[i];
                        tf[element.id as string] = element;
                    }
                    setTeamFilter(tf);
                }} />
            </List>
        </Paper>
        <Grid container spacing={3} sx={{ marginTop: "0px" }}>
            {sources ? sources.filter((src) => {
                if (Object.keys(teamFilter).length === 0) {
                    return true;
                }

                if (src.teams?.find((srcTeam) => {
                    return teamFilter[srcTeam] !== undefined;
                })) {
                    return true;
                }

                return false;
            }).filter((src) => {
                if (searchTerm === "") {
                    return true;
                }
                return src.name.toLowerCase().includes(searchTerm.toLowerCase());
            }).map((k) => {
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
                    case "outofsync":
                        avatar = <Avatar sx={{ bgcolor: orange[500] }} aria-label="recipe">
                            <PublishedWithChangesIcon />
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
                            action={
                                k.status && k.status.jobs && Object.keys(k.status.jobs).length > 0 ? <IconButton
                                    color='primary'
                                    onClick={toggleDrawer(true, k)}
                                    aria-label="info">
                                    <InfoIcon />
                                </IconButton> : undefined
                            }
                        />
                        <CardContent>
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
                            </List>
                            <Grid container spacing={3}>
                                <Grid item xs={12} md={6} lg={6}>
                                    <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }} disablePadding={true} dense={true}>
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
                            <List component={Stack} direction="row">
                                {k.teams && teams ? k.teams.map((team) => {
                                    for (let i = 0; i < teams.length; i++) {
                                        const element = teams[i];
                                        if (element.id === team) {
                                            return element;
                                        }
                                    }
                                    throw new Error("expected to find team");
                                }).sort((a, b) => {
                                    if (a === undefined || b === undefined) {
                                        return 0;
                                    }
                                    return a.name.localeCompare(b.name);
                                }).map((team) => {
                                    team = team as Team;
                                    return (
                                        <ListItem key={team.id} sx={{ paddingRight: "0px", paddingLeft: "4px" }}>
                                            <Chip
                                                size='small'
                                                label={team.name}
                                            />
                                        </ListItem>
                                    );
                                }) : undefined}
                            </List>
                            <span style={{ flexGrow: "1" }}></span>

                            <Tooltip title="Edit teams">
                                <IconButton aria-label="teams" color='primary' onClick={() => {
                                    if (!k.id) {
                                        return;
                                    }
                                    handleClickOpenEditTeams({
                                        id: k.id,
                                        teams: k.teams
                                    });
                                }}>
                                    <GroupAddIcon />
                                </IconButton>
                            </Tooltip>

                            {k.paused !== true ? <Tooltip title="Pause">
                                <IconButton aria-label="pause" color='primary' onClick={() => {
                                    if (!k.id) {
                                        return;
                                    }
                                    SourceService.pauseSource(k.id, true)
                                        .then(() => {
                                            NotificationService.notifySuccess(`Paused ${k.url} ...`);
                                        });
                                }}>
                                    <PauseIcon />
                                </IconButton>
                            </Tooltip> : undefined}
                            {k.paused === true ? <Tooltip title="Resume">
                                <IconButton aria-label="resume" color='primary' onClick={() => {
                                    if (!k.id) {
                                        return;
                                    }
                                    SourceService.pauseSource(k.id, false)
                                        .then(() => {
                                            NotificationService.notifySuccess(`Resumed watch on ${k.url} ...`);
                                        });
                                }}>
                                    <NotStartedIcon />
                                </IconButton>
                            </Tooltip> : undefined}
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
            {sources && sources.length === 0 ? <Container sx={{ textAlign: "center" }}>
                <Typography>
                    No sources configured
                </Typography>
            </Container> : undefined}
            {sources === undefined ? <React.Fragment>
                <Grid item xs={12} md={6} lg={4}>
                    <Skeleton variant="rectangular" height={300} />
                </Grid>
                <Grid item xs={12} md={6} lg={4}>
                    <Skeleton variant="rectangular" height={300} />
                </Grid>
                <Grid item xs={12} md={6} lg={4}>
                    <Skeleton variant="rectangular" height={300} />
                </Grid>
            </React.Fragment> : undefined}
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
                <div>
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
                </div>
                <FormInputMultiCheckbox
                    name="teams"
                    control={control}
                    required={false}
                    label="Assigned teams"
                    setValue={setValue}
                    options={teams ? teams.map((t) => {
                        return {
                            label: t.name,
                            value: t.id as string
                        }
                    }) : []} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleClose() }}>Cancel</Button>
                <Button onClick={handleSubmit(onSubmit)}>Save</Button>
            </DialogActions>
        </Dialog>
        <Dialog open={openEditTeams !== undefined} onClose={handleCloseEditTeams}>
            <DialogTitle>Manage Teams</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Select the teams for this source.
                </DialogContentText>
                <FormInputMultiCheckbox
                    name="teams"
                    control={controlEditTeams}
                    required={false}
                    label="Assigned teams"
                    setValue={setValueEditTeams}
                    options={teams ? teams.map((t) => {
                        const len = openEditTeams?.teams?.filter((inner) => { return inner === t.id; })
                        return {
                            label: t.name,
                            value: t.id as string,
                            selected: len ? len.length > 0 : false
                        }
                    }) : []} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleCloseEditTeams() }}>Cancel</Button>
                <Button onClick={handleSubmitEditTeams(onSubmitEditTeams)}>Save</Button>
            </DialogActions>
        </Dialog>
        {detailDrawerState.source ? <SourceDetailDrawer open={detailDrawerState.open}
            onClose={toggleDrawer(false, undefined)}
            source={detailDrawerState.source}></SourceDetailDrawer> : undefined}
    </div >;
}
