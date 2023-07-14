import * as React from 'react';
import Grid from '@mui/material/Grid';
import { Subscription } from 'rxjs';
import { Team, userIsTeamMember } from '../domain/Team';
import DeleteIcon from '@mui/icons-material/Delete';
import RealTimeAccess from '../services/RealTimeAccess';
import { Card, CardHeader, CardContent, Typography, CardActions, IconButton, Avatar, List, ListItem, ListItemText, Fab, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Tooltip, Container, Skeleton, Paper, Stack, TextField } from '@mui/material';
import { teal } from '@mui/material/colors';
import AddIcon from '@mui/icons-material/Add';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import { useForm } from "react-hook-form";
import TeamService from '../services/TeamService';
import NotificationService from '../services/NotificationService';
import { FormInputText } from '../components/form-components/FormInputText';
import { FormInputMultiCheckbox } from '../components/form-components/FormInputMultiCheckbox';
import { User } from '../domain/User';
import { useAuth } from '../services/auth/useAuth';

interface ITeamFormInput {
    name: string;
    members?: string[]
}

const defaultTeamValues = {
    name: ""
};

interface IEditMembersFormInput {
    id: string;
    members?: string[];
}

const defaultEditMembersValues = {
    members: []
};

export default function Teams() {
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

    const methods = useForm<ITeamFormInput>({ defaultValues: defaultTeamValues });
    const { handleSubmit, reset, control, setValue } = methods;
    const onSubmit = (data: ITeamFormInput) => {
        // TODO validate

        console.log(data);

        TeamService.createTeam({
            name: data.name,
            members: data.members
        })
            .then(() => {
                NotificationService.notifySuccess(`Team ${data.name} created...`);
                setOpen(false);
                reset();
            })
            .catch((err) => {
                console.log(err);
            });

    };

    const [openEditMembers, setOpenEditMembers] = React.useState<IEditMembersFormInput | undefined>(undefined);

    const handleClickOpenEditMembers = (data: IEditMembersFormInput) => {
        setOpenEditMembers(data);
    };

    const handleCloseEditMembers = (_ev?: any | undefined, reason?: string | undefined) => {
        if (reason && reason === "backdropClick")
            return;
        setOpenEditMembers(undefined);
    };

    const methodsEditMembers = useForm<IEditMembersFormInput>({ defaultValues: defaultEditMembersValues });
    const handleSubmitEditMembers = methodsEditMembers.handleSubmit;
    const resetEditMembers = methodsEditMembers.reset;
    const controlEditMembers = methodsEditMembers.control;
    const setValueEditMembers = methodsEditMembers.setValue;
    const onSubmitEditMembers = (data: IEditMembersFormInput) => {
        // TODO validate

        console.log(data);
        if (openEditMembers === undefined) {
            return;
        }

        TeamService.updateAssignedMembers(openEditMembers.id, data.members)
            .then(() => {
                NotificationService.notifySuccess(`Updated assigned members...`);
                setOpenEditMembers(undefined);
                resetEditMembers();
            })
            .catch((err: any) => {
                console.log(err);
            });
    };

    const [teams, setTeams] = React.useState<Team[] | undefined>(undefined);

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

    const [users, setUsers] = React.useState<User[] | undefined>(undefined);

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<User>("users").then((s) => {
            sub = s.subscribe((users) => {
                if (users === undefined) {
                    setUsers(undefined);
                    return;
                }
                var objArray: User[] = [];
                for (const key in users) {
                    if (Object.prototype.hasOwnProperty.call(users, key)) {
                        const element = users[key];
                        objArray.push(element);
                    }
                }
                objArray.sort((a, b) => {
                    return a.username.localeCompare(b.username);
                });
                setUsers(objArray);
            });
        })
        return () => {
            sub?.unsubscribe();
        };
    }, []);

    const [searchTerm, setSearchTerm] = React.useState<string>('');
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
            </List>
        </Paper>
        <Grid container spacing={3} sx={{ marginTop: "0px" }}>
            {teams ? teams.filter((t) => {
                if (searchTerm === "") {
                    return true;
                }
                return t.name.toLowerCase().includes(searchTerm.toLowerCase());
            }).map((k) => {
                return <Grid key={k.name} item xs={12} md={4} lg={3}>
                    <Card sx={{ maxWidth: 345 }}>
                        <CardHeader
                            avatar={
                                <Avatar sx={{ bgcolor: teal[500] }} aria-label="recipe">
                                    {k.name.toUpperCase().charAt(0)}
                                </Avatar>
                            }
                            title={k.name}
                            subheader={k.created ? new Date(k.created).toLocaleString() : ""}
                        />
                        <CardContent>
                            <Typography variant="body2" color="text.primary">
                                Members:
                            </Typography>
                            <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }} disablePadding={true} dense={true}>
                                {users ? users.filter((u) => {
                                    if (k.members === undefined) {
                                        return false;
                                    }
                                    return k.members.filter((m) => {
                                        return m === u.id;
                                    }).length > 0;
                                }).map((u) => {
                                    return <ListItem key={u.id} alignItems="flex-start" dense={true} disablePadding={true}>
                                        <ListItemText
                                            primary={u.email ? u.email : u.username}
                                            sx={{
                                                fontWeight: "bold"
                                            }}
                                        />
                                    </ListItem>
                                }) : undefined}
                            </List>
                        </CardContent>
                        {auth.user && userIsTeamMember(k, auth.user.id) ? <CardActions disableSpacing >
                            <span style={{ flexGrow: "1" }}></span>
                            <Tooltip title="Edit members">
                                <IconButton aria-label="members" color='primary' onClick={() => {
                                    if (!k.id) {
                                        return;
                                    }
                                    handleClickOpenEditMembers({
                                        id: k.id,
                                        members: k.members
                                    });
                                }}>
                                    <GroupAddIcon />
                                </IconButton>
                            </Tooltip>
                            <Tooltip title="Delete">
                                <IconButton aria-label="delete" color='primary' onClick={() => {
                                    if (window.confirm(`Do you really want to delete ${k.name}?`) === true) {
                                        if (!k.id) {
                                            return;
                                        }
                                        TeamService.deleteTeam(k.id)
                                            .then(() => {
                                                NotificationService.notifySuccess(`Removed team ${k.name}`);
                                            })
                                            .catch((err) => {
                                                console.log(err);
                                            });
                                    }
                                }}>
                                    <DeleteIcon />
                                </IconButton>
                            </Tooltip>
                        </CardActions> : undefined}
                    </Card>
                </Grid>
            }) : undefined}
            {teams && teams.length === 0 ? <Container sx={{ textAlign: "center" }}>
                <Typography>
                    No teams configured
                </Typography>
            </Container> : undefined}
            {teams === undefined ? <React.Fragment>
                <Grid item xs={12} md={4} lg={3}>
                    <Skeleton variant="rectangular" height={150} />
                </Grid>
                <Grid item xs={12} md={4} lg={3}>
                    <Skeleton variant="rectangular" height={150} />
                </Grid>
                <Grid item xs={12} md={4} lg={3}>
                    <Skeleton variant="rectangular" height={150} />
                </Grid>
                <Grid item xs={12} md={4} lg={3}>
                    <Skeleton variant="rectangular" height={150} />
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
            <DialogTitle>Add new team</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Fill in the form to create a new team.
                </DialogContentText>
                <FormInputText
                    name="name"
                    control={control}
                    required={true}
                    autoFocus={true}
                    label="Name" />
                <FormInputMultiCheckbox
                    name="members"
                    control={control}
                    required={false}
                    label="Assigned members"
                    setValue={setValue}
                    options={users ? users.map((t) => {
                        return {
                            label: t.email ? t.email : t.username,
                            value: t.id as string
                        }
                    }) : []} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleClose() }}>Cancel</Button>
                <Button onClick={handleSubmit(onSubmit)}>Save</Button>
            </DialogActions>
        </Dialog>
        <Dialog open={openEditMembers !== undefined} onClose={handleCloseEditMembers}>
            <DialogTitle>Manage Members</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Select members for this team.
                </DialogContentText>
                <FormInputMultiCheckbox
                    name="members"
                    control={controlEditMembers}
                    required={false}
                    label="Assigned members"
                    setValue={setValueEditMembers}
                    options={users ? users.map((t) => {
                        const len = openEditMembers?.members?.filter((inner) => { return inner === t.id; })
                        return {
                            label: t.email ? t.email : t.username,
                            value: t.id as string,
                            selected: len ? len.length > 0 : false
                        }
                    }) : []} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleCloseEditMembers() }}>Cancel</Button>
                <Button onClick={handleSubmitEditMembers(onSubmitEditMembers)}>Save</Button>
            </DialogActions>
        </Dialog>
    </div >;
}
