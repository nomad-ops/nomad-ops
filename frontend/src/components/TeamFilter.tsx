import * as React from 'react';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import ListItemText from '@mui/material/ListItemText';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Checkbox from '@mui/material/Checkbox';
import { Team } from '../domain/Team';

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
    PaperProps: {
        style: {
            maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
            width: 250,
        },
    },
};

export default function TeamFilter({ teams, userID, onChange }: { teams: Team[] | undefined, userID: string | undefined, onChange: (selectedTeams: Team[]) => void }) {
    const [selectedTeams, setSelectedTeams] = React.useState<Team[]>([]);

    React.useEffect(() => {
        if (teams === undefined) {
            return;
        }
        if (userID === undefined) {
            return;
        }
        const res = teams.filter((t) => {
            if (t.members === undefined) {
                return false;
            }
            return t.members.indexOf(userID) > -1;
        });
        setSelectedTeams(res);
        onChange(res);
    }, [teams, userID]);

    const handleChange = (event: SelectChangeEvent<string[]>) => {
        const {
            target: { value },
        } = event;
        // On autofill we get a stringified value. ? What to do here ?
        if (typeof value === 'string') {
            const res = value.split(",").map((id) => {
                const res = teams?.find((inner) => {
                    return inner.id === id;
                });
                return res as Team;
            });
            setSelectedTeams(res);
            onChange(res);
            return;
        }
        const res = value.map((id) => {
            const res = teams?.find((inner) => {
                return inner.id === id;
            });
            return res as Team;
        });
        setSelectedTeams(
            res
        );
        onChange(res);
    };

    return (
        <div>
            <FormControl sx={{ m: 1, width: 300 }} size="small">
                <InputLabel id="team-filter-multiple-checkbox-label">Filter by teams</InputLabel>
                <Select
                    labelId="team-filter-multiple-checkbox-label"
                    id="team-filter-multiple-checkbox"
                    multiple
                    value={selectedTeams.map((t) => { return t.id as string })}
                    onChange={handleChange}
                    input={<OutlinedInput label="Filter by teams" />}
                    renderValue={(selected) => selected.map((t) => {
                        const res = teams?.find((inner) => {
                            return inner.id === t;
                        })
                        if (res === undefined) {
                            return "Team not found";
                        }
                        return res.name;
                    }).join(', ')}
                    MenuProps={MenuProps}
                >
                    {teams?.map((t) => (
                        <MenuItem key={t.id} value={t.id}>
                            <Checkbox checked={selectedTeams.filter((s) => {
                                return s.id == t.id;
                            }).length > 0} />
                            <ListItemText primary={t.name} />
                        </MenuItem>
                    ))}
                </Select>
            </FormControl>
        </div>
    );
}