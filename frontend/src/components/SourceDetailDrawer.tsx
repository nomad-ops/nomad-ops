import React from "react";
import { Box, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Divider, Drawer, Toolbar, ListSubheader, Typography, ListItemAvatar, IconButton } from "@mui/material";
import MailIcon from '@mui/icons-material/Mail';
import InboxIcon from '@mui/icons-material/Inbox';
import CheckIcon from '@mui/icons-material/Check';
import CancelIcon from '@mui/icons-material/Cancel';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import ErrorIcon from '@mui/icons-material/Error';
import LoopIcon from '@mui/icons-material/Loop';
import AddIcon from '@mui/icons-material/Add';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import SyncIcon from '@mui/icons-material/Sync';
import NotStartedIcon from '@mui/icons-material/NotStarted';
import HourglassBottomIcon from '@mui/icons-material/HourglassBottom';
import QuestionMarkIcon from '@mui/icons-material/QuestionMark';
import { Source } from "../domain/Source";
import { JobInfo } from "../domain/JobInfo";
import NomadService from "../services/NomadService";
import { NomadURLs } from "../domain/NomadURLs";

export default function SourceDetailDrawer({ open, onClose, source }: {
    open: boolean,
    onClose: (event: React.KeyboardEvent | React.MouseEvent) => void,
    source: Source
}) {

    const [nomadURLs, setNomadURLs] = React.useState<NomadURLs | undefined>(undefined);

    React.useEffect(() => {
        NomadService.getNomadURLs()
            .then((urls) => {
                setNomadURLs(urls);
            })
    }, []);

    const [jobInfos, setJobInfos] = React.useState<JobInfo[] | undefined>(undefined);

    React.useEffect(() => {
        if (source.status === undefined) {
            return;
        }
        if (source.status?.jobs === undefined) {
            return;
        }
        const keys = Object.keys(source.status.jobs);
        const promiseArray: Promise<JobInfo>[] = [];
        for (let i = 0; i < keys.length; i++) {
            const element = keys[i];
            promiseArray.push(Promise.all([NomadService.getJobSummary(element, source.namespace as string),
            NomadService.listAllocations(element, source.namespace as string)])
                .then((results) => {
                    const taskGroups = Object.keys(results[0].Summary);
                    return {
                        name: element,
                        namespace: results[0].Namespace,
                        taskGroups: taskGroups.map((t) => {
                            return {
                                name: t,
                                status: "",
                                allocations: results[1].filter((a: any) => {
                                    return a.TaskGroup === t;
                                })
                                    .map((a: any) => {
                                        const tasks = Object.keys(a.TaskStates);
                                        return {
                                            id: a.ID,
                                            tasks: tasks.map((task) => {
                                                return {
                                                    name: task,
                                                    status: a.TaskStates[task].State,
                                                    startedAt: new Date(a.TaskStates[task].StartedAt).toLocaleString(),
                                                    finishedAt: new Date(a.TaskStates[task].FinishedAt).toLocaleString(),
                                                    lastRestart: new Date(a.TaskStates[task].LastRestart).toLocaleString()
                                                }
                                            })
                                        }
                                    })
                            };
                        })
                    }
                }))
        }
        Promise.all(promiseArray)
            .then((results) => {
                setJobInfos(results);
            });
    }, [source]);

    const list = () => (
        <Box
            sx={{ width: 450 }}
            role="presentation"
        >
            <List subheader={
                <ListSubheader component="div">
                    Jobs
                </ListSubheader>
            }>
                {jobInfos ? jobInfos.map((jobInfo) => {
                    return <React.Fragment key={jobInfo.name} >
                        <ListItem secondaryAction={
                            nomadURLs ? <a href={nomadURLs.ui + "/ui/jobs/" + jobInfo.name + "@" + jobInfo.namespace} target="_blank">
                                <IconButton edge="end" aria-label="delete">
                                    <OpenInNewIcon />
                                </IconButton>
                            </a> : undefined
                        }>
                            {jobInfo.name}
                        </ListItem>
                        {jobInfo.taskGroups.map((taskGroupInfo) => {
                            return <React.Fragment key={'taskgroup' + taskGroupInfo.name} >
                                <List sx={{ paddingLeft: "10px" }} subheader={
                                    <ListSubheader component="div" sx={{ lineHeight: "normal" }}>
                                        TaskGroups
                                    </ListSubheader>
                                }>
                                    <ListItem>
                                        {taskGroupInfo.name}
                                    </ListItem>
                                    <List sx={{ paddingLeft: "14px" }} subheader={
                                        <ListSubheader component="div" sx={{ lineHeight: "normal" }}>
                                            Allocations
                                        </ListSubheader>
                                    }>
                                        {taskGroupInfo.allocations.map((allocationInfo) => {
                                            return <React.Fragment key={'allocation' + allocationInfo.id}>
                                                <ListItem>
                                                    {allocationInfo.id}
                                                </ListItem>
                                                <List sx={{ paddingLeft: "14px" }} subheader={
                                                    <ListSubheader component="div" sx={{ lineHeight: "normal" }}>
                                                        Tasks
                                                    </ListSubheader>
                                                }>
                                                    {allocationInfo.tasks.map((taskInfo) => {
                                                        let statusIcon = <QuestionMarkIcon />
                                                        let time = taskInfo.startedAt;
                                                        switch (taskInfo.status) {
                                                            case "running":
                                                                statusIcon = <CheckIcon />
                                                                break;
                                                            case "dead":
                                                                statusIcon = <CancelIcon />
                                                                time = taskInfo.finishedAt
                                                                break;

                                                            default:
                                                                break;
                                                        }
                                                        return <ListItem key={'tasks' + taskInfo.name}>
                                                            <ListItemAvatar>
                                                                {statusIcon}
                                                            </ListItemAvatar>
                                                            <ListItemText
                                                                primary={taskInfo.name}
                                                                secondary={
                                                                    <React.Fragment>
                                                                        <Typography
                                                                            sx={{ display: 'inline' }}
                                                                            component="span"
                                                                            variant="body2"
                                                                            color="text.primary"
                                                                        >
                                                                            {taskInfo.status}
                                                                        </Typography>
                                                                        {" — " + time}
                                                                    </React.Fragment>
                                                                }
                                                            />
                                                        </ListItem>
                                                    })}
                                                </List>
                                            </React.Fragment>
                                        })}
                                    </List>
                                </List>
                            </React.Fragment>
                        })}
                        <Divider />
                    </React.Fragment>
                }) : undefined}
            </List>
        </Box >
    );
    return <Drawer
        anchor={'right'}
        open={open}
        onClose={onClose}
    >
        <Toolbar />
        {list()}
    </Drawer>
}