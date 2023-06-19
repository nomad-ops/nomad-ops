import * as React from 'react';
import Grid from '@mui/material/Grid';

import RealTimeAccess from '../services/RealTimeAccess';
import { VaultToken } from '../domain/VaultToken';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Avatar from '@mui/material/Avatar';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import { Button, CardActions, Chip, Container, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Fab, List, ListItem, Paper, Skeleton, Stack, TextField, Tooltip } from '@mui/material';
import { teal } from '@mui/material/colors';
import { Subscription } from 'rxjs';
import { FormInputText } from '../components/form-components/FormInputText';
import { useForm } from 'react-hook-form';
import VaultTokenService from '../services/VaultTokenService';
import NotificationService from '../services/NotificationService';
import { FormTextArea } from '../components/form-components/FormTextArea';
import { Team } from '../domain/Team';
import { FormInputDropdown } from '../components/form-components/FormInputDropdown';

interface IVaultTokenFormInput {
    name: string;
    value: string;
    team: string;
}

const defaultVaultTokenValues = {
    name: "",
    value: "",
    team: ""
};

export default function VaultTokens() {
    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = (_ev?: any | undefined, reason?: string | undefined) => {
        if (reason && reason === "backdropClick")
            return;
        setOpen(false);
    };

    const methods = useForm<IVaultTokenFormInput>({ defaultValues: defaultVaultTokenValues });
    const { handleSubmit, reset, control } = methods;
    const onSubmit = (data: IVaultTokenFormInput) => {
        // TODO validate

        VaultTokenService.createVaultToken({
            name: data.name,
            value: data.value,
            team: data.team
        })
            .then(() => {
                NotificationService.notifySuccess(`Created VaultToken ${data.name}...`);
                setOpen(false);
                reset();
            });
    };

    const [vaultTokens, setVaultTokens] = React.useState<VaultToken[] | undefined>(undefined);

    React.useEffect(() => {
        var sub: Subscription | undefined = undefined;
        RealTimeAccess.GetStore<VaultToken>("vault_tokens").then((s) => {
            sub = s.subscribe((VaultTokens) => {
                if (VaultTokens === undefined) {
                    setVaultTokens(undefined);
                    return;
                }
                var objArray: VaultToken[] = [];
                for (const VaultToken in VaultTokens) {
                    if (Object.prototype.hasOwnProperty.call(VaultTokens, VaultToken)) {
                        const element = VaultTokens[VaultToken];
                        objArray.push(element);
                    }
                }
                objArray.sort((a, b) => {
                    return a.name.localeCompare(b.name);
                });
                setVaultTokens(objArray);
            });
        })
        return () => {
            sub?.unsubscribe();
        };
    }, []);

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
            {vaultTokens ? vaultTokens.filter((k) => {
                if (searchTerm === "") {
                    return true;
                }
                return k.name.toLowerCase().includes(searchTerm.toLowerCase());
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
                                * * * *
                            </Typography>
                        </CardContent>
                        <CardActions disableSpacing >
                            <List component={Stack} direction="row">
                                {k.team && teams ? teams.filter((team) => {
                                    return team.id === k.team;
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
                            <Tooltip title="Delete">
                                <IconButton aria-label="delete" color='primary' onClick={() => {
                                    if (window.confirm(`Do you really want to delete ${k.name}?`) === true) {
                                        if (!k.id) {
                                            return;
                                        }
                                        VaultTokenService.deleteVaultToken(k.id)
                                            .then(() => {
                                                NotificationService.notifySuccess(`Removed VaultToken ${k.name}`);
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
            {vaultTokens && vaultTokens.length === 0 ? <Container sx={{ textAlign: "center" }}>
                <Typography>
                    No Vault Tokens configured
                </Typography>
            </Container> : undefined}
            {vaultTokens === undefined ? <React.Fragment>
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
        <Dialog open={open} onClose={handleClose} maxWidth={false} fullWidth>
            <DialogTitle>Add new Vault Token</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Fill in the form to add a new Vault Token.
                </DialogContentText>
                <FormInputText
                    name="name"
                    control={control}
                    required={true}
                    autoFocus={true}
                    label="Name" />
                <FormTextArea
                    name="value"
                    control={control}
                    required={true} label={'VaultToken value'} />
                <FormInputDropdown
                    name="team"
                    control={control}
                    required={false}
                    label="Team"
                    options={teams ? teams.filter((t) => {
                        return t.id !== undefined
                    }).map((t) => {
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
    </div>;
}