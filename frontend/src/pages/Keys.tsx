import * as React from 'react';
import Grid from '@mui/material/Grid';

import RealTimeAccess from '../services/RealTimeAccess';
import { Key } from '../domain/Key';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import Avatar from '@mui/material/Avatar';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import { Button, CardActions, Container, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Fab, List, Paper, Skeleton, Stack, TextField, Tooltip } from '@mui/material';
import { teal } from '@mui/material/colors';
import { Subscription } from 'rxjs';
import { FormInputText } from '../components/form-components/FormInputText';
import { useForm } from 'react-hook-form';
import KeyService from '../services/KeyService';
import NotificationService from '../services/NotificationService';
import { FormTextArea } from '../components/form-components/FormTextArea';

interface IKeyFormInput {
    name: string;
    value: string;
}

const defaultKeyValues = {
    name: "",
    value: ""
};

export default function Keys() {
    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = (_ev?: any | undefined, reason?: string | undefined) => {
        if (reason && reason === "backdropClick")
            return;
        setOpen(false);
    };

    const methods = useForm<IKeyFormInput>({ defaultValues: defaultKeyValues });
    const { handleSubmit, reset, control } = methods;
    const onSubmit = (data: IKeyFormInput) => {
        // TODO validate

        KeyService.createKey({
            name: data.name,
            value: data.value
        })
            .then(() => {
                NotificationService.notifySuccess(`Created key ${data.name}...`);
                setOpen(false);
                reset();
            });
    };

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
                setKeys(objArray);
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
            {keys ? keys.filter((k) => {
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
                            <span style={{ flexGrow: "1" }}></span>
                            <Tooltip title="Delete">
                                <IconButton aria-label="delete" color='primary' onClick={() => {
                                    if (window.confirm(`Do you really want to delete ${k.name}?`) === true) {
                                        if (!k.id) {
                                            return;
                                        }
                                        KeyService.deleteKey(k.id)
                                            .then(() => {
                                                NotificationService.notifySuccess(`Removed key ${k.name}`);
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
            {keys && keys.length === 0 ? <Container sx={{ textAlign: "center" }}>
                <Typography>
                    No keys configured
                </Typography>
            </Container> : undefined}
            {keys === undefined ? <React.Fragment>
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
            <DialogTitle>Add new key</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Fill in the form to add a new Key.
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
                    required={true} label={'Key value'} />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => { handleClose() }}>Cancel</Button>
                <Button onClick={handleSubmit(onSubmit)}>Save</Button>
            </DialogActions>
        </Dialog>
    </div>;
}